package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type JobPosting struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Applicant struct {
	ID        string `json:"id"`
	JobID     string `json:"job_id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Submitted int64  `json:"submitted"`
}

func startRecruitServer() *httptest.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	jobs := map[string]JobPosting{}
	applicants := map[string]Applicant{}
	applicantsByJob := map[string][]string{}

	// create job
	r.POST("/api/jobs", func(c *gin.Context) {
		var in JobPosting
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		id := in.ID
		if id == "" {
			id = time.Now().Format("20060102150405")
		}
		in.ID = id
		jobs[id] = in
		c.JSON(http.StatusCreated, in)
	})

	// apply to job
	r.POST("/api/jobs/:id/apply", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := jobs[id]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		var in struct {
			FullName, Email string `json:"full_name"`
		}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		aid := "app-" + time.Now().Format("150405")
		ap := Applicant{ID: aid, JobID: id, FullName: in.FullName, Email: in.Email, Submitted: time.Now().Unix()}
		applicants[aid] = ap
		applicantsByJob[id] = append(applicantsByJob[id], aid)
		c.JSON(http.StatusCreated, ap)
	})

	// list applicants for job
	r.GET("/api/jobs/:id/applicants", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := jobs[id]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		ids := applicantsByJob[id]
		var out []Applicant
		for _, aid := range ids {
			if a, ok := applicants[aid]; ok {
				out = append(out, a)
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	})

	return httptest.NewServer(r)
}

func TestRecruitmentFlow(t *testing.T) {
	ts := startRecruitServer()
	defer ts.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	// create job
	job := map[string]string{"title": "Software Engineer"}
	jb, _ := json.Marshal(job)
	resp, err := client.Post(ts.URL+"/api/jobs", "application/json", bytes.NewReader(jb))
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}
	var created JobPosting
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode job: %v", err)
	}

	// apply
	app := map[string]string{"full_name": "Alice Applicant", "email": "alice@example.com"}
	ab, _ := json.Marshal(app)
	apr, err := client.Post(ts.URL+"/api/jobs/"+created.ID+"/apply", "application/json", bytes.NewReader(ab))
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	defer apr.Body.Close()
	if apr.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(apr.Body)
		t.Fatalf("expected 201 on apply, got %d: %s", apr.StatusCode, string(body))
	}
	var applied Applicant
	if err := json.NewDecoder(apr.Body).Decode(&applied); err != nil {
		t.Fatalf("decode applicant: %v", err)
	}

	// list applicants
	lr, err := client.Get(ts.URL + "/api/jobs/" + created.ID + "/applicants")
	if err != nil {
		t.Fatalf("list applicants failed: %v", err)
	}
	defer lr.Body.Close()
	if lr.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(lr.Body)
		t.Fatalf("expected 200 on list, got %d: %s", lr.StatusCode, string(body))
	}
	var out struct {
		Items []Applicant `json:"items"`
		Total int         `json:"total"`
	}
	if err := json.NewDecoder(lr.Body).Decode(&out); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].Email != "alice@example.com" {
		t.Fatalf("unexpected applicants: %#v", out)
	}
}
