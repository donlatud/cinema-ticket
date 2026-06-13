package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                string
	MongoURI            string
	JWTSecret           string
	FirebaseCredentials string
	RedisAddr           string
	RedisPassword       string
	RedisTLS            bool
	LockTTLSeconds      int
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

	redisAddrRaw := os.Getenv("REDIS_ADDR")
	if redisAddrRaw == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}

	useTLS := strings.HasPrefix(redisAddrRaw, "rediss://")
	if value := os.Getenv("REDIS_TLS"); value != "" {
		useTLS = strings.EqualFold(value, "true")
	}

	redisAddr := strings.TrimPrefix(redisAddrRaw, "redis://")
	redisAddr = strings.TrimPrefix(redisAddr, "rediss://")

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		return nil, fmt.Errorf("REDIS_PASSWORD is required")
	}

	lockTTLSeconds := 300
	if value := os.Getenv("LOCK_TTL_SECONDS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("LOCK_TTL_SECONDS must be a positive integer")
		}
		lockTTLSeconds = parsed
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
		RedisAddr:           redisAddr,
		RedisPassword:       redisPassword,
		RedisTLS:            useTLS,
		LockTTLSeconds:      lockTTLSeconds,
	}, nil
}
