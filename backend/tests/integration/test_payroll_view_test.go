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

type PayrollRecord struct {
	ID         string  `json:"id"`
	EmployeeID string  `json:"employee_id"`
	Amount     float64 `json:"amount"`
	Period     string  `json:"period"`
}

func startPayrollServer() *httptest.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	records := map[string]PayrollRecord{}
	recordsByEmployee := map[string][]string{}

	// create payroll
	r.POST("/api/payroll", func(c *gin.Context) {
		var in PayrollRecord
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		id := in.ID
		if id == "" {
			id = "pay-" + time.Now().Format("150405")
		}
		in.ID = id
		records[id] = in
		recordsByEmployee[in.EmployeeID] = append(recordsByEmployee[in.EmployeeID], id)
		c.JSON(http.StatusCreated, in)
	})

	// view payroll for employee
	r.GET("/api/payroll/employee/:id", func(c *gin.Context) {
		id := c.Param("id")
		ids := recordsByEmployee[id]
		var out []PayrollRecord
		for _, rid := range ids {
			if rcd, ok := records[rid]; ok {
				out = append(out, rcd)
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	})

	return httptest.NewServer(r)
}

func TestPayrollView(t *testing.T) {
	ts := startPayrollServer()
	defer ts.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	// create payroll for employee
	rec := PayrollRecord{EmployeeID: "emp-pay-1", Amount: 1234.56, Period: "2025-08"}
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
	var created PayrollRecord
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
		Items []PayrollRecord `json:"items"`
		Total int             `json:"total"`
	}
	if err := json.NewDecoder(vr.Body).Decode(&out); err != nil {
		t.Fatalf("decode payroll list: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].Amount != created.Amount {
		t.Fatalf("unexpected payroll list: %#v", out)
	}
}
