package repository

import (
	"context"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type AuditLogFilter struct {
	Event string
	From  *time.Time
	To    *time.Time
}

func (r *AuditRepository) Find(ctx context.Context, filter AuditLogFilter) ([]model.AuditLog, error) {
	query := bson.M{}
	if filter.Event != "" {
		query["event"] = filter.Event
	}
	if filter.From != nil || filter.To != nil {
		createdAt := bson.M{}
		if filter.From != nil {
			createdAt["$gte"] = *filter.From
		}
		if filter.To != nil {
			createdAt["$lte"] = *filter.To
		}
		query["created_at"] = createdAt
	}

	cursor, err := r.col.Find(ctx, query, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
