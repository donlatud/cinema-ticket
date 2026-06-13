package repository

import (
	"context"
	"time"

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

func (r *SeatRepository) FindByShowtimeAndSeatNos(ctx context.Context, showtimeID primitive.ObjectID, seatNos []string) ([]model.Seat, error) {
	cursor, err := r.col.Find(ctx, bson.M{
		"showtime_id": showtimeID,
		"seat_no":     bson.M{"$in": seatNos},
	})
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

func (r *SeatRepository) LockSeat(ctx context.Context, showtimeID primitive.ObjectID, seatNo string, userID, bookingID primitive.ObjectID, lockedAt time.Time, lockToken string) (bool, error) {
	res, err := r.col.UpdateOne(ctx, bson.M{
		"showtime_id": showtimeID,
		"seat_no":     seatNo,
		"status":      model.SeatAvailable,
	}, bson.M{
		"$set": bson.M{
			"status":     model.SeatLocked,
			"locked_by":  userID,
			"locked_at":  lockedAt,
			"lock_token": lockToken,
			"booking_id": bookingID,
		},
	})
	if err != nil {
		return false, err
	}
	return res.ModifiedCount == 1, nil
}

func (r *SeatRepository) UnlockSeat(ctx context.Context, showtimeID primitive.ObjectID, seatNo string, bookingID primitive.ObjectID) error {
	_, err := r.col.UpdateOne(ctx, bson.M{
		"showtime_id": showtimeID,
		"seat_no":     seatNo,
		"booking_id":  bookingID,
		"status":      model.SeatLocked,
	}, bson.M{
		"$set": bson.M{
			"status": model.SeatAvailable,
		},
		"$unset": bson.M{
			"locked_by":  "",
			"locked_at":  "",
			"lock_token": "",
			"booking_id": "",
		},
	})
	return err
}

func (r *SeatRepository) BookSeat(ctx context.Context, showtimeID primitive.ObjectID, seatNo string, bookingID primitive.ObjectID) error {
	_, err := r.col.UpdateOne(ctx, bson.M{
		"showtime_id": showtimeID,
		"seat_no":     seatNo,
		"booking_id":  bookingID,
		"status":      model.SeatLocked,
	}, bson.M{
		"$set": bson.M{
			"status": model.SeatBooked,
		},
		"$unset": bson.M{
			"locked_by":  "",
			"locked_at":  "",
			"lock_token": "",
		},
	})
	return err
}
