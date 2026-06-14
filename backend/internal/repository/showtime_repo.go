package repository

import (
	"context"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShowtimeRepository struct {
	col *mongo.Collection
}

func NewShowtimeRepository(db *mongo.Database) *ShowtimeRepository {
	return &ShowtimeRepository{col: db.Collection("showtimes")}
}

func (r *ShowtimeRepository) Insert(ctx context.Context, showtime *model.Showtime) error {
	res, err := r.col.InsertOne(ctx, showtime)
	if err != nil {
		return err
	}
	showtime.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ShowtimeRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Showtime, error) {
	var showtime model.Showtime
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&showtime)
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func (r *ShowtimeRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}

func (r *ShowtimeRepository) FindAll(ctx context.Context) ([]model.Showtime, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var showtimes []model.Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, err
	}
	return showtimes, nil
}

func (r *ShowtimeRepository) FindByMovieID(ctx context.Context, movieID primitive.ObjectID) ([]model.Showtime, error) {
	cursor, err := r.col.Find(ctx, bson.M{"movie_id": movieID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var showtimes []model.Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, err
	}
	return showtimes, nil
}

func (r *ShowtimeRepository) FindByStartDate(ctx context.Context, day time.Time) ([]model.Showtime, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	cursor, err := r.col.Find(ctx, bson.M{
		"start_time": bson.M{"$gte": start, "$lt": end},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var showtimes []model.Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, err
	}
	return showtimes, nil
}
