package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cinema-booking/backend/internal/lock"
	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/realtime"
	"github.com/cinema-booking/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const bookingTTL = 5 * time.Minute

var (
	ErrShowtimeNotFound  = errors.New("showtime not found")
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatUnavailable   = errors.New("seat unavailable")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingForbidden  = errors.New("booking forbidden")
	ErrBookingNotPending = errors.New("booking not pending")
	ErrBookingExpired    = errors.New("booking expired")
)

type Service struct {
	showtimes *repository.ShowtimeRepository
	seats     *repository.SeatRepository
	bookings  *repository.BookingRepository
	movies    *repository.MovieRepository
	lock      *lock.RedisLock
	hub       *realtime.Hub
}

func NewService(
	showtimes *repository.ShowtimeRepository,
	seats *repository.SeatRepository,
	bookings *repository.BookingRepository,
	movies *repository.MovieRepository,
	seatLock *lock.RedisLock,
	hub *realtime.Hub,
) *Service {
	return &Service{
		showtimes: showtimes,
		seats:     seats,
		bookings:  bookings,
		movies:    movies,
		lock:      seatLock,
		hub:       hub,
	}
}

type ShowtimeListItem struct {
	ID         string    `json:"id"`
	MovieID    string    `json:"movie_id"`
	MovieTitle string    `json:"movie_title"`
	StartTime  time.Time `json:"start_time"`
	Screen     string    `json:"screen"`
	Price      float64   `json:"price"`
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
	for _, seat := range seats {
		if seat.Status != model.SeatAvailable {
			return nil, fmt.Errorf("%w: %s", ErrSeatUnavailable, seat.SeatNo)
		}
	}

	showtimeIDHex := showtimeID.Hex()
	userIDHex := userID.Hex()
	lockTokens := make(map[string]string, len(seatNos))
	acquiredRedis := make([]string, 0, len(seatNos))

	for _, seatNo := range seatNos {
		token, err := s.lock.AcquireLock(ctx, showtimeIDHex, seatNo, userIDHex)
		if err != nil {
			s.releaseRedisLocks(ctx, showtimeIDHex, acquiredRedis, lockTokens)
			if errors.Is(err, lock.ErrLockNotAcquired) {
				return nil, fmt.Errorf("%w: %s", ErrSeatUnavailable, seatNo)
			}
			return nil, err
		}
		lockTokens[seatNo] = token
		acquiredRedis = append(acquiredRedis, seatNo)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(bookingTTL)
	booking := &model.Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		SeatNos:    seatNos,
		LockTokens: lockTokens,
		Status:     model.BookingPending,
		Amount:     showtime.Price * float64(len(seatNos)),
		CreatedAt:  now,
		ExpiresAt:  &expiresAt,
	}
	if err := s.bookings.Insert(ctx, booking); err != nil {
		s.releaseRedisLocks(ctx, showtimeIDHex, acquiredRedis, lockTokens)
		return nil, err
	}

	lockedMongo := make([]string, 0, len(seatNos))
	for _, seatNo := range seatNos {
		ok, err := s.seats.LockSeat(ctx, showtimeID, seatNo, userID, booking.ID, now, lockTokens[seatNo])
		if err != nil {
			s.rollbackLock(ctx, showtimeID, lockedMongo, booking, lockTokens)
			return nil, err
		}
		if !ok {
			s.rollbackLock(ctx, showtimeID, lockedMongo, booking, lockTokens)
			return nil, fmt.Errorf("%w: %s", ErrSeatUnavailable, seatNo)
		}
		lockedMongo = append(lockedMongo, seatNo)
	}

	s.broadcastSeats(showtimeID, lockedMongo, model.SeatLocked)

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

	s.releaseRedisLocks(ctx, booking.ShowtimeID.Hex(), booking.SeatNos, booking.LockTokens)

	if err := s.bookings.UpdateStatus(ctx, booking.ID, model.BookingPaid, &now); err != nil {
		return nil, err
	}

	s.broadcastSeats(booking.ShowtimeID, booking.SeatNos, model.SeatBooked)

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

	s.broadcastSeats(booking.ShowtimeID, booking.SeatNos, model.SeatAvailable)

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
		s.broadcastSeats(booking.ShowtimeID, booking.SeatNos, model.SeatAvailable)
		log.Printf("BOOKING_TIMEOUT booking_id=%s showtime_id=%s seats=%v",
			booking.ID.Hex(), booking.ShowtimeID.Hex(), booking.SeatNos)
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
	s.releaseRedisLocks(ctx, booking.ShowtimeID.Hex(), booking.SeatNos, booking.LockTokens)
	for _, seatNo := range booking.SeatNos {
		if err := s.seats.UnlockSeat(ctx, booking.ShowtimeID, seatNo, booking.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) releaseRedisLocks(ctx context.Context, showtimeID string, seatNos []string, tokens map[string]string) {
	for _, seatNo := range seatNos {
		token := tokens[seatNo]
		if token == "" {
			continue
		}
		_ = s.lock.ReleaseLock(ctx, showtimeID, seatNo, token)
	}
}

func (s *Service) rollbackLock(ctx context.Context, showtimeID primitive.ObjectID, lockedMongo []string, booking *model.Booking, lockTokens map[string]string) {
	for _, seatNo := range lockedMongo {
		_ = s.seats.UnlockSeat(ctx, showtimeID, seatNo, booking.ID)
	}
	s.releaseRedisLocks(ctx, showtimeID.Hex(), booking.SeatNos, lockTokens)
	_ = s.bookings.UpdateStatus(ctx, booking.ID, model.BookingExpired, nil)
}

func (s *Service) broadcastSeats(showtimeID primitive.ObjectID, seatNos []string, status string) {
	if s.hub == nil {
		return
	}
	showtimeHex := showtimeID.Hex()
	for _, seatNo := range seatNos {
		s.hub.BroadcastSeatUpdate(showtimeHex, seatNo, status)
	}
}
