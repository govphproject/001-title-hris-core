package contract

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
)

var baseURL string

type Employee struct {
    EmployeeID   string                 `json:"employee_id"`
    LegalName    map[string]interface{} `json:"legal_name"`
    PreferredName string                `json:"preferred_name,omitempty"`
    Email        string                 `json:"email,omitempty"`
    HireDate     string                 `json:"hire_date,omitempty"`
    Version      int                    `json:"version,omitempty"`
}

func httpClient() *http.Client {
    return &http.Client{Timeout: 5 * time.Second}
}

func TestEmployeesContract(t *testing.T) {
    // If HRIS_TEST_EXTERNAL=1 is set, hit external server; otherwise start in-process server.
    if os.Getenv("HRIS_TEST_EXTERNAL") == "1" {
        // external server expected at localhost:8080
        baseURL = "http://localhost:8080/api"
        if err := pingHealth(); err != nil {
            t.Fatalf("backend health check failed: %v", err)
        }
    } else {
        // start in-process Gin server with the minimal mock handlers
        gin.SetMode(gin.ReleaseMode)
        router := gin.New()
        router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
        api := router.Group("/api")
        {
            // reuse simple in-memory handlers mirroring main.go
            var (
                storeMu   = make(chan struct{}, 1)
                employeesIn = map[string]Employee{}
            )
            // simple mutex abstraction using channel
            lock := func() { storeMu <- struct{}{} }
            unlock := func() { <-storeMu }

            // list
            api.GET("/employees", func(c *gin.Context) {
                lock()
                items := make([]Employee, 0, len(employeesIn))
                for _, e := range employeesIn {
                    items = append(items, e)
                }
                unlock()
                c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
            })
            // create
            api.POST("/employees", func(c *gin.Context) {
                var in map[string]interface{}
                if err := c.BindJSON(&in); err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
                    return
                }
                legal, _ := in["legal_name"].(map[string]interface{})
                hireDate, _ := in["hire_date"].(string)
                email, _ := in["email"].(string)
                id := fmt.Sprintf("emp-%d", time.Now().UnixNano())
                emp := Employee{EmployeeID: id, LegalName: legal, Email: email, HireDate: hireDate, Version: 1}
                lock()
                employeesIn[id] = emp
                unlock()
                c.JSON(http.StatusCreated, emp)
            })
            api.GET("/employees/:id", func(c *gin.Context) {
                id := c.Param("id")
                lock()
                e, ok := employeesIn[id]
                unlock()
                if !ok {
                    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                    return
                }
                c.JSON(http.StatusOK, e)
            })
            api.PUT("/employees/:id", func(c *gin.Context) {
                id := c.Param("id")
                var in map[string]interface{}
                if err := c.BindJSON(&in); err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
                    return
                }
                lock()
                e, ok := employeesIn[id]
                if !ok {
                    unlock()
                    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                    return
                }
                if v, ok := in["version"].(float64); ok {
                    if int(v) != e.Version {
                        unlock()
                        c.JSON(http.StatusConflict, gin.H{"error": "version mismatch"})
                        return
                    }
                }
                if pn, ok := in["preferred_name"].(string); ok {
                    e.PreferredName = pn
                }
                e.Version += 1
                employeesIn[id] = e
                unlock()
                c.JSON(http.StatusOK, e)
            })
            api.DELETE("/employees/:id", func(c *gin.Context) {
                id := c.Param("id")
                lock()
                if _, ok := employeesIn[id]; !ok {
                    unlock()
                    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                    return
                }
                delete(employeesIn, id)
                unlock()
                c.Status(http.StatusNoContent)
            })
        }
        ts := httptest.NewServer(router)
        defer ts.Close()
        baseURL = ts.URL + "/api"
    }

    // 1) GET /employees
    t.Run("GET /employees", func(t *testing.T) {
        resp, err := httpClient().Get(baseURL + "/employees")
        if err != nil {
            t.Fatalf("GET /employees request failed: %v", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            t.Fatalf("expected 200 OK, got %d: %s", resp.StatusCode, string(body))
        }
        // basic response shape check
        var body struct{
            Items []interface{} `json:"items"`
            Total int `json:"total"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
            t.Fatalf("failed to decode GET /employees response: %v", err)
        }
    })

    // 2) POST /employees -> expect 201 and returned employee
    var created Employee
    t.Run("POST /employees", func(t *testing.T) {
        payload := map[string]interface{}{
            "legal_name": map[string]string{"first":"Contract","last":"Tester"},
            "email": fmt.Sprintf("contract-%d@example.com", time.Now().Unix()),
            "hire_date": "2025-09-01",
        }
        b, _ := json.Marshal(payload)
        resp, err := httpClient().Post(baseURL+"/employees", "application/json", bytes.NewReader(b))
        if err != nil {
            t.Fatalf("POST /employees request failed: %v", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusCreated {
            body, _ := io.ReadAll(resp.Body)
            t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
        }
        if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
            t.Fatalf("failed to decode created employee: %v", err)
        }
        if created.EmployeeID == "" {
            t.Fatalf("created employee missing employee_id")
        }
    })

    // 3) GET /employees/{id}
    t.Run("GET /employees/{id}", func(t *testing.T) {
        if created.EmployeeID == "" {
            t.Skip("create step failed; skipping GET by id")
        }
        resp, err := httpClient().Get(baseURL + "/employees/" + created.EmployeeID)
        if err != nil {
            t.Fatalf("GET /employees/{id} request failed: %v", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            t.Fatalf("expected 200 OK, got %d: %s", resp.StatusCode, string(body))
        }
        var fetched Employee
        if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
            t.Fatalf("failed to decode employee by id: %v", err)
        }
        if fetched.EmployeeID != created.EmployeeID {
            t.Fatalf("fetched employee_id mismatch: want %s got %s", created.EmployeeID, fetched.EmployeeID)
        }
    })

    // 4) PUT /employees/{id} (update) - expect 200
    t.Run("PUT /employees/{id}", func(t *testing.T) {
        if created.EmployeeID == "" {
            t.Skip("create step failed; skipping update")
        }
        update := map[string]interface{}{
            "preferred_name": "Updated Contract",
            "version": created.Version,
        }
        b, _ := json.Marshal(update)
        req, _ := http.NewRequest(http.MethodPut, baseURL+"/employees/"+created.EmployeeID, bytes.NewReader(b))
        req.Header.Set("Content-Type", "application/json")
        resp, err := httpClient().Do(req)
        if err != nil {
            t.Fatalf("PUT /employees/{id} request failed: %v", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            t.Fatalf("expected 200 OK on update, got %d: %s", resp.StatusCode, string(body))
        }
    })

    // 5) DELETE /employees/{id} - expect 204
    t.Run("DELETE /employees/{id}", func(t *testing.T) {
        if created.EmployeeID == "" {
            t.Skip("create step failed; skipping delete")
        }
        req, _ := http.NewRequest(http.MethodDelete, baseURL+"/employees/"+created.EmployeeID, nil)
        resp, err := httpClient().Do(req)
        if err != nil {
            t.Fatalf("DELETE /employees/{id} request failed: %v", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusNoContent {
            body, _ := io.ReadAll(resp.Body)
            t.Fatalf("expected 204 No Content on delete, got %d: %s", resp.StatusCode, string(body))
        }
    })
}

func pingHealth() error {
    resp, err := httpClient().Get("http://localhost:8080/health")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("health check returned %d: %s", resp.StatusCode, string(b))
    }
    return nil
}
