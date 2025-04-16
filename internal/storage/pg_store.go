package storage

import (
    "github.com/jinzhu/gorm"
    "github.com/yourusername/golang-job-scheduler/internal/model"
)

// Store is an interface for saving & retrieving Job data.
type Store interface {
    SaveJob(j *model.Job) error
    GetJob(id string) (*model.Job, error)
}

// PGStore is a Postgres-backed implementation of the Store interface.
type PGStore struct {
    db *gorm.DB
}

// NewPGStore returns a new PGStore with the given Gorm DB connection.
func NewPGStore(db *gorm.DB) *PGStore {
    return &PGStore{db: db}
}

// Ensure PGStore implements the Store interface at compile time.
var _ Store = (*PGStore)(nil)

// SaveJob saves or updates the job in the DB using GORM.
func (s *PGStore) SaveJob(j *model.Job) error {
    return s.db.Save(j).Error
}

// GetJob retrieves a job by ID. Returns an error if not found or on DB failure.
func (s *PGStore) GetJob(id string) (*model.Job, error) {
    var job model.Job
    if err := s.db.Where("id = ?", id).First(&job).Error; err != nil {
        return nil, err
    }
    return &job, nil
}
