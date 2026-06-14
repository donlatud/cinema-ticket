package main

import (
	"context"
	"log"
	"time"

	"github.com/cinema-booking/backend/internal/admin"
	"github.com/cinema-booking/backend/internal/auth"
	"github.com/cinema-booking/backend/internal/booking"
	"github.com/cinema-booking/backend/internal/config"
	"github.com/cinema-booking/backend/internal/database"
	"github.com/cinema-booking/backend/internal/lock"
	"github.com/cinema-booking/backend/internal/mq"
	"github.com/cinema-booking/backend/internal/realtime"
	"github.com/cinema-booking/backend/internal/repository"
	"github.com/cinema-booking/backend/internal/router"
	"github.com/cinema-booking/backend/internal/seat"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := database.ConnectRedis(ctx, database.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		UseTLS:   cfg.RedisTLS,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer mqClient.Close()

	verifier, err := auth.NewFirebaseVerifier(ctx, cfg.FirebaseCredentials)
	if err != nil {
		log.Fatal(err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	authHandler := auth.NewHandler(verifier, jwtService, userRepo)

	showtimeRepo := repository.NewShowtimeRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	auditService := admin.NewAuditService(auditRepo)
	dashboardService := admin.NewDashboardService(bookingRepo, showtimeRepo, movieRepo, userRepo, auditRepo)
	adminHandler := admin.NewHandler(dashboardService)
	producer := mq.NewProducer(mqClient.Channel())
	consumer := mq.NewConsumer(mqClient.Channel(), auditService)
	consumer.Start(ctx)

	seatLock := lock.NewRedisLock(redisClient, time.Duration(cfg.LockTTLSeconds)*time.Second)
	hub := realtime.NewHub()
	bookingService := booking.NewService(
		showtimeRepo, seatRepo, bookingRepo, movieRepo,
		seatLock, hub, producer, auditService,
	)
	booking.StartExpiryWorker(ctx, bookingService)

	seatHandler := seat.NewHandler(bookingService)
	bookingHandler := booking.NewHandler(bookingService)
	wsHandler := realtime.NewHandler(hub)

	r := router.Setup(authHandler, jwtService, seatHandler, bookingHandler, wsHandler, adminHandler)

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
