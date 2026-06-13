package repository

import (
	"context"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SeatRepository struct {
	col *mongo.Collection
}

func NewSeatRepository(db *mongo.Database) *SeatRepository {
	return &SeatRepository{col: db.Collection("seats")}
}

func (r *SeatRepository) InsertMany(ctx context.Context, seats []model.Seat) error {
	if len(seats) == 0 {
		return nil
	}
	docs := make([]interface{}, len(seats))
	for i := range seats {
		docs[i] = seats[i]
	}
	_, err := r.col.InsertMany(ctx, docs)
	return err
}

func (r *SeatRepository) FindByShowtime(ctx context.Context, showtimeID primitive.ObjectID) ([]model.Seat, error) {
	cursor, err := r.col.Find(ctx, bson.M{"showtime_id": showtimeID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var seats []model.Seat
	if err := cursor.All(ctx, &seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *SeatRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
