package services

import (
	"context"
	"testing"

	"github.com/ronaldpalay/hris/src/models"
)

func TestRecruitmentService_PostJobAndApply(t *testing.T) {
	repo := NewInMemoryRecruitmentRepo()
	svc := NewRecruitmentService(repo)
	ctx := context.Background()

	job := &models.JobPosting{JobID: "job-1", Title: "Engineer"}
	if err := svc.PostJob(ctx, job); err != nil {
		t.Fatalf("post job failed: %v", err)
	}

	got, err := svc.GetJob(ctx, "job-1")
	if err != nil {
		t.Fatalf("get job failed: %v", err)
	}
	if got.Title != "Engineer" || got.JobID != "job-1" {
		t.Fatalf("unexpected job: %#v", got)
	}

	// apply candidate
	cand := &models.Candidate{CandidateID: "cand-1", Name: "Alice", Email: "a@example.com"}
	if err := svc.ApplyCandidate(ctx, cand); err != nil {
		t.Fatalf("apply candidate failed: %v", err)
	}
	gc, err := svc.GetCandidate(ctx, "cand-1")
	if err != nil {
		t.Fatalf("get candidate failed: %v", err)
	}
	if gc.Name != "Alice" || gc.CandidateID != "cand-1" {
		t.Fatalf("unexpected candidate: %#v", gc)
	}
}

func TestRecruitmentService_EdgeCases(t *testing.T) {
	repo := NewInMemoryRecruitmentRepo()
	svc := NewRecruitmentService(repo)
	ctx := context.Background()

	// invalid job
	if err := svc.PostJob(ctx, &models.JobPosting{}); err == nil {
		t.Fatalf("expected invalid job error")
	}

	// duplicate job
	job := &models.JobPosting{JobID: "job-dup", Title: "Dev"}
	if err := svc.PostJob(ctx, job); err != nil {
		t.Fatalf("post job failed: %v", err)
	}
	if err := svc.PostJob(ctx, job); err == nil {
		t.Fatalf("expected duplicate job error")
	}

	// invalid candidate
	if err := svc.ApplyCandidate(ctx, &models.Candidate{}); err == nil {
		t.Fatalf("expected invalid candidate error")
	}
}
