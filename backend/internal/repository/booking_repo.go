package repository

import (
	"context"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *BookingRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Booking, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, paidAt *time.Time) error {
	update := bson.M{"status": status}
	if paidAt != nil {
		update["paid_at"] = paidAt
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *BookingRepository) FindExpiredPending(ctx context.Context, before time.Time) ([]model.Booking, error) {
	cursor, err := r.col.Find(ctx, bson.M{
		"status":     model.BookingPending,
		"expires_at": bson.M{"$lt": before},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *BookingRepository) FindAdmin(ctx context.Context, filter bson.M) ([]model.Booking, error) {
	cursor, err := r.col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}
