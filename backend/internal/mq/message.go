package mq

const (
	QueueBookingSuccess = "booking.success"
	QueueBookingTimeout = "booking.timeout"
	QueueSeatReleased   = "seat.released"
)

type BookingEvent struct {
	BookingID  string   `json:"booking_id"`
	UserID     string   `json:"user_id"`
	ShowtimeID string   `json:"showtime_id"`
	SeatNos    []string `json:"seat_nos"`
}
