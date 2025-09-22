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

	apipkg "github.com/ronaldpalay/hris/src/api"
	"github.com/ronaldpalay/hris/src/services"
)

func TestDepartmentEndToEnd(t *testing.T) {
	// build a router only with department routes (avoids importing package main)
	r := gin.Default()
	apiGroup := r.Group("/api")
	deptRepo := services.NewInMemoryDepartmentRepo()
	apipkg.RegisterDepartmentRoutes(apiGroup, deptRepo)

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// create
	payload := map[string]string{"id": "dept-e2e-1", "name": "Engineering"}
	b, _ := json.Marshal(payload)
	resp, err := client.Post(ts.URL+"/api/departments", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("create dept failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}

	// link employee
	emp := map[string]string{"employee_id": "e2e-emp-1"}
	rb, _ := json.Marshal(emp)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/departments/dept-e2e-1/employees", bytes.NewReader(rb))
	req.Header.Set("Content-Type", "application/json")
	ar, err := client.Do(req)
	if err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	defer ar.Body.Close()
	if ar.StatusCode != http.StatusOK && ar.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(ar.Body)
		t.Fatalf("expected 200 or 204 on assign, got %d: %s", ar.StatusCode, string(body))
	}

	// list employees
	lr, err := client.Get(ts.URL + "/api/departments/dept-e2e-1/employees")
	if err != nil {
		t.Fatalf("list employees failed: %v", err)
	}
	defer lr.Body.Close()
	if lr.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(lr.Body)
		t.Fatalf("expected 200 on list, got %d: %s", lr.StatusCode, string(body))
	}
	var out struct {
		Items []string `json:"items"`
		Total int      `json:"total"`
	}
	if err := json.NewDecoder(lr.Body).Decode(&out); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0] != "e2e-emp-1" {
		t.Fatalf("unexpected list result: %#v", out)
	}
}
