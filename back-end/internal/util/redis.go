package util

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func StoreToken(rdb *redis.Client, userID string, token string, expiry time.Duration) error {
	return rdb.Set(context.Background(), "refresh:"+userID, token, expiry).Err()
}

func GetToken(rdb *redis.Client, userID string) (string, error) {
	val, err := rdb.Get(context.Background(), "refresh:"+userID).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func DeleteToken(rdb *redis.Client, userID string) error {
	return rdb.Del(context.Background(), "refresh:"+userID).Err()
}
