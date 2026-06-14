package mq

import (
	"context"
	"log"

	"github.com/cinema-booking/backend/internal/admin"
	"github.com/cinema-booking/backend/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Consumer struct {
	ch    *amqp.Channel
	audit *admin.AuditService
}

func NewConsumer(ch *amqp.Channel, audit *admin.AuditService) *Consumer {
	return &Consumer{ch: ch, audit: audit}
}

func (c *Consumer) Start(ctx context.Context) {
	queues := []struct {
		name  string
		event string
		mock  bool
	}{
		{QueueBookingSuccess, model.AuditBookingSuccess, true},
		{QueueBookingTimeout, model.AuditBookingTimeout, false},
		{QueueSeatReleased, model.AuditSeatReleased, false},
	}

	for _, q := range queues {
		queue := q
		go c.consume(ctx, queue.name, queue.event, queue.mock)
	}
}

func (c *Consumer) consume(ctx context.Context, queueName, auditEvent string, mockNotify bool) {
	deliveries, err := c.ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("mq consumer %s: %v", queueName, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			c.handleDelivery(ctx, delivery, auditEvent, mockNotify)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery, auditEvent string, mockNotify bool) {
	event, err := BookingEventFromBytes(delivery.Body)
	if err != nil {
		log.Printf("mq consumer decode: %v", err)
		return
	}

	userID, _ := primitive.ObjectIDFromHex(event.UserID)
	showtimeID, _ := primitive.ObjectIDFromHex(event.ShowtimeID)
	uid := userID
	stid := showtimeID

	seatNo := ""
	if len(event.SeatNos) > 0 {
		seatNo = event.SeatNos[0]
	}

	detail := "booking_id=" + event.BookingID
	if len(event.SeatNos) > 1 {
		detail += " seats=" + joinSeatNos(event.SeatNos)
	}

	if mockNotify && auditEvent == model.AuditBookingSuccess {
		log.Printf("mock notification: booking %s confirmed for seats %v", event.BookingID, event.SeatNos)
	}

	if err := c.audit.Log(ctx, auditEvent, &uid, &stid, seatNo, detail); err != nil {
		log.Printf("mq consumer audit %s: %v", auditEvent, err)
	}
}

func joinSeatNos(seats []string) string {
	result := ""
	for i, seat := range seats {
		if i > 0 {
			result += ","
		}
		result += seat
	}
	return result
}
