package services

import (
	"context"
	"errors"
	"time"

	"github.com/ronaldpalay/hris/src/models"
)

type RecruitmentService struct {
	repo RecruitmentRepo
}

func NewRecruitmentService(r RecruitmentRepo) *RecruitmentService {
	return &RecruitmentService{repo: r}
}

var (
	ErrInvalidJob       = errors.New("invalid job")
	ErrInvalidCandidate = errors.New("invalid candidate")
)

func (s *RecruitmentService) PostJob(ctx context.Context, j *models.JobPosting) error {
	if j == nil || j.JobID == "" || j.Title == "" {
		return ErrInvalidJob
	}
	j.PostedAt = time.Now().Unix()
	return s.repo.CreateJob(ctx, j)
}

func (s *RecruitmentService) GetJob(ctx context.Context, id string) (*models.JobPosting, error) {
	return s.repo.GetJob(ctx, id)
}

func (s *RecruitmentService) UpdateJob(ctx context.Context, j *models.JobPosting, expectedVersion int) error {
	if j == nil || j.JobID == "" {
		return ErrInvalidJob
	}
	return s.repo.UpdateJob(ctx, j, expectedVersion)
}

func (s *RecruitmentService) DeleteJob(ctx context.Context, id string) error {
	return s.repo.DeleteJob(ctx, id)
}

func (s *RecruitmentService) ListJobs(ctx context.Context) ([]models.JobPosting, error) {
	return s.repo.ListJobs(ctx)
}

// Candidates
func (s *RecruitmentService) ApplyCandidate(ctx context.Context, c *models.Candidate) error {
	if c == nil || c.CandidateID == "" || c.Name == "" || c.Email == "" {
		return ErrInvalidCandidate
	}
	c.AppliedAt = time.Now().Unix()
	c.Status = "applied"
	return s.repo.CreateCandidate(ctx, c)
}

func (s *RecruitmentService) GetCandidate(ctx context.Context, id string) (*models.Candidate, error) {
	return s.repo.GetCandidate(ctx, id)
}

func (s *RecruitmentService) UpdateCandidate(ctx context.Context, c *models.Candidate, expectedVersion int) error {
	if c == nil || c.CandidateID == "" {
		return ErrInvalidCandidate
	}
	return s.repo.UpdateCandidate(ctx, c, expectedVersion)
}

func (s *RecruitmentService) DeleteCandidate(ctx context.Context, id string) error {
	return s.repo.DeleteCandidate(ctx, id)
}

func (s *RecruitmentService) ListCandidates(ctx context.Context) ([]models.Candidate, error) {
	return s.repo.ListCandidates(ctx)
}
