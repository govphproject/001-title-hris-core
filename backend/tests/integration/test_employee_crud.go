package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type Employee struct {
	EmployeeID    string                 `json:"employee_id"`
	LegalName     map[string]interface{} `json:"legal_name"`
	PreferredName string                 `json:"preferred_name,omitempty"`
	Email         string                 `json:"email,omitempty"`
	HireDate      string                 `json:"hire_date,omitempty"`
	Version       int                    `json:"version,omitempty"`
}

type AuditEntry struct {
	Action     string `json:"action"`
	EmployeeID string `json:"employee_id"`
	When       int64  `json:"when"`
}

func startEmployeeServer() *httptest.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	store := map[string]Employee{}
	audits := []AuditEntry{}

	// create
	r.POST("/api/employees", func(c *gin.Context) {
		var in map[string]interface{}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		legal, _ := in["legal_name"].(map[string]interface{})
		hireDate, _ := in["hire_date"].(string)
		email, _ := in["email"].(string)
		id := fmt.Sprintf("emp-%d", time.Now().UnixNano())
		e := Employee{EmployeeID: id, LegalName: legal, Email: email, HireDate: hireDate, Version: 1}
		store[id] = e
		audits = append(audits, AuditEntry{Action: "create", EmployeeID: id, When: time.Now().Unix()})
		c.JSON(http.StatusCreated, e)
	})

	// get
	r.GET("/api/employees/:id", func(c *gin.Context) {
		id := c.Param("id")
		e, ok := store[id]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, e)
	})

	// update
	r.PUT("/api/employees/:id", func(c *gin.Context) {
		id := c.Param("id")
		var in map[string]interface{}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		cur, ok := store[id]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if v, ok := in["version"].(float64); ok {
			if int(v) != cur.Version {
				c.JSON(http.StatusConflict, gin.H{"error": "version mismatch"})
				return
			}
		}
		if pn, ok := in["preferred_name"].(string); ok {
			cur.PreferredName = pn
		}
		cur.Version += 1
		store[id] = cur
		audits = append(audits, AuditEntry{Action: "update", EmployeeID: id, When: time.Now().Unix()})
		c.JSON(http.StatusOK, cur)
	})

	// delete
	r.DELETE("/api/employees/:id", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := store[id]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		delete(store, id)
		audits = append(audits, AuditEntry{Action: "delete", EmployeeID: id, When: time.Now().Unix()})
		c.Status(http.StatusNoContent)
	})

	// audit query endpoint for test verification
	r.GET("/api/_audit/:id", func(c *gin.Context) {
		id := c.Param("id")
		var out []AuditEntry
		for _, a := range audits {
			if a.EmployeeID == id {
				out = append(out, a)
			}
		}
		c.JSON(http.StatusOK, gin.H{"audit": out})
	})

	return httptest.NewServer(r)
}

func TestEmployeeCRUD(t *testing.T) {
	ts := startEmployeeServer()
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// 1) create
	payload := map[string]interface{}{
		"legal_name": map[string]string{"first": "Jane", "last": "Doe"},
		"email":      "jane.doe@example.com",
		"hire_date":  "2025-09-10",
	}
	b, _ := json.Marshal(payload)
	resp, err := client.Post(ts.URL+"/api/employees", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /employees failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}
	var created Employee
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created employee: %v", err)
	}
	if created.EmployeeID == "" {
		t.Fatalf("created missing employee_id")
	}

	// 2) get by id
	r2, err := client.Get(ts.URL + "/api/employees/" + created.EmployeeID)
	if err != nil {
		t.Fatalf("GET by id failed: %v", err)
	}
	defer r2.Body.Close()
	if r2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(r2.Body)
		t.Fatalf("expected 200 OK, got %d: %s", r2.StatusCode, string(body))
	}
	var fetched Employee
	if err := json.NewDecoder(r2.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode fetched: %v", err)
	}
	if fetched.EmployeeID != created.EmployeeID {
		t.Fatalf("fetched id mismatch")
	}

	// 3) update preferred_name with version check
	update := map[string]interface{}{"preferred_name": "Janie", "version": created.Version}
	ub, _ := json.Marshal(update)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/employees/"+created.EmployeeID, bytes.NewReader(ub))
	req.Header.Set("Content-Type", "application/json")
	upResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT failed: %v", err)
	}
	defer upResp.Body.Close()
	if upResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(upResp.Body)
		t.Fatalf("expected 200 OK on update, got %d: %s", upResp.StatusCode, string(body))
	}
	var updated Employee
	if err := json.NewDecoder(upResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.Version != created.Version+1 {
		t.Fatalf("version not incremented: got %d want %d", updated.Version, created.Version+1)
	}
	if updated.PreferredName != "Janie" {
		t.Fatalf("preferred name not updated: %s", updated.PreferredName)
	}

	// 4) delete
	delReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/employees/"+created.EmployeeID, nil)
	delResp, err := client.Do(delReq)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(delResp.Body)
		t.Fatalf("expected 204 No Content on delete, got %d: %s", delResp.StatusCode, string(body))
	}

	// 5) GET should now 404
	g3, err := client.Get(ts.URL + "/api/employees/" + created.EmployeeID)
	if err != nil {
		t.Fatalf("GET after delete failed: %v", err)
	}
	defer g3.Body.Close()
	if g3.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(g3.Body)
		t.Fatalf("expected 404 after delete, got %d: %s", g3.StatusCode, string(body))
	}

	// 6) verify audit entries: should have create, update, delete
	ar, err := client.Get(ts.URL + "/api/_audit/" + created.EmployeeID)
	if err != nil {
		t.Fatalf("audit query failed: %v", err)
	}
	defer ar.Body.Close()
	if ar.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(ar.Body)
		t.Fatalf("expected 200 on audit, got %d: %s", ar.StatusCode, string(body))
	}
	var ab struct {
		Audit []AuditEntry `json:"audit"`
	}
	if err := json.NewDecoder(ar.Body).Decode(&ab); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	actions := map[string]bool{}
	for _, a := range ab.Audit {
		actions[a.Action] = true
	}
	if !actions["create"] || !actions["update"] || !actions["delete"] {
		t.Fatalf("audit missing expected actions: %+v", actions)
	}
}
