package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Showtime struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	Screen    string             `bson:"screen" json:"screen"`
	Price     float64            `bson:"price" json:"price"`
}
