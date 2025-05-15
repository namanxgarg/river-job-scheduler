package queue

import (
	"context"

	"github.com/namangarg/job-scheduler/pkg/job"
)

// Queue defines the interface for job queues
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, j *job.Job) error

	// Dequeue retrieves a job from the queue
	Dequeue(ctx context.Context, queueName string) (*job.Job, error)

	// Retry adds a job back to the queue for retry
	Retry(ctx context.Context, j *job.Job) error

	// Remove removes a job from the queue
	Remove(ctx context.Context, jobID string) error

	// Close closes the queue connection
	Close() error
}
