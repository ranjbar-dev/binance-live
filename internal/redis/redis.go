package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/binance-live/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Client wraps the Redis client
type Client struct {
	client *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// New creates a new Redis client
func New(cfg *config.RedisConfig, logger *zap.Logger) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established",
		zap.String("addr", cfg.GetRedisAddr()),
	)

	return &Client{
		client: client,
		logger: logger,
		ttl:    time.Duration(cfg.LiveDataTTL) * time.Second,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// PublishJSON publishes a JSON message to a channel
func (c *Client) PublishJSON(ctx context.Context, channel string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := c.client.Publish(ctx, channel, jsonData).Err(); err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	return nil
}

// PublishProtobuf publishes a protobuf message to a channel
func (c *Client) PublishProtobuf(ctx context.Context, channel string, data proto.Message) error {
	protoData, err := proto.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf data: %w", err)
	}

	if err := c.client.Publish(ctx, channel, protoData).Err(); err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	return nil
}

// SetJSON sets a key with JSON value and TTL
func (c *Client) SetJSON(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if ttl == 0 {
		ttl = c.ttl
	}

	if err := c.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key in Redis: %w", err)
	}

	return nil
}

// SetProtobuf sets a key with protobuf value and TTL
func (c *Client) SetProtobuf(ctx context.Context, key string, data proto.Message, ttl time.Duration) error {
	protoData, err := proto.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf data: %w", err)
	}

	if ttl == 0 {
		ttl = c.ttl
	}

	if err := c.client.Set(ctx, key, protoData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key in Redis: %w", err)
	}

	return nil
}

// GetJSON gets a key and unmarshals JSON value
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key from Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// GetProtobuf gets a key and unmarshals protobuf value
func (c *Client) GetProtobuf(ctx context.Context, key string, dest proto.Message) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key from Redis: %w", err)
	}

	if err := proto.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}

	return nil
}

// SetHash sets multiple fields in a hash
func (c *Client) SetHash(ctx context.Context, key string, fields map[string]interface{}) error {
	if err := c.client.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("failed to set hash: %w", err)
	}

	// Set TTL on the hash
	if err := c.client.Expire(ctx, key, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// GetHash gets all fields from a hash
func (c *Client) GetHash(ctx context.Context, key string) (map[string]string, error) {
	result, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash: %w", err)
	}

	return result, nil
}

// Delete deletes keys
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result > 0, nil
}

// HealthCheck checks if Redis is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
