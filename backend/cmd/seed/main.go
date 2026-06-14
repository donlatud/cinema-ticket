package main

import (
	"context"
	"log"
	"os"

	"github.com/cinema-booking/backend/internal/config"
	"github.com/cinema-booking/backend/internal/database"
	"github.com/cinema-booking/backend/internal/seed"
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

	if err := seed.Run(ctx, db); err != nil {
		log.Fatal(err)
	}

	log.Println("seed finished successfully")
	os.Exit(0)
}
