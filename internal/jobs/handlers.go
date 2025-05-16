package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/namangarg/job-scheduler/pkg/job"
)

// HandleEmailJob processes an email sending job
func HandleEmailJob(ctx context.Context, j *job.Job) error {
	var emailJob EmailJob
	if err := json.Unmarshal(j.Payload, &emailJob); err != nil {
		return fmt.Errorf("failed to unmarshal email job: %v", err)
	}

	// Simulate email sending with varying processing times
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	// Simulate occasional failures (10% chance)
	if rand.Float32() < 0.1 {
		return fmt.Errorf("failed to send email to %s", emailJob.To)
	}

	return nil
}

// HandleImageProcessingJob processes an image processing job
func HandleImageProcessingJob(ctx context.Context, j *job.Job) error {
	var imgJob ImageProcessingJob
	if err := json.Unmarshal(j.Payload, &imgJob); err != nil {
		return fmt.Errorf("failed to unmarshal image job: %v", err)
	}

	// Simulate image processing with varying times based on operation
	switch imgJob.Operation {
	case "resize":
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
	case "compress":
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
	case "convert":
		time.Sleep(time.Duration(2000+rand.Intn(3000)) * time.Millisecond)
	default:
		return fmt.Errorf("unknown operation: %s", imgJob.Operation)
	}

	// Simulate occasional failures (5% chance)
	if rand.Float32() < 0.05 {
		return fmt.Errorf("failed to process image: %s", imgJob.ImageURL)
	}

	return nil
}

// HandleDataSyncJob processes a data synchronization job
func HandleDataSyncJob(ctx context.Context, j *job.Job) error {
	var syncJob DataSyncJob
	if err := json.Unmarshal(j.Payload, &syncJob); err != nil {
		return fmt.Errorf("failed to unmarshal sync job: %v", err)
	}

	// Simulate data sync with time proportional to batch size
	processingTime := time.Duration(syncJob.BatchSize*100) * time.Millisecond
	time.Sleep(processingTime)

	// Simulate occasional failures (15% chance)
	if rand.Float32() < 0.15 {
		return fmt.Errorf("failed to sync data from %s to %s", syncJob.Source, syncJob.Destination)
	}

	return nil
}
