package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cinema-booking/backend/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	ch *amqp.Channel
}

func NewProducer(ch *amqp.Channel) *Producer {
	return &Producer{ch: ch}
}

func (p *Producer) PublishBookingSuccess(ctx context.Context, booking *model.Booking) error {
	return p.publish(ctx, QueueBookingSuccess, booking)
}

func (p *Producer) PublishBookingTimeout(ctx context.Context, booking *model.Booking) error {
	return p.publish(ctx, QueueBookingTimeout, booking)
}

func (p *Producer) PublishSeatReleased(ctx context.Context, booking *model.Booking) error {
	return p.publish(ctx, QueueSeatReleased, booking)
}

func (p *Producer) publish(ctx context.Context, queue string, booking *model.Booking) error {
	payload, err := json.Marshal(BookingEvent{
		BookingID:  booking.ID.Hex(),
		UserID:     booking.UserID.Hex(),
		ShowtimeID: booking.ShowtimeID.Hex(),
		SeatNos:    booking.SeatNos,
	})
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	})
}

func BookingEventFromBytes(body []byte) (BookingEvent, error) {
	var event BookingEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return event, fmt.Errorf("unmarshal booking event: %w", err)
	}
	return event, nil
}
