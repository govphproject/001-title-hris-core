package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestLoginAndProtectedEndpoint(t *testing.T) {
    r := NewRouter(context.Background())
    ts := httptest.NewServer(r)
    defer ts.Close()

    creds := map[string]string{"username": "admin", "password": "password"}
    b, _ := json.Marshal(creds)
    resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(b))
    if err != nil {
        t.Fatalf("login request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("login status: %d", resp.StatusCode)
    }
    var body struct{ Token string `json:"token"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        t.Fatalf("failed to decode login response: %v", err)
    }
    if body.Token == "" {
        t.Fatalf("empty token returned")
    }

    req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/secure/employees", nil)
    req.Header.Set("Authorization", "Bearer "+body.Token)
    client := &http.Client{Timeout: 5 * time.Second}
    r2, err := client.Do(req)
    if err != nil {
        t.Fatalf("protected request failed: %v", err)
    }
    defer r2.Body.Close()
    if r2.StatusCode != http.StatusOK {
        t.Fatalf("expected 200 OK, got %d", r2.StatusCode)
    }
}

func TestLoginFailure(t *testing.T) {
    r := NewRouter(context.Background())
    ts := httptest.NewServer(r)
    defer ts.Close()

    creds := map[string]string{"username": "admin", "password": "wrong"}
    b, _ := json.Marshal(creds)
    resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(b))
    if err != nil {
        t.Fatalf("login request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusUnauthorized {
        t.Fatalf("expected 401 Unauthorized for bad creds, got %d", resp.StatusCode)
    }
}

func TestAdminAccess(t *testing.T) {
    r := NewRouter(context.Background())
    ts := httptest.NewServer(r)
    defer ts.Close()

    creds := map[string]string{"username": "admin", "password": "password"}
    b, _ := json.Marshal(creds)
    resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(b))
    if err != nil {
        t.Fatalf("login request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("login status: %d", resp.StatusCode)
    }
    var body struct{ Token string `json:"token"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        t.Fatalf("failed to decode login response: %v", err)
    }

    req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/secure/admin", nil)
    req.Header.Set("Authorization", "Bearer "+body.Token)
    client := &http.Client{Timeout: 5 * time.Second}
    r2, err := client.Do(req)
    if err != nil {
        t.Fatalf("admin request failed: %v", err)
    }
    defer r2.Body.Close()
    if r2.StatusCode != http.StatusOK {
        t.Fatalf("expected 200 OK for admin, got %d", r2.StatusCode)
    }
}

func TestCreateUserAndLogin(t *testing.T) {
    r := NewRouter(context.Background())
    ts := httptest.NewServer(r)
    defer ts.Close()

    // login as admin
    creds := map[string]string{"username": "admin", "password": "password"}
    b, _ := json.Marshal(creds)
    resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(b))
    if err != nil {
        t.Fatalf("login request failed: %v", err)
    }
    defer resp.Body.Close()
    var body struct{ Token string `json:"token"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        t.Fatalf("failed to decode login response: %v", err)
    }

    // create user
    newUser := map[string]interface{}{"username": "alice", "password": "password123", "roles": []string{"user"}}
    nb, _ := json.Marshal(newUser)
    req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/secure/users", bytes.NewReader(nb))
    req.Header.Set("Authorization", "Bearer "+body.Token)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 5 * time.Second}
    r2, err := client.Do(req)
    if err != nil {
        t.Fatalf("create user request failed: %v", err)
    }
    defer r2.Body.Close()
    if r2.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201 Created, got %d", r2.StatusCode)
    }

    // login as alice
    creds2 := map[string]string{"username": "alice", "password": "password123"}
    b2, _ := json.Marshal(creds2)
    resp2, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(b2))
    if err != nil {
        t.Fatalf("alice login request failed: %v", err)
    }
    defer resp2.Body.Close()
    if resp2.StatusCode != http.StatusOK {
        t.Fatalf("alice login status: %d", resp2.StatusCode)
    }
}
