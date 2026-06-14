package admin

import (
	"context"
	"errors"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrFilterRequired = errors.New("at least one filter is required")

type BookingListFilter struct {
	MovieID *primitive.ObjectID
	UserID  *primitive.ObjectID
	Date    *time.Time
}

type AuditListFilter struct {
	Event string
	From  *time.Time
	To    *time.Time
}

type AdminBookingItem struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	UserEmail  string     `json:"user_email"`
	UserName   string     `json:"user_name"`
	MovieTitle string     `json:"movie_title"`
	ShowtimeID string     `json:"showtime_id"`
	Screen     string     `json:"screen"`
	StartTime  time.Time  `json:"start_time"`
	SeatNos    []string   `json:"seat_nos"`
	Status     string     `json:"status"`
	Amount     float64    `json:"amount"`
	CreatedAt  time.Time  `json:"created_at"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
}

type DashboardService struct {
	bookings  *repository.BookingRepository
	showtimes *repository.ShowtimeRepository
	movies    *repository.MovieRepository
	users     *repository.UserRepository
	audits    *repository.AuditRepository
}

func NewDashboardService(
	bookings *repository.BookingRepository,
	showtimes *repository.ShowtimeRepository,
	movies *repository.MovieRepository,
	users *repository.UserRepository,
	audits *repository.AuditRepository,
) *DashboardService {
	return &DashboardService{
		bookings:  bookings,
		showtimes: showtimes,
		movies:    movies,
		users:     users,
		audits:    audits,
	}
}

func (s *DashboardService) ListBookings(ctx context.Context, filter BookingListFilter) ([]AdminBookingItem, error) {
	if filter.MovieID == nil && filter.UserID == nil && filter.Date == nil {
		return nil, ErrFilterRequired
	}

	showtimeFilterNeeded := filter.MovieID != nil || filter.Date != nil
	var showtimeIDs []primitive.ObjectID

	if showtimeFilterNeeded {
		ids, err := s.resolveShowtimeIDs(ctx, filter)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []AdminBookingItem{}, nil
		}
		showtimeIDs = ids
	}

	query := bson.M{}
	if filter.UserID != nil {
		query["user_id"] = *filter.UserID
	}
	if len(showtimeIDs) > 0 {
		query["showtime_id"] = bson.M{"$in": showtimeIDs}
	}

	bookings, err := s.bookings.FindAdmin(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.mapBookingItems(ctx, bookings)
}

func (s *DashboardService) ListAuditLogs(ctx context.Context, filter AuditListFilter) ([]model.AuditLog, error) {
	return s.audits.Find(ctx, repository.AuditLogFilter{
		Event: filter.Event,
		From:  filter.From,
		To:    filter.To,
	})
}

func (s *DashboardService) resolveShowtimeIDs(ctx context.Context, filter BookingListFilter) ([]primitive.ObjectID, error) {
	var candidates [][]primitive.ObjectID

	if filter.MovieID != nil {
		showtimes, err := s.showtimes.FindByMovieID(ctx, *filter.MovieID)
		if err != nil {
			return nil, err
		}
		ids := make([]primitive.ObjectID, 0, len(showtimes))
		for _, st := range showtimes {
			ids = append(ids, st.ID)
		}
		candidates = append(candidates, ids)
	}

	if filter.Date != nil {
		showtimes, err := s.showtimes.FindByStartDate(ctx, *filter.Date)
		if err != nil {
			return nil, err
		}
		ids := make([]primitive.ObjectID, 0, len(showtimes))
		for _, st := range showtimes {
			ids = append(ids, st.ID)
		}
		candidates = append(candidates, ids)
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	lookup := make(map[primitive.ObjectID]int)
	for _, id := range candidates[0] {
		lookup[id]++
	}
	for _, id := range candidates[1] {
		if lookup[id] > 0 {
			lookup[id]++
		}
	}

	result := make([]primitive.ObjectID, 0)
	for id, count := range lookup {
		if count == len(candidates) {
			result = append(result, id)
		}
	}
	return result, nil
}

func (s *DashboardService) mapBookingItems(ctx context.Context, bookings []model.Booking) ([]AdminBookingItem, error) {
	showtimeCache := make(map[primitive.ObjectID]*model.Showtime)
	movieCache := make(map[primitive.ObjectID]string)
	userCache := make(map[primitive.ObjectID]*model.User)

	items := make([]AdminBookingItem, 0, len(bookings))
	for _, booking := range bookings {
		showtime, ok := showtimeCache[booking.ShowtimeID]
		if !ok {
			st, err := s.showtimes.FindByID(ctx, booking.ShowtimeID)
			if err != nil {
				continue
			}
			showtime = st
			showtimeCache[booking.ShowtimeID] = st
		}

		movieTitle, ok := movieCache[showtime.MovieID]
		if !ok {
			movie, err := s.movies.FindByID(ctx, showtime.MovieID)
			if err != nil {
				movieTitle = ""
			} else {
				movieTitle = movie.Title
			}
			movieCache[showtime.MovieID] = movieTitle
		}

		user, ok := userCache[booking.UserID]
		if !ok {
			u, err := s.users.FindByID(ctx, booking.UserID)
			if err == nil {
				user = u
			}
			userCache[booking.UserID] = user
		}

		item := AdminBookingItem{
			ID:         booking.ID.Hex(),
			UserID:     booking.UserID.Hex(),
			MovieTitle: movieTitle,
			ShowtimeID: booking.ShowtimeID.Hex(),
			Screen:     showtime.Screen,
			StartTime:  showtime.StartTime,
			SeatNos:    booking.SeatNos,
			Status:     booking.Status,
			Amount:     booking.Amount,
			CreatedAt:  booking.CreatedAt,
			PaidAt:     booking.PaidAt,
		}
		if user != nil {
			item.UserEmail = user.Email
			item.UserName = user.Name
		}
		items = append(items, item)
	}

	return items, nil
}
