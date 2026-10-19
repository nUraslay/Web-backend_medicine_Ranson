package redis

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"ranson-backend/internal/app/config"
)

const servicePrefix = "ranson."

type Client struct {
	client *goredis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	c := goredis.NewClient(&goredis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к redis: %w", err)
	}

	return &Client{client: c}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
