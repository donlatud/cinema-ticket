package admin

import (
	"context"
	"time"

	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(
	ctx context.Context,
	event string,
	userID *primitive.ObjectID,
	showtimeID *primitive.ObjectID,
	seatNo string,
	detail string,
) error {
	entry := &model.AuditLog{
		Event:     event,
		UserID:    userID,
		ShowtimeID: showtimeID,
		SeatNo:    seatNo,
		Detail:    detail,
		CreatedAt: time.Now().UTC(),
	}
	return s.repo.Insert(ctx, entry)
}
