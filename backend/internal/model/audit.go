package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AuditBookingSuccess = "BOOKING_SUCCESS"
	AuditBookingTimeout = "BOOKING_TIMEOUT"
	AuditSeatReleased   = "SEAT_RELEASED"
	AuditSystemError    = "SYSTEM_ERROR"
)

type AuditLog struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Event      string              `bson:"event" json:"event"`
	UserID     *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ShowtimeID *primitive.ObjectID `bson:"showtime_id,omitempty" json:"showtime_id,omitempty"`
	SeatNo     string              `bson:"seat_no,omitempty" json:"seat_no,omitempty"`
	Detail     string              `bson:"detail,omitempty" json:"detail,omitempty"`
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
}
