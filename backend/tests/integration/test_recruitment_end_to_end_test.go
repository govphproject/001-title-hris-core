package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestRecruitmentEndToEnd(t *testing.T) {
	ts := startRecruitServer()
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// create job
	job := map[string]string{"id": "job-e2e-1", "title": "SWE"}
	b, _ := json.Marshal(job)
	resp, err := client.Post(ts.URL+"/api/jobs", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post job failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 created, got %d: %s", resp.StatusCode, string(body))
	}

	// apply to the job
	app := map[string]string{"full_name": "Bob Applicant", "email": "b@example.com"}
	ab, _ := json.Marshal(app)
	apr, err := client.Post(ts.URL+"/api/jobs/job-e2e-1/apply", "application/json", bytes.NewReader(ab))
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	defer apr.Body.Close()
	if apr.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(apr.Body)
		t.Fatalf("expected 201 on apply, got %d: %s", apr.StatusCode, string(body))
	}

	// list applicants for job
	lr, err := client.Get(ts.URL + "/api/jobs/job-e2e-1/applicants")
	if err != nil {
		t.Fatalf("list applicants failed: %v", err)
	}
	defer lr.Body.Close()
	if lr.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(lr.Body)
		t.Fatalf("expected 200 on list, got %d: %s", lr.StatusCode, string(body))
	}
}
