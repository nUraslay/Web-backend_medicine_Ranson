package redis

import (
	"context"
	"time"
)

const jwtBlacklistPrefix = "jwt.blacklist."

func getJWTKey(token string) string {
	return servicePrefix + jwtBlacklistPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return c.client.Set(ctx, getJWTKey(token), "1", ttl).Err()
}

func (c *Client) IsJWTBlacklisted(ctx context.Context, token string) (bool, error) {
	n, err := c.client.Exists(ctx, getJWTKey(token)).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
