package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SeatAvailable = "AVAILABLE"
	SeatLocked    = "LOCKED"
	SeatBooked    = "BOOKED"
)

type Seat struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ShowtimeID primitive.ObjectID  `bson:"showtime_id" json:"showtime_id"`
	SeatNo     string              `bson:"seat_no" json:"seat_no"`
	Status     string              `bson:"status" json:"status"`
	LockedBy   *primitive.ObjectID `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
	LockedAt   *time.Time          `bson:"locked_at,omitempty" json:"locked_at,omitempty"`
	LockToken  string              `bson:"lock_token,omitempty" json:"lock_token,omitempty"`
	BookingID  *primitive.ObjectID `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
}
