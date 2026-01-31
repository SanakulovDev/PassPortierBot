package security

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient initializes a new Redis client using environment variables.
func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
