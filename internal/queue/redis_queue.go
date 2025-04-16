package queue

import (
    "context"

    "github.com/go-redis/redis/v8"
    "github.com/rs/zerolog/log"
)

var ctx = context.Background()

func InitRedis(addr, pass string) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: pass,
        DB:       0,
    })
    if _, err := rdb.Ping(ctx).Result(); err != nil {
        log.Fatal().Err(err).Msg("Failed to connect redis")
    }
    return rdb
}

func PushTask(rdb *redis.Client, queueName, data string) error {
    return rdb.RPush(ctx, queueName, data).Err()
}

func PopTask(rdb *redis.Client, queueName string) (string, error) {
    val, err := rdb.BLPop(ctx, 0, queueName).Result()
    if err != nil {
        return "", err
    }
    if len(val) < 2 {
        return "", nil
    }
    return val[1], nil
}
