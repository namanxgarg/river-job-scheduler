package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/namangarg/job-scheduler/pkg/job"
	"github.com/namangarg/job-scheduler/pkg/queue"
	"github.com/namangarg/job-scheduler/pkg/storage"
	"github.com/namangarg/job-scheduler/pkg/worker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for monitoring job processing
var (
	// Track how long each job takes to process
	jobLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "job_processing_latency_seconds",
		Help:    "Time taken to process jobs",
		Buckets: prometheus.DefBuckets,
	}, []string{"job_type"})

	// Count successful jobs
	jobSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "job_success_total",
		Help: "Total number of successful jobs",
	}, []string{"job_type"})

	// Count failed jobs
	jobFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "job_failure_total",
		Help: "Total number of failed jobs",
	}, []string{"job_type"})
)

// Define our job types with their payloads
type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type ImageProcessingJob struct {
	ImageURL   string                 `json:"image_url"`
	Operation  string                 `json:"operation"` // resize, compress, filter
	Parameters map[string]interface{} `json:"parameters"`
}

type DataSyncJob struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Table       string `json:"table"`
	BatchSize   int    `json:"batch_size"`
}

func main() {
	// Connect to Redis for job queuing
	redisQueue, err := queue.NewRedisQueue("localhost:6379", "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisQueue.Close()

	// Connect to PostgreSQL for job storage
	postgresStorage, err := storage.NewPostgresStorage("postgres://postgres:postgres@localhost:5432/jobs?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresStorage.Close()

	// Create a worker pool with 10 concurrent workers
	w := worker.NewWorker(redisQueue, postgresStorage, 10)

	// Register handlers for different job types
	w.RegisterHandler("send_email", handleEmailJob)
	w.RegisterHandler("process_image", handleImageProcessingJob)
	w.RegisterHandler("sync_data", handleDataSyncJob)

	// Start the worker
	ctx := context.Background()
	w.Start(ctx, "default")

	// Run our benchmark
	runBenchmark(redisQueue)

	// Wait for jobs to complete
	time.Sleep(30 * time.Second)

	// Stop the worker
	w.Stop()
}

func runBenchmark(q queue.Queue) {
	var wg sync.WaitGroup
	startTime := time.Now()
	totalJobs := 50000 // We'll process 50k jobs

	// Create and enqueue jobs
	for i := 0; i < totalJobs; i++ {
		wg.Add(1)
		go func(jobNum int) {
			defer wg.Done()

			// Randomly choose a job type
			jobType := rand.Intn(3)
			var j *job.Job

			switch jobType {
			case 0:
				// Create an email job
				emailJob := EmailJob{
					To:      fmt.Sprintf("user%d@example.com", jobNum),
					Subject: "Test Email",
					Body:    "This is a test email",
				}
				payload, _ := json.Marshal(emailJob)
				j = job.NewJob("send_email", payload, "default")

			case 1:
				// Create an image processing job
				imageJob := ImageProcessingJob{
					ImageURL:  fmt.Sprintf("https://example.com/images/%d.jpg", jobNum),
					Operation: "resize",
					Parameters: map[string]interface{}{
						"width":  800,
						"height": 600,
					},
				}
				payload, _ := json.Marshal(imageJob)
				j = job.NewJob("process_image", payload, "default")

			case 2:
				// Create a data sync job
				syncJob := DataSyncJob{
					Source:      "source_db",
					Destination: "target_db",
					Table:       "users",
					BatchSize:   1000,
				}
				payload, _ := json.Marshal(syncJob)
				j = job.NewJob("sync_data", payload, "default")
			}

			// Add retry capability to some jobs
			if rand.Float32() < 0.1 { // 10% of jobs will have retries
				j.WithRetries(3)
			}

			// Add delay to some jobs
			if rand.Float32() < 0.2 { // 20% of jobs will be delayed
				// Use a maximum delay of 5 minutes (300 seconds)
				delay := time.Duration(rand.Intn(300)) * time.Second
				// Ensure the timestamp is within PostgreSQL's integer range
				scheduledTime := time.Now().Add(delay)
				if scheduledTime.Unix() > 2147483647 { // PostgreSQL integer max
					scheduledTime = time.Unix(2147483647, 0)
				}
				j.WithSchedule(scheduledTime)
			}

			// Add the job to the queue
			if err := q.Enqueue(context.Background(), j); err != nil {
				log.Printf("Failed to enqueue job: %v", err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Calculate and display results
	jobsPerSecond := float64(totalJobs) / duration.Seconds()
	log.Printf("\nBenchmark Results:")
	log.Printf("==================")
	log.Printf("Total Jobs: %d", totalJobs)
	log.Printf("Duration: %v", duration)
	log.Printf("Jobs per second: %.2f", jobsPerSecond)
	log.Printf("Average latency: %.2f ms", float64(duration.Milliseconds())/float64(totalJobs))
	log.Printf("Concurrent workers: 10")
}

func handleEmailJob(ctx context.Context, j *job.Job) error {
	startTime := time.Now()
	defer func() {
		jobLatency.WithLabelValues("send_email").Observe(time.Since(startTime).Seconds())
	}()

	var emailJob EmailJob
	if err := json.Unmarshal(j.Payload, &emailJob); err != nil {
		jobFailure.WithLabelValues("send_email").Inc()
		return err
	}

	// Simulate email sending
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// Simulate occasional failures (5% of emails fail)
	if rand.Float32() < 0.05 {
		jobFailure.WithLabelValues("send_email").Inc()
		return fmt.Errorf("failed to send email to %s", emailJob.To)
	}

	jobSuccess.WithLabelValues("send_email").Inc()
	return nil
}

func handleImageProcessingJob(ctx context.Context, j *job.Job) error {
	startTime := time.Now()
	defer func() {
		jobLatency.WithLabelValues("process_image").Observe(time.Since(startTime).Seconds())
	}()

	var imageJob ImageProcessingJob
	if err := json.Unmarshal(j.Payload, &imageJob); err != nil {
		jobFailure.WithLabelValues("process_image").Inc()
		return err
	}

	// Simulate image processing
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	// Simulate occasional failures (3% of image processing fails)
	if rand.Float32() < 0.03 {
		jobFailure.WithLabelValues("process_image").Inc()
		return fmt.Errorf("failed to process image %s", imageJob.ImageURL)
	}

	jobSuccess.WithLabelValues("process_image").Inc()
	return nil
}

func handleDataSyncJob(ctx context.Context, j *job.Job) error {
	startTime := time.Now()
	defer func() {
		jobLatency.WithLabelValues("sync_data").Observe(time.Since(startTime).Seconds())
	}()

	var syncJob DataSyncJob
	if err := json.Unmarshal(j.Payload, &syncJob); err != nil {
		jobFailure.WithLabelValues("sync_data").Inc()
		return err
	}

	// Simulate data sync
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// Simulate occasional failures (2% of syncs fail)
	if rand.Float32() < 0.02 {
		jobFailure.WithLabelValues("sync_data").Inc()
		return fmt.Errorf("failed to sync data from %s to %s", syncJob.Source, syncJob.Destination)
	}

	jobSuccess.WithLabelValues("sync_data").Inc()
	return nil
}
