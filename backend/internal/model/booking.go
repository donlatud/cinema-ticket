package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BookingPending = "PENDING"
	BookingPaid    = "PAID"
	BookingExpired = "EXPIRED"
)

type Booking struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	SeatNos    []string           `bson:"seat_nos" json:"seat_nos"`
	LockTokens map[string]string  `bson:"lock_tokens,omitempty" json:"lock_tokens,omitempty"`
	Status     string             `bson:"status" json:"status"`
	Amount     float64            `bson:"amount" json:"amount"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	PaidAt     *time.Time         `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
	ExpiresAt  *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}
