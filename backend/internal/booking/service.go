package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const bookingTTL = 5 * time.Minute

var (
	ErrShowtimeNotFound   = errors.New("showtime not found")
	ErrSeatNotFound       = errors.New("seat not found")
	ErrSeatUnavailable    = errors.New("seat unavailable")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrBookingForbidden   = errors.New("booking forbidden")
	ErrBookingNotPending  = errors.New("booking not pending")
	ErrBookingExpired     = errors.New("booking expired")
)

type Service struct {
	showtimes *repository.ShowtimeRepository
	seats     *repository.SeatRepository
	bookings  *repository.BookingRepository
	movies    *repository.MovieRepository
}

func NewService(
	showtimes *repository.ShowtimeRepository,
	seats *repository.SeatRepository,
	bookings *repository.BookingRepository,
	movies *repository.MovieRepository,
) *Service {
	return &Service{
		showtimes: showtimes,
		seats:     seats,
		bookings:  bookings,
		movies:    movies,
	}
}

type ShowtimeListItem struct {
	ID        string    `json:"id"`
	MovieID   string    `json:"movie_id"`
	MovieTitle string   `json:"movie_title"`
	StartTime time.Time `json:"start_time"`
	Screen    string    `json:"screen"`
	Price     float64   `json:"price"`
}

func (s *Service) ListShowtimes(ctx context.Context) ([]ShowtimeListItem, error) {
	showtimes, err := s.showtimes.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]ShowtimeListItem, 0, len(showtimes))
	for _, st := range showtimes {
		item := ShowtimeListItem{
			ID:        st.ID.Hex(),
			MovieID:   st.MovieID.Hex(),
			StartTime: st.StartTime,
			Screen:    st.Screen,
			Price:     st.Price,
		}
		movie, err := s.movies.FindByID(ctx, st.MovieID)
		if err == nil {
			item.MovieTitle = movie.Title
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) ListSeats(ctx context.Context, showtimeID primitive.ObjectID) ([]model.Seat, error) {
	if _, err := s.showtimes.FindByID(ctx, showtimeID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrShowtimeNotFound
		}
		return nil, err
	}
	return s.seats.FindByShowtime(ctx, showtimeID)
}

func (s *Service) LockSeats(ctx context.Context, showtimeID, userID primitive.ObjectID, seatNos []string) (*model.Booking, error) {
	showtime, err := s.showtimes.FindByID(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrShowtimeNotFound
		}
		return nil, err
	}

	seats, err := s.seats.FindByShowtimeAndSeatNos(ctx, showtimeID, seatNos)
	if err != nil {
		return nil, err
	}
	if len(seats) != len(seatNos) {
		return nil, ErrSeatNotFound
	}

	now := time.Now().UTC()
	expiresAt := now.Add(bookingTTL)
	lockToken := uuid.New().String()

	booking := &model.Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		SeatNos:    seatNos,
		Status:     model.BookingPending,
		Amount:     showtime.Price * float64(len(seatNos)),
		CreatedAt:  now,
		ExpiresAt:  &expiresAt,
	}
	if err := s.bookings.Insert(ctx, booking); err != nil {
		return nil, err
	}

	locked := make([]string, 0, len(seatNos))
	for _, seatNo := range seatNos {
		ok, err := s.seats.LockSeat(ctx, showtimeID, seatNo, userID, booking.ID, now, lockToken)
		if err != nil {
			s.rollbackLocks(ctx, showtimeID, locked, booking.ID)
			return nil, err
		}
		if !ok {
			s.rollbackLocks(ctx, showtimeID, locked, booking.ID)
			return nil, fmt.Errorf("%w: %s", ErrSeatUnavailable, seatNo)
		}
		locked = append(locked, seatNo)
	}

	return booking, nil
}

func (s *Service) Pay(ctx context.Context, bookingID, userID primitive.ObjectID) (*model.Booking, error) {
	booking, err := s.getOwnedPendingBooking(ctx, bookingID, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, seatNo := range booking.SeatNos {
		if err := s.seats.BookSeat(ctx, booking.ShowtimeID, seatNo, booking.ID); err != nil {
			return nil, err
		}
	}

	if err := s.bookings.UpdateStatus(ctx, booking.ID, model.BookingPaid, &now); err != nil {
		return nil, err
	}

	booking.Status = model.BookingPaid
	booking.PaidAt = &now
	return booking, nil
}

func (s *Service) Cancel(ctx context.Context, bookingID, userID primitive.ObjectID) (*model.Booking, error) {
	booking, err := s.getOwnedPendingBooking(ctx, bookingID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.releaseBookingSeats(ctx, booking); err != nil {
		return nil, err
	}
	if err := s.bookings.UpdateStatus(ctx, booking.ID, model.BookingExpired, nil); err != nil {
		return nil, err
	}

	booking.Status = model.BookingExpired
	return booking, nil
}

func (s *Service) ListMyBookings(ctx context.Context, userID primitive.ObjectID) ([]model.Booking, error) {
	return s.bookings.FindByUserID(ctx, userID)
}

func (s *Service) ExpirePendingBookings(ctx context.Context) error {
	bookings, err := s.bookings.FindExpiredPending(ctx, time.Now().UTC())
	if err != nil {
		return err
	}

	for _, booking := range bookings {
		if err := s.releaseBookingSeats(ctx, &booking); err != nil {
			return err
		}
		if err := s.bookings.UpdateStatus(ctx, booking.ID, model.BookingExpired, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getOwnedPendingBooking(ctx context.Context, bookingID, userID primitive.ObjectID) (*model.Booking, error) {
	booking, err := s.bookings.FindByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	if booking.UserID != userID {
		return nil, ErrBookingForbidden
	}
	if booking.Status != model.BookingPending {
		return nil, ErrBookingNotPending
	}
	if booking.ExpiresAt != nil && booking.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrBookingExpired
	}
	return booking, nil
}

func (s *Service) releaseBookingSeats(ctx context.Context, booking *model.Booking) error {
	for _, seatNo := range booking.SeatNos {
		if err := s.seats.UnlockSeat(ctx, booking.ShowtimeID, seatNo, booking.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) rollbackLocks(ctx context.Context, showtimeID primitive.ObjectID, seatNos []string, bookingID primitive.ObjectID) {
	for _, seatNo := range seatNos {
		_ = s.seats.UnlockSeat(ctx, showtimeID, seatNo, bookingID)
	}
	_ = s.bookings.UpdateStatus(ctx, bookingID, model.BookingExpired, nil)
}
