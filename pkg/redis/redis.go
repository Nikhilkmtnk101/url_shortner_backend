package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nikhil/url-shortner-backend/config"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	instance *Client
	once     sync.Once
)

type Client struct {
	client *redis.Client
}

// GetRedisClient returns a singleton instance of RedisClient
func GetRedisClient(config config.RedisConfig) (*Client, error) {
	var err error
	once.Do(func() {
		instance, err = newRedisClient(config)
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// newRedisClient initializes and returns a new Redis client
func newRedisClient(config config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Host, config.Port),
	})

	// Test the connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	return &Client{client: client}, nil
}

func (r *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %v", err)
		}
		strValue = string(jsonBytes)
	}

	err := r.client.Set(ctx, key, strValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	return nil
}

func (r *Client) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s not found", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s: %v", key, err)
	}

	return value, nil
}

func (r *Client) GetWithUnmarshal(ctx context.Context, key string, dest interface{}) error {
	value, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(value), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %v", err)
	}

	return nil
}

func (r *Client) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %v", key, err)
	}

	return nil
}

func (r *Client) Close() error {
	return r.client.Close()
}

func (r *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence %s: %v", key, err)
	}

	return result == 1, nil
}

func (r *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("failed to marshal value: %v", err)
		}
		strValue = string(jsonBytes)
	}

	success, err := r.client.SetNX(ctx, key, strValue, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set key %s: %v", key, err)
	}

	return success, nil
}
