package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/namangarg/job-scheduler/internal/jobs"
	"github.com/namangarg/job-scheduler/pkg/job"
	"github.com/namangarg/job-scheduler/pkg/queue"
	"github.com/namangarg/job-scheduler/pkg/storage"
	"github.com/namangarg/job-scheduler/pkg/worker"
)

func main() {
	// Initialize Redis queue
	q, err := queue.NewRedisQueue("localhost:6379", "", 0)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	defer q.Close()

	// Initialize PostgreSQL storage
	s, err := storage.NewPostgresStorage("postgres://postgres:postgres@localhost:5432/jobs?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	// Create worker pool
	w := worker.NewWorker(q, s, 10) // 10 concurrent workers

	// Register job handlers
	w.RegisterHandler("email", jobs.HandleEmailJob)
	w.RegisterHandler("image_processing", jobs.HandleImageProcessingJob)
	w.RegisterHandler("data_sync", jobs.HandleDataSyncJob)

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx, "default")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Run benchmark
	startTime := time.Now()
	totalJobs := 50000
	jobsProcessed := 0
	jobsFailed := 0
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create and enqueue jobs
	for i := 0; i < totalJobs; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var j *job.Job

			// Randomly select job type
			switch rand.Intn(3) {
			case 0:
				j = jobs.CreateEmailJob(
					fmt.Sprintf("user%d@example.com", i),
					"Test Email",
					"This is a test email",
					nil,
				)
			case 1:
				j = jobs.CreateImageProcessingJob(
					fmt.Sprintf("https://example.com/images/%d.jpg", i),
					"resize",
					map[string]interface{}{"width": 800, "height": 600},
				)
			case 2:
				j = jobs.CreateDataSyncJob(
					"source_db",
					"destination_db",
					"users",
					rand.Intn(1000)+100,
				)
			}

			if err := q.Enqueue(ctx, j); err != nil {
				log.Printf("Failed to enqueue job: %v", err)
				mu.Lock()
				jobsFailed++
				mu.Unlock()
				return
			}
			mu.Lock()
			jobsProcessed++
			mu.Unlock()
		}(i)
	}

	// Wait for all jobs to be enqueued
	wg.Wait()

	// Wait for all jobs to complete or timeout
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			goto PrintResults
		case <-ticker.C:
			// Print progress
			log.Printf("Processed: %d, Failed: %d", jobsProcessed, jobsFailed)
		}
	}

PrintResults:
	duration := time.Since(startTime)
	jobsPerSecond := float64(jobsProcessed) / duration.Seconds()

	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("Total Jobs: %d\n", totalJobs)
	fmt.Printf("Jobs Processed: %d\n", jobsProcessed)
	fmt.Printf("Jobs Failed: %d\n", jobsFailed)
	fmt.Printf("Success Rate: %.2f%%\n", float64(jobsProcessed-jobsFailed)/float64(totalJobs)*100)
	fmt.Printf("Processing Time: %v\n", duration)
	fmt.Printf("Jobs per Second: %.2f\n", jobsPerSecond)
}
