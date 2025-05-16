package job

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a job
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRetrying  Status = "retrying"
)

// Priority represents the job priority level
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// Job represents a task to be executed
type Job struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Payload        json.RawMessage `json:"payload"`
	Status         Status          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	ScheduledFor   time.Time       `json:"scheduled_for"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	RetryCount     int             `json:"retry_count"`
	MaxRetries     int             `json:"max_retries"`
	LastError      string          `json:"last_error,omitempty"`
	Queue          string          `json:"queue"`
	Priority       Priority        `json:"priority"`
	Timeout        time.Duration   `json:"timeout"`
	ProcessingTime time.Duration   `json:"processing_time,omitempty"`
	Deadline       time.Time       `json:"deadline,omitempty"`
	UniqueKey      string          `json:"unique_key,omitempty"`
	WorkerID       string          `json:"worker_id,omitempty"`
}

// NewJob creates a new job with default values
func NewJob(name string, payload json.RawMessage, queue string) *Job {
	now := time.Now()
	return &Job{
		ID:           uuid.New().String(),
		Name:         name,
		Payload:      payload,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		ScheduledFor: now,
		Queue:        queue,
		Priority:     PriorityNormal,
		MaxRetries:   3,
		Timeout:      5 * time.Minute,
		Deadline:     now.Add(5 * time.Minute),
	}
}

// WithRetries sets the maximum number of retries
func (j *Job) WithRetries(maxRetries int) *Job {
	j.MaxRetries = maxRetries
	return j
}

// WithTimeout sets the job timeout and deadline
func (j *Job) WithTimeout(timeout time.Duration) *Job {
	j.Timeout = timeout
	j.Deadline = time.Now().Add(timeout)
	return j
}

// WithPriority sets the job priority
func (j *Job) WithPriority(priority Priority) *Job {
	j.Priority = priority
	return j
}

// WithUniqueKey sets a unique key for deduplication
func (j *Job) WithUniqueKey(key string) *Job {
	j.UniqueKey = key
	return j
}

// WithSchedule sets the scheduled execution time
func (j *Job) WithSchedule(scheduledFor time.Time) *Job {
	j.ScheduledFor = scheduledFor
	return j
}

// ShouldRetry determines if the job should be retried
func (j *Job) ShouldRetry() bool {
	return j.RetryCount < j.MaxRetries
}

// NextRetryDelay calculates the next retry delay using exponential backoff
func (j *Job) NextRetryDelay() time.Duration {
	// Exponential backoff: 2^retryCount * baseDelay
	baseDelay := time.Second
	return time.Duration(1<<uint(j.RetryCount)) * baseDelay
}

// IsOverdue checks if the job has exceeded its deadline
func (j *Job) IsOverdue() bool {
	return !j.Deadline.IsZero() && time.Now().After(j.Deadline)
}
