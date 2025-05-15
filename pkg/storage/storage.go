package storage

import (
	"context"

	"github.com/namangarg/job-scheduler/pkg/job"
)

// Storage defines the interface for job storage
type Storage interface {
	// Save persists a job to storage
	Save(ctx context.Context, j *job.Job) error

	// Get retrieves a job by ID
	Get(ctx context.Context, id string) (*job.Job, error)

	// List retrieves jobs based on filters
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*job.Job, error)

	// Delete removes a job from storage
	Delete(ctx context.Context, id string) error

	// Close closes the storage connection
	Close() error
}
