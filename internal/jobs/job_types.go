package jobs

import (
	"encoding/json"
	"time"

	"github.com/namangarg/job-scheduler/pkg/job"
)

// EmailJob represents an email sending task
type EmailJob struct {
	To      string   `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	CC      []string `json:"cc,omitempty"`
}

// ImageProcessingJob represents an image processing task
type ImageProcessingJob struct {
	ImageURL   string                 `json:"image_url"`
	Operation  string                 `json:"operation"` // resize, compress, convert
	Parameters map[string]interface{} `json:"parameters"`
}

// DataSyncJob represents a data synchronization task
type DataSyncJob struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	DataType    string `json:"data_type"`
	BatchSize   int    `json:"batch_size"`
}

// CreateEmailJob creates a new email job
func CreateEmailJob(to, subject, body string, cc []string) *job.Job {
	payload, _ := json.Marshal(EmailJob{
		To:      to,
		Subject: subject,
		Body:    body,
		CC:      cc,
	})

	return job.NewJob("email", payload, "email_queue").
		WithPriority(job.PriorityHigh).
		WithTimeout(2 * time.Minute)
}

// CreateImageProcessingJob creates a new image processing job
func CreateImageProcessingJob(imageURL, operation string, params map[string]interface{}) *job.Job {
	payload, _ := json.Marshal(ImageProcessingJob{
		ImageURL:   imageURL,
		Operation:  operation,
		Parameters: params,
	})

	return job.NewJob("image_processing", payload, "image_queue").
		WithPriority(job.PriorityNormal).
		WithTimeout(5 * time.Minute)
}

// CreateDataSyncJob creates a new data sync job
func CreateDataSyncJob(source, destination, dataType string, batchSize int) *job.Job {
	payload, _ := json.Marshal(DataSyncJob{
		Source:      source,
		Destination: destination,
		DataType:    dataType,
		BatchSize:   batchSize,
	})

	return job.NewJob("data_sync", payload, "sync_queue").
		WithPriority(job.PriorityLow).
		WithTimeout(15 * time.Minute)
}
