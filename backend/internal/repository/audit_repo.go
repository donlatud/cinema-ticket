package repository

import (
	"context"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditRepository struct {
	col *mongo.Collection
}

func NewAuditRepository(db *mongo.Database) *AuditRepository {
	return &AuditRepository{col: db.Collection("audit_logs")}
}

func (r *AuditRepository) Insert(ctx context.Context, log *model.AuditLog) error {
	_, err := r.col.InsertOne(ctx, log)
	return err
}

func (r *AuditRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
