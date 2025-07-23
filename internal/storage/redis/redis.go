package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlexMickh/proj-user/internal/models"
	"github.com/redis/go-redis/v9"
)

type Cash interface {
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LRange(ctx context.Context, key string, start int64, stop int64) *redis.StringSliceCmd
}

type Redis struct {
	rdb        Cash
	expiration time.Duration
}

func New(rdb Cash, expiration time.Duration) *Redis {
	return &Redis{
		rdb:        rdb,
		expiration: expiration,
	}
}

func (r *Redis) SaveUser(ctx context.Context, user models.User) error {
	const op = "storage.redis.SaveUser"

	err := r.saveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.rdb.RPush(ctx, genSkillsKey(user.ID), user.Skills).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.rdb.Expire(ctx, genSkillsKey(user.ID), r.expiration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Redis) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "storage.redis.UserByEmail"

	var cursor uint64
	var keys []string
	var err error
	keys, cursor, err = r.rdb.Scan(ctx, cursor, "*&"+email, 2).Result()
	if err != nil || len(keys) == 0 {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = r.rdb.HGetAll(ctx, keys[0]).Scan(&user)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	arr := strings.Split(keys[0], "&")
	user.ID = arr[0]
	user.Email = arr[1]

	user.Skills, err = r.rdb.LRange(ctx, genSkillsKey(user.ID), 0, -1).Result()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = r.rdb.Expire(ctx, keys[0], r.expiration).Err()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = r.rdb.Expire(ctx, genSkillsKey(user.ID), r.expiration).Err()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *Redis) UpdateUser(ctx context.Context, user models.User) error {
	const op = "storage.redis.UpdateUser"

	if err := r.saveUser(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Redis) saveUser(ctx context.Context, user models.User) error {
	const op = "storage.redis.saveUser"

	key := genKey(user.ID, user.Email)

	err := r.rdb.HSet(ctx, key, user).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.rdb.Expire(ctx, key, r.expiration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func genKey(id, email string) string {
	return id + "&" + email
}

func genSkillsKey(id string) string {
	return "skills:" + id
}
