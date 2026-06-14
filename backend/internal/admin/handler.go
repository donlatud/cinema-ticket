package admin

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	dashboard *DashboardService
}

func NewHandler(dashboard *DashboardService) *Handler {
	return &Handler{dashboard: dashboard}
}

func (h *Handler) ListBookings(c *gin.Context) {
	filter, err := parseBookingFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.dashboard.ListBookings(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrFilterRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one filter is required: movie_id, date, or user_id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": items})
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	filter, err := parseAuditFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, err := h.dashboard.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

func parseBookingFilter(c *gin.Context) (BookingListFilter, error) {
	var filter BookingListFilter
	hasFilter := false

	if movieIDHex := c.Query("movie_id"); movieIDHex != "" {
		movieID, err := primitive.ObjectIDFromHex(movieIDHex)
		if err != nil {
			return filter, errors.New("invalid movie_id")
		}
		filter.MovieID = &movieID
		hasFilter = true
	}

	if userIDHex := c.Query("user_id"); userIDHex != "" {
		userID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			return filter, errors.New("invalid user_id")
		}
		filter.UserID = &userID
		hasFilter = true
	}

	if dateValue := c.Query("date"); dateValue != "" {
		day, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return filter, errors.New("invalid date format, use YYYY-MM-DD")
		}
		utcDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
		filter.Date = &utcDay
		hasFilter = true
	}

	if !hasFilter {
		return filter, ErrFilterRequired
	}

	return filter, nil
}

func parseAuditFilter(c *gin.Context) (AuditListFilter, error) {
	filter := AuditListFilter{Event: c.Query("event")}

	if fromValue := c.Query("from"); fromValue != "" {
		from, err := time.Parse(time.RFC3339, fromValue)
		if err != nil {
			day, dayErr := time.Parse("2006-01-02", fromValue)
			if dayErr != nil {
				return filter, errors.New("invalid from format, use YYYY-MM-DD or RFC3339")
			}
			from = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
		}
		filter.From = &from
	}

	if toValue := c.Query("to"); toValue != "" {
		to, err := time.Parse(time.RFC3339, toValue)
		if err != nil {
			day, dayErr := time.Parse("2006-01-02", toValue)
			if dayErr != nil {
				return filter, errors.New("invalid to format, use YYYY-MM-DD or RFC3339")
			}
			to = time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.UTC)
		}
		filter.To = &to
	}

	return filter, nil
}
