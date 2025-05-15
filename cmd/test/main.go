package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/namangarg/job-scheduler/pkg/job"
	"github.com/namangarg/job-scheduler/pkg/queue"
	"github.com/namangarg/job-scheduler/pkg/storage"
	"github.com/namangarg/job-scheduler/pkg/worker"
)

func main() {
	// Initialize Redis queue
	redisQueue, err := queue.NewRedisQueue("localhost:6379", "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisQueue.Close()

	// Initialize PostgreSQL storage
	postgresStorage, err := storage.NewPostgresStorage("postgres://postgres:postgres@localhost:5432/jobs?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresStorage.Close()

	// Create worker
	w := worker.NewWorker(redisQueue, postgresStorage, 5)

	// Register job handlers
	w.RegisterHandler("print_message", handlePrintMessage)
	w.RegisterHandler("delayed_job", handleDelayedJob)

	// Start worker
	ctx := context.Background()
	w.Start(ctx, "default")

	// Create and enqueue some test jobs
	createTestJobs(redisQueue)

	// Wait for jobs to complete
	time.Sleep(10 * time.Second)

	// Stop worker
	w.Stop()
}

func createTestJobs(q queue.Queue) {
	// Create an immediate job
	immediateJob := job.NewJob("print_message", json.RawMessage(`{"message": "Hello, World!"}`), "default")
	if err := q.Enqueue(context.Background(), immediateJob); err != nil {
		log.Printf("Failed to enqueue immediate job: %v", err)
	}

	// Create a delayed job
	delayedJob := job.NewJob("delayed_job", json.RawMessage(`{"message": "This is a delayed job"}`), "default").
		WithSchedule(time.Now().Add(5 * time.Second))
	if err := q.Enqueue(context.Background(), delayedJob); err != nil {
		log.Printf("Failed to enqueue delayed job: %v", err)
	}

	// Create a job with retries
	retryJob := job.NewJob("print_message", json.RawMessage(`{"message": "This job will fail and retry"}`), "default").
		WithRetries(3)
	if err := q.Enqueue(context.Background(), retryJob); err != nil {
		log.Printf("Failed to enqueue retry job: %v", err)
	}
}

func handlePrintMessage(ctx context.Context, j *job.Job) error {
	var data struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return err
	}

	// Simulate a failing job for the retry test
	if data.Message == "This job will fail and retry" {
		return fmt.Errorf("simulated failure")
	}

	log.Printf("Processing message: %s", data.Message)
	return nil
}

func handleDelayedJob(ctx context.Context, j *job.Job) error {
	var data struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return err
	}

	log.Printf("Processing delayed job: %s", data.Message)
	return nil
}
