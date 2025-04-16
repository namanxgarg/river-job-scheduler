package jobs

import (
	"encoding/json"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/golang-job-scheduler/internal/model"
	"github.com/yourusername/golang-job-scheduler/internal/queue"
	"github.com/yourusername/golang-job-scheduler/internal/storage"
)

// DefaultQueue is the Redis list name used to queue jobs.
const DefaultQueue = "jobs"

// JobService encapsulates scheduling, retrieving, and processing jobs.
type JobService struct {
	rdb   *redis.Client
	store storage.Store // store interface that deals with model.Job
}

// NewJobService returns a new instance of JobService.
func NewJobService(rdb *redis.Client, store storage.Store) *JobService {
	return &JobService{
		rdb:   rdb,
		store: store,
	}
}

// ScheduleJob creates a new job, saves it in the store (e.g., Postgres),
// and pushes the job to the Redis queue.
func (js *JobService) ScheduleJob(payload string) (string, error) {
	j := model.Job{
		ID:         uuid.New().String(),
		Payload:    payload,
		Status:     model.StatusQueued,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := js.store.SaveJob(&j); err != nil {
		return "", err
	}

	data, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	if err := queue.PushTask(js.rdb, DefaultQueue, string(data)); err != nil {
		return "", err
	}

	log.Info().Msgf("Scheduled job %s with payload: %s", j.ID, j.Payload)
	return j.ID, nil
}

// GetJob retrieves a job by its ID from the underlying store.
func (js *JobService) GetJob(id string) (*model.Job, error) {
	return js.store.GetJob(id)
}

// StartWorker begins a loop that continuously pops jobs from the Redis queue
// and processes them using an exponential backoff strategy for retries.
func (js *JobService) StartWorker(workerID string) {
	go func() {
		for {
			taskStr, err := queue.PopTask(js.rdb, DefaultQueue)
			if err != nil {
				log.Error().Err(err).Msgf("[%s] PopTask failed", workerID)
				time.Sleep(3 * time.Second)
				continue
			}
			if taskStr == "" {
				continue
			}

			var j model.Job
			if err := json.Unmarshal([]byte(taskStr), &j); err != nil {
				log.Error().Err(err).Msgf("[%s] Failed to unmarshal job", workerID)
				continue
			}
			js.processJob(&j, workerID)
		}
	}()
}

// processJob runs the provided task using exponential backoff. If the task fails
// after the maximum number of retries, the job status is set to FAILED.
// func (js *JobService) processJob(j *Job, workerID string) {
//     log.Info().Msgf("[%s] Processing job %s", workerID, j.ID)

//     j.Status = StatusProcessing
//     j.UpdatedAt = time.Now()
//     if err := js.store.SaveJob(j); err != nil {
//         log.Error().Err(err).Msgf("[%s] updating job status to PROCESSING failed", workerID)
//         return
//     }

//     // The operation to retry
//     operation := func() error {
//         result, err := runExampleTask(j.Payload)
//         if err != nil {
//             return err
//         }
//         j.Result = result
//         return nil
//     }

//     // Configure exponential backoff policy
//     b := backoff.NewExponentialBackOff()
//     b.InitialInterval = 1 * time.Second
//     b.MaxInterval = 5 * time.Second
//     b.MaxElapsedTime = 15 * time.Second

//     // Convert j.MaxRetries (assumed as int) to uint64 for backoff.WithMaxRetries.
//     err := backoff.Retry(operation, backoff.WithMaxRetries(b, uint64(j.MaxRetries)))
//     if err != nil {
//         j.Status = StatusFailed
//         j.Result = "Failed after retries: " + err.Error()
//         j.Retries = j.MaxRetries
//         j.UpdatedAt = time.Now()
//         _ = js.store.SaveJob(j)
//         log.Error().Msgf("[%s] job %s permanently failed: %v", workerID, j.ID, err)
//         return
//     }

//     j.Status = StatusCompleted
//     j.UpdatedAt = time.Now()
//     if err := js.store.SaveJob(j); err != nil {
//         log.Error().Err(err).Msgf("[%s] failed to save job result for job %s", workerID, j.ID)
//         return
//     }
//     log.Info().Msgf("[%s] job %s completed", workerID, j.ID)
// }


func (js *JobService) processJob(j *model.Job, workerID string) {
    log.Info().Msgf("[%s] Processing job %s", workerID, j.ID)

    // Set job status to processing using the constant from the model package.
    j.Status = model.StatusProcessing
    j.UpdatedAt = time.Now()
    if err := js.store.SaveJob(j); err != nil {
        log.Error().Err(err).Msgf("[%s] updating job status to PROCESSING failed", workerID)
        return
    }

    // The operation to attempt with exponential backoff
    operation := func() error {
        result, err := runExampleTask(j.Payload)
        if err != nil {
            return err
        }
        j.Result = result
        return nil
    }

    // Configure exponential backoff policy
    b := backoff.NewExponentialBackOff()
    b.InitialInterval = 1 * time.Second
    b.MaxInterval = 5 * time.Second
    b.MaxElapsedTime = 15 * time.Second

    // Convert j.MaxRetries (assumed as int) to uint64 for backoff.WithMaxRetries.
    err := backoff.Retry(operation, backoff.WithMaxRetries(b, uint64(j.MaxRetries)))
    if err != nil {
        j.Status = model.StatusFailed
        j.Result = "Failed after retries: " + err.Error()
        j.Retries = j.MaxRetries
        j.UpdatedAt = time.Now()
        _ = js.store.SaveJob(j)
        log.Error().Msgf("[%s] job %s permanently failed: %v", workerID, j.ID, err)
        return
    }

    j.Status = model.StatusCompleted
    j.UpdatedAt = time.Now()
    if err := js.store.SaveJob(j); err != nil {
        log.Error().Err(err).Msgf("[%s] failed to save job result for job %s", workerID, j.ID)
        return
    }
    log.Info().Msgf("[%s] job %s completed", workerID, j.ID)
}
