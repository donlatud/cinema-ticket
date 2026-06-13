package repository

import (
	"context"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{col: db.Collection("users")}
}

func (r *UserRepository) Insert(ctx context.Context, user *model.User) error {
	res, err := r.col.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var user model.User
	err := r.col.FindOne(ctx, bson.M{"firebase_uid": firebaseUID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpsertFromFirebase(ctx context.Context, firebaseUID, email, name string) (*model.User, error) {
	if existing, err := r.FindByFirebaseUID(ctx, firebaseUID); err == nil {
		return r.updateProfile(ctx, existing.ID, firebaseUID, email, name, existing.Role)
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	if existing, err := r.FindByEmail(ctx, email); err == nil {
		return r.updateProfile(ctx, existing.ID, firebaseUID, email, name, existing.Role)
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	user := &model.User{
		FirebaseUID: firebaseUID,
		Email:       email,
		Name:        name,
		Role:        model.RoleUser,
		CreatedAt:   time.Now().UTC(),
	}
	if err := r.Insert(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) updateProfile(
	ctx context.Context,
	id primitive.ObjectID,
	firebaseUID, email, name, role string,
) (*model.User, error) {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"firebase_uid": firebaseUID,
			"email":        email,
			"name":         name,
		}},
	)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          id,
		FirebaseUID: firebaseUID,
		Email:       email,
		Name:        name,
		Role:        role,
	}, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
