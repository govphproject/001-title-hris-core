package services

import (
	"context"
	"errors"
	"sync"

	"github.com/ronaldpalay/hris/src/models"
)

// RecruitmentRepo stores job postings and candidates.
type RecruitmentRepo interface {
	// Jobs
	CreateJob(ctx context.Context, j *models.JobPosting) error
	GetJob(ctx context.Context, id string) (*models.JobPosting, error)
	UpdateJob(ctx context.Context, j *models.JobPosting, expectedVersion int) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context) ([]models.JobPosting, error)

	// Candidates
	CreateCandidate(ctx context.Context, c *models.Candidate) error
	GetCandidate(ctx context.Context, id string) (*models.Candidate, error)
	UpdateCandidate(ctx context.Context, c *models.Candidate, expectedVersion int) error
	DeleteCandidate(ctx context.Context, id string) error
	ListCandidates(ctx context.Context) ([]models.Candidate, error)
}

// InMemoryRecruitmentRepo is a thread-safe in-memory repo for tests and dev.
type InMemoryRecruitmentRepo struct {
	mu         sync.RWMutex
	jobs       map[string]models.JobPosting
	candidates map[string]models.Candidate
}

func NewInMemoryRecruitmentRepo() *InMemoryRecruitmentRepo {
	return &InMemoryRecruitmentRepo{
		jobs:       make(map[string]models.JobPosting),
		candidates: make(map[string]models.Candidate),
	}
}

var (
	ErrNotFound        = errors.New("not found")
	ErrVersionMismatch = errors.New("version mismatch")
)

func (r *InMemoryRecruitmentRepo) CreateJob(ctx context.Context, j *models.JobPosting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if j == nil || j.JobID == "" {
		return errors.New("invalid job")
	}
	if _, ok := r.jobs[j.JobID]; ok {
		return errors.New("already exists")
	}
	j.Version = 1
	r.jobs[j.JobID] = *j
	return nil
}

func (r *InMemoryRecruitmentRepo) GetJob(ctx context.Context, id string) (*models.JobPosting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.jobs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &j, nil
}

func (r *InMemoryRecruitmentRepo) UpdateJob(ctx context.Context, j *models.JobPosting, expectedVersion int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.jobs[j.JobID]
	if !ok {
		return ErrNotFound
	}
	if expectedVersion != 0 && cur.Version != expectedVersion {
		return ErrVersionMismatch
	}
	j.Version = cur.Version + 1
	r.jobs[j.JobID] = *j
	return nil
}

func (r *InMemoryRecruitmentRepo) DeleteJob(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[id]; !ok {
		return ErrNotFound
	}
	delete(r.jobs, id)
	return nil
}

func (r *InMemoryRecruitmentRepo) ListJobs(ctx context.Context) ([]models.JobPosting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.JobPosting, 0, len(r.jobs))
	for _, v := range r.jobs {
		out = append(out, v)
	}
	return out, nil
}

// Candidates
func (r *InMemoryRecruitmentRepo) CreateCandidate(ctx context.Context, c *models.Candidate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c == nil || c.CandidateID == "" {
		return errors.New("invalid candidate")
	}
	if _, ok := r.candidates[c.CandidateID]; ok {
		return errors.New("already exists")
	}
	c.Version = 1
	r.candidates[c.CandidateID] = *c
	return nil
}

func (r *InMemoryRecruitmentRepo) GetCandidate(ctx context.Context, id string) (*models.Candidate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.candidates[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &c, nil
}

func (r *InMemoryRecruitmentRepo) UpdateCandidate(ctx context.Context, c *models.Candidate, expectedVersion int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.candidates[c.CandidateID]
	if !ok {
		return ErrNotFound
	}
	if expectedVersion != 0 && cur.Version != expectedVersion {
		return ErrVersionMismatch
	}
	c.Version = cur.Version + 1
	r.candidates[c.CandidateID] = *c
	return nil
}

func (r *InMemoryRecruitmentRepo) DeleteCandidate(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.candidates[id]; !ok {
		return ErrNotFound
	}
	delete(r.candidates, id)
	return nil
}

func (r *InMemoryRecruitmentRepo) ListCandidates(ctx context.Context) ([]models.Candidate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.Candidate, 0, len(r.candidates))
	for _, v := range r.candidates {
		out = append(out, v)
	}
	return out, nil
}
