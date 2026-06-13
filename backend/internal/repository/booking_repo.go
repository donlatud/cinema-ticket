package repository

import (
	"context"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository struct {
	col *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{col: db.Collection("bookings")}
}

func (r *BookingRepository) Insert(ctx context.Context, booking *model.Booking) error {
	res, err := r.col.InsertOne(ctx, booking)
	if err != nil {
		return err
	}
	booking.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *BookingRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Booking, error) {
	var booking model.Booking
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
