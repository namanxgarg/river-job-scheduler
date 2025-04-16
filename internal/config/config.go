package config

import (
    "os"
)

type Config struct {
    DBHost        string
    DBPort        string
    DBUser        string
    DBPass        string
    DBName        string

    RedisAddr     string
    RedisPassword string

    SchedulerPort string
}

func LoadConfig() *Config {
    return &Config{
        DBHost:        getEnv("DB_HOST", "localhost"),
        DBPort:        getEnv("DB_PORT", "5432"),
        DBUser:        getEnv("DB_USER", "scheduler_user"),
        DBPass:        getEnv("DB_PASS", "scheduler_pass"),
        DBName:        getEnv("DB_NAME", "scheduler_db"),
        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword: os.Getenv("REDIS_PASSWORD"),
        SchedulerPort: os.Getenv("SCHEDULER_PORT"),
    }
}

func getEnv(key, fallback string) string {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    return v
}
