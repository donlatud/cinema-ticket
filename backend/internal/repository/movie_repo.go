package repository

import (
	"context"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MovieRepository struct {
	col *mongo.Collection
}

func NewMovieRepository(db *mongo.Database) *MovieRepository {
	return &MovieRepository{col: db.Collection("movies")}
}

func (r *MovieRepository) Insert(ctx context.Context, movie *model.Movie) error {
	res, err := r.col.InsertOne(ctx, movie)
	if err != nil {
		return err
	}
	movie.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MovieRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Movie, error) {
	var movie model.Movie
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
