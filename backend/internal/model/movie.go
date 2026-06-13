package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	PosterURL   string             `bson:"poster_url" json:"poster_url"`
	DurationMin int                `bson:"duration_min" json:"duration_min"`
}
