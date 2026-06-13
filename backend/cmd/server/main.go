package main

import (
	"context"
	"log"

	"github.com/cinema-booking/backend/internal/auth"
	"github.com/cinema-booking/backend/internal/config"
	"github.com/cinema-booking/backend/internal/database"
	"github.com/cinema-booking/backend/internal/repository"
	"github.com/cinema-booking/backend/internal/router"
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

	verifier, err := auth.NewFirebaseVerifier(ctx, cfg.FirebaseCredentials)
	if err != nil {
		log.Fatal(err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	authHandler := auth.NewHandler(verifier, jwtService, userRepo)

	r := router.Setup(authHandler, jwtService)

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
