package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPayrollViewAgainstRouter(t *testing.T) {
	r := NewRouter(context.Background())
	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// create payroll for employee with gross/deductions/taxes
	rec := map[string]interface{}{"employee_id": "emp-pay-1", "gross": 1500.00, "deductions": 100.00, "taxes": 165.44, "net": 1500.00 - 100.00 - 165.44, "period": "2025-08"}
	rb, _ := json.Marshal(rec)
	resp, err := client.Post(ts.URL+"/api/payroll", "application/json", bytes.NewReader(rb))
	if err != nil {
		t.Fatalf("create payroll failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}
	var created struct {
		ID         string  `json:"id"`
		EmployeeID string  `json:"employee_id"`
		Gross      float64 `json:"gross"`
		Deductions float64 `json:"deductions"`
		Taxes      float64 `json:"taxes"`
		Net        float64 `json:"net"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created payroll: %v", err)
	}

	// employee views payroll
	vr, err := client.Get(ts.URL + "/api/payroll/employee/" + created.EmployeeID)
	if err != nil {
		t.Fatalf("view payroll failed: %v", err)
	}
	defer vr.Body.Close()
	if vr.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(vr.Body)
		t.Fatalf("expected 200 on view, got %d: %s", vr.StatusCode, string(body))
	}
	var out struct {
		Items []struct {
			Gross      float64 `json:"gross"`
			Deductions float64 `json:"deductions"`
			Taxes      float64 `json:"taxes"`
			Net        float64 `json:"net"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(vr.Body).Decode(&out); err != nil {
		t.Fatalf("decode payroll list: %v", err)
	}
	if out.Total < 1 || len(out.Items) < 1 {
		t.Fatalf("unexpected payroll list: %#v", out)
	}
	expectedNet := created.Gross - created.Deductions - created.Taxes
	if out.Items[0].Net != expectedNet {
		t.Fatalf("net mismatch: expected %v got %v", expectedNet, out.Items[0].Net)
	}
}
