package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                 string
	MongoURI             string
	JWTSecret            string
	FirebaseCredentials  string
}

func Load() (*Config, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")
	if firebaseCredentials == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:                port,
		MongoURI:            mongoURI,
		JWTSecret:           jwtSecret,
		FirebaseCredentials: firebaseCredentials,
	}, nil
}
