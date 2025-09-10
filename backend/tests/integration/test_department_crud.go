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

type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func startDeptServer() *httptest.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	depts := map[string]Department{}
	employeesByDept := map[string]map[string]bool{}

	// create department
	r.POST("/api/departments", func(c *gin.Context) {
		var in Department
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		id := in.ID
		if id == "" {
			id = time.Now().Format("20060102150405")
		}
		in.ID = id
		depts[id] = in
		if _, ok := employeesByDept[id]; !ok {
			employeesByDept[id] = map[string]bool{}
		}
		c.JSON(http.StatusCreated, in)
	})

	// assign employee to department
	// POST /api/departments/:id/employees with { employee_id }
	r.POST("/api/departments/:id/employees", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := depts[id]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "dept not found"})
			return
		}
		var in struct {
			EmployeeID string `json:"employee_id"`
		}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		employeesByDept[id][in.EmployeeID] = true
		c.Status(http.StatusNoContent)
	})

	// list employees in dept
	r.GET("/api/departments/:id/employees", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := depts[id]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "dept not found"})
			return
		}
		var list []string
		for eid := range employeesByDept[id] {
			list = append(list, eid)
		}
		c.JSON(http.StatusOK, gin.H{"items": list, "total": len(list)})
	})

	return httptest.NewServer(r)
}

func TestDepartmentCRUDAndLinking(t *testing.T) {
	ts := startDeptServer()
	defer ts.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	// create dept
	payload := map[string]string{"name": "Engineering"}
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
	var dept Department
	if err := json.NewDecoder(resp.Body).Decode(&dept); err != nil {
		t.Fatalf("decode dept: %v", err)
	}
	if dept.ID == "" {
		t.Fatalf("dept missing id")
	}

	// assign an employee
	empID := "emp-link-1"
	assign := map[string]string{"employee_id": empID}
	ab, _ := json.Marshal(assign)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/departments/"+dept.ID+"/employees", bytes.NewReader(ab))
	req.Header.Set("Content-Type", "application/json")
	ar, err := client.Do(req)
	if err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	defer ar.Body.Close()
	if ar.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(ar.Body)
		t.Fatalf("expected 204 on assign, got %d: %s", ar.StatusCode, string(body))
	}

	// list employees
	lr, err := client.Get(ts.URL + "/api/departments/" + dept.ID + "/employees")
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
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0] != empID {
		t.Fatalf("unexpected list result: %#v", out)
	}
}
