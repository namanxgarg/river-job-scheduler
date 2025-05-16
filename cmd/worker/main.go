package main

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"os"

	"github.com/yourusername/golang-job-scheduler/internal/config"
	"github.com/yourusername/golang-job-scheduler/internal/jobs"
	"github.com/yourusername/golang-job-scheduler/internal/queue"
	"github.com/yourusername/golang-job-scheduler/internal/storage"
)

func main() {
	godotenv.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zlog.With().Str("service", "worker").Logger()

	cfg := config.LoadConfig()

	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = "worker-default"
	}

	// Init DB
	db, err := storage.InitDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect DB")
	}
	defer db.Close()

	// No migration needed if scheduler did it, but up to you
	storage.AutoMigrate(db)

	rdb := queue.InitRedis(cfg.RedisAddr, cfg.RedisPassword)
	defer rdb.Close()

	store := storage.NewPGStore(db)

	jobService := jobs.NewJobService(rdb, store)

	logger.Info().Msgf("[%s] Worker starting...", workerID)
	jobService.StartWorker(workerID)

	// Keep worker alive
	for {
		time.Sleep(15 * time.Second)
	}
}
