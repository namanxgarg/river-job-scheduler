package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/yourusername/golang-job-scheduler/internal/api"
	"github.com/yourusername/golang-job-scheduler/internal/config"
	"github.com/yourusername/golang-job-scheduler/internal/jobs"
	"github.com/yourusername/golang-job-scheduler/internal/queue"
	"github.com/yourusername/golang-job-scheduler/internal/storage"
)

func main() {
	// Load .env
	godotenv.Load()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zlog.With().Str("service", "scheduler").Logger()

	cfg := config.LoadConfig()

	// Init DB
	db, err := storage.InitDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to DB")
	}
	defer db.Close()

	// AutoMigrate or run your own migrations
	storage.AutoMigrate(db)

	// Initialize Redis
	rdb := queue.InitRedis(cfg.RedisAddr, cfg.RedisPassword)
	defer rdb.Close()

	// Create store with DB
	store := storage.NewPGStore(db)

	// Create job service
	jobService := jobs.NewJobService(rdb, store)

	// Setup router
	r := mux.NewRouter()
	api.AttachSchedulerAPI(r, jobService)

	port := cfg.SchedulerPort
	if port == "" {
		port = "8080"
	}

	logger.Info().Msgf("Scheduler running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
