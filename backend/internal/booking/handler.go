package booking

import (
	"errors"
	"net/http"

	"github.com/cinema-booking/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Pay(c *gin.Context) {
	bookingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.service.Pay(c.Request.Context(), bookingID, userID)
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": result})
}

func (h *Handler) Cancel(c *gin.Context) {
	bookingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.service.Cancel(c.Request.Context(), bookingID, userID)
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": result})
}

func (h *Handler) ListMy(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	bookings, err := h.service.ListMyBookings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func userIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(c.GetString(auth.ContextUserIDKey))
}

func writeBookingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrBookingNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
	case errors.Is(err, ErrBookingForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, ErrBookingNotPending):
		c.JSON(http.StatusConflict, gin.H{"error": "booking is not pending"})
	case errors.Is(err, ErrBookingExpired):
		c.JSON(http.StatusGone, gin.H{"error": "booking expired"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking operation failed"})
	}
}
