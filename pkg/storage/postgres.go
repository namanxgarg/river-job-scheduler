package storage

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/namangarg/job-scheduler/pkg/job"
	"github.com/sirupsen/logrus"
)

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db:     db,
		logger: logrus.New(),
	}, nil
}

// runMigrations executes database migrations
func runMigrations(db *sql.DB) error {
	// Read and execute the migration script
	migrationSQL := `
	-- Drop the existing table
	DROP TABLE IF EXISTS jobs;

	-- Recreate the table with the updated schema
	CREATE TABLE jobs (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		payload JSONB NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		scheduled_for TIMESTAMP WITH TIME ZONE NOT NULL,
		started_at TIMESTAMP WITH TIME ZONE,
		completed_at TIMESTAMP WITH TIME ZONE,
		retry_count INTEGER NOT NULL DEFAULT 0,
		max_retries INTEGER NOT NULL DEFAULT 3,
		last_error TEXT,
		queue VARCHAR(255) NOT NULL,
		priority INTEGER NOT NULL DEFAULT 0,
		timeout BIGINT NOT NULL,
		unique_key VARCHAR(255),
		worker_id VARCHAR(255)
	);

	-- Recreate indexes
	CREATE INDEX idx_jobs_status ON jobs(status);
	CREATE INDEX idx_jobs_queue ON jobs(queue);
	CREATE INDEX idx_jobs_scheduled_for ON jobs(scheduled_for);
	CREATE INDEX idx_jobs_unique_key ON jobs(unique_key);
	`

	_, err := db.Exec(migrationSQL)
	return err
}

// Save persists a job to the database
func (s *PostgresStorage) Save(ctx context.Context, j *job.Job) error {
	query := `
	INSERT INTO jobs (
		id, name, payload, status, created_at, updated_at, scheduled_for,
		started_at, completed_at, retry_count, max_retries, last_error,
		queue, priority, timeout, unique_key, worker_id
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
	)
	ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		updated_at = EXCLUDED.updated_at,
		started_at = EXCLUDED.started_at,
		completed_at = EXCLUDED.completed_at,
		retry_count = EXCLUDED.retry_count,
		last_error = EXCLUDED.last_error,
		worker_id = EXCLUDED.worker_id
	`

	_, err := s.db.ExecContext(ctx, query,
		j.ID, j.Name, j.Payload, j.Status, j.CreatedAt, j.UpdatedAt, j.ScheduledFor,
		j.StartedAt, j.CompletedAt, j.RetryCount, j.MaxRetries, j.LastError,
		j.Queue, j.Priority, j.Timeout, j.UniqueKey, j.WorkerID,
	)

	return err
}

// Get retrieves a job by ID
func (s *PostgresStorage) Get(ctx context.Context, id string) (*job.Job, error) {
	query := `
	SELECT id, name, payload, status, created_at, updated_at, scheduled_for,
		started_at, completed_at, retry_count, max_retries, last_error,
		queue, priority, timeout, unique_key, worker_id
	FROM jobs WHERE id = $1
	`

	var j job.Job
	var startedAt, completedAt pq.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID, &j.Name, &j.Payload, &j.Status, &j.CreatedAt, &j.UpdatedAt, &j.ScheduledFor,
		&startedAt, &completedAt, &j.RetryCount, &j.MaxRetries, &j.LastError,
		&j.Queue, &j.Priority, &j.Timeout, &j.UniqueKey, &j.WorkerID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		j.CompletedAt = &completedAt.Time
	}

	return &j, nil
}

// List retrieves jobs based on filters
func (s *PostgresStorage) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*job.Job, error) {
	query := `
	SELECT id, name, payload, status, created_at, updated_at, scheduled_for,
		started_at, completed_at, retry_count, max_retries, last_error,
		queue, priority, timeout, unique_key, worker_id
	FROM jobs
	WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	for key, value := range filter {
		query += " AND " + key + " = $" + string(rune('0'+argCount))
		args = append(args, value)
		argCount++
	}

	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount)) +
		" OFFSET $" + string(rune('0'+argCount+1))
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*job.Job
	for rows.Next() {
		var j job.Job
		var startedAt, completedAt pq.NullTime

		err := rows.Scan(
			&j.ID, &j.Name, &j.Payload, &j.Status, &j.CreatedAt, &j.UpdatedAt, &j.ScheduledFor,
			&startedAt, &completedAt, &j.RetryCount, &j.MaxRetries, &j.LastError,
			&j.Queue, &j.Priority, &j.Timeout, &j.UniqueKey, &j.WorkerID,
		)
		if err != nil {
			return nil, err
		}

		if startedAt.Valid {
			j.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			j.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, &j)
	}

	return jobs, rows.Err()
}

// Delete removes a job from the database
func (s *PostgresStorage) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM jobs WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
