package router

import (
	"net/http"

	"github.com/cinema-booking/backend/internal/auth"
	bookinghandler "github.com/cinema-booking/backend/internal/booking"
	"github.com/cinema-booking/backend/internal/admin"
	"github.com/cinema-booking/backend/internal/model"
	"github.com/cinema-booking/backend/internal/realtime"
	"github.com/cinema-booking/backend/internal/seat"
	"github.com/gin-gonic/gin"
)

func Setup(
	authHandler *auth.Handler,
	jwtService *auth.JWTService,
	seatHandler *seat.Handler,
	bookingHandler *bookinghandler.Handler,
	wsHandler *realtime.Handler,
	adminHandler *admin.Handler,
) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ws", wsHandler.ServeWS)

	api := r.Group("/api")
	api.POST("/auth/login", authHandler.Login)

	api.GET("/showtimes", seatHandler.ListShowtimes)
	api.GET("/showtimes/:id/seats", seatHandler.ListSeats)

	protected := api.Group("", auth.AuthMiddleware(jwtService))
	protected.GET("/me", authHandler.Me)
	protected.POST("/showtimes/:id/seats/lock", seatHandler.LockSeats)
	protected.POST("/bookings/:id/pay", bookingHandler.Pay)
	protected.POST("/bookings/:id/cancel", bookingHandler.Cancel)
	protected.GET("/bookings/my", bookingHandler.ListMy)

	admin := api.Group("/admin", auth.AuthMiddleware(jwtService), auth.RequireRole(model.RoleAdmin))
	admin.GET("/bookings", adminHandler.ListBookings)
	admin.GET("/audit-logs", adminHandler.ListAuditLogs)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
