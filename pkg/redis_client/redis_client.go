package redis_client

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexMickh/proj-user/pkg/utils/retry"
	"github.com/redis/go-redis/v9"
)

func New(
	ctx context.Context,
	addr string,
	user string,
	password string,
	db int,
) (*redis.Client, error) {
	const op = "redis-client.New"

	var rdb *redis.Client

	err := retry.WithDelay(5, 500*time.Millisecond, func() error {
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})

		err := rdb.Ping(ctx).Err()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rdb, nil
}
