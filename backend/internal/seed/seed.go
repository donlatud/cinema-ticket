package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cinema-booking/backend/internal/database"
	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run(ctx context.Context, db *mongo.Database) error {
	if err := database.EnsureIndexes(ctx, db); err != nil {
		return err
	}

	movieRepo := repository.NewMovieRepository(db)
	count, err := movieRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("count movies: %w", err)
	}
	if count > 0 {
		log.Println("seed data already exists, skipping")
		return nil
	}

	movie := &model.Movie{
		Title:       "Avengers: Endgame",
		PosterURL:   "https://placehold.co/300x450?text=Avengers",
		DurationMin: 181,
	}
	if err := movieRepo.Insert(ctx, movie); err != nil {
		return fmt.Errorf("insert movie: %w", err)
	}

	showtimeRepo := repository.NewShowtimeRepository(db)
	now := time.Now().UTC()
	showtimes := []model.Showtime{
		{
			MovieID:   movie.ID,
			StartTime: time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, time.UTC).Add(24 * time.Hour),
			Screen:    "Screen 1",
			Price:     12.50,
		},
		{
			MovieID:   movie.ID,
			StartTime: time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			Screen:    "Screen 2",
			Price:     15.00,
		},
	}
	for i := range showtimes {
		if err := showtimeRepo.Insert(ctx, &showtimes[i]); err != nil {
			return fmt.Errorf("insert showtime: %w", err)
		}
	}

	seatRepo := repository.NewSeatRepository(db)
	for _, showtime := range showtimes {
		seats := buildSeats(showtime.ID)
		if err := seatRepo.InsertMany(ctx, seats); err != nil {
			return fmt.Errorf("insert seats for showtime %s: %w", showtime.ID.Hex(), err)
		}
	}

	userRepo := repository.NewUserRepository(db)
	admin := &model.User{
		FirebaseUID: "seed-admin-uid",
		Email:       "admin@cinema.com",
		Name:        "Admin User",
		Role:        model.RoleAdmin,
		CreatedAt:   time.Now().UTC(),
	}
	if err := userRepo.Insert(ctx, admin); err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}

	log.Printf("seed complete: 1 movie, %d showtimes, %d seats, 1 admin user",
		len(showtimes), len(showtimes)*100)
	return nil
}

func buildSeats(showtimeID primitive.ObjectID) []model.Seat {
	rows := []rune("ABCDEFGHIJ")
	seats := make([]model.Seat, 0, 100)

	for _, row := range rows {
		for col := 1; col <= 10; col++ {
			seats = append(seats, model.Seat{
				ShowtimeID: showtimeID,
				SeatNo:     fmt.Sprintf("%c%d", row, col),
				Status:     model.SeatAvailable,
			})
		}
	}

	return seats
}
