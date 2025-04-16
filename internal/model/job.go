package model

import "time"

type JobStatus string

const (
    StatusQueued     JobStatus = "QUEUED"
    StatusProcessing JobStatus = "PROCESSING"
    StatusCompleted  JobStatus = "COMPLETED"
    StatusFailed     JobStatus = "FAILED"
)

type Job struct {
    ID         string    `gorm:"primary_key"`
    Payload    string
    Status     JobStatus
    Result     string
    Retries    int
    MaxRetries int
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
