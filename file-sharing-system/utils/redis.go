package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client
var ctx = context.Background()


func ConnectRedis() {
    Rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379", 
        Password: "",           
        DB: 0,                  
    })

   
    _, err := Rdb.Ping(ctx).Result()
    if err != nil {
        log.Fatal("Unable to connect to Redis:", err)
    }
}


func SetCache(key string, value string, expiration time.Duration) error {
    return Rdb.Set(ctx, key, value, expiration).Err()
}


func GetCache(key string) (string, error) {
    return Rdb.Get(ctx, key).Result()
}


func DeleteCache(key string) error {
    return Rdb.Del(ctx, key).Err()
}


func RateLimiter(userID string, limit int, duration time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limiter:%s", userID)

	count, err := Rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	
	if count == 1 {
		err = Rdb.Expire(ctx, key, duration).Err()
		if err != nil {
			return false, err
		}
	}


	if count > int64(limit) {
		return false, errors.New("rate limit exceeded")
	}

	return true, nil
}