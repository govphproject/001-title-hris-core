package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("test-jwt-secret")

func startAuthServer() *httptest.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/api/auth/login", func(c *gin.Context) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		if creds.Username != "admin" || creds.Password != "password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": creds.Username,
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		s, _ := token.SignedString(jwtSecret)
		c.JSON(http.StatusOK, gin.H{"token": s})
	})

	auth := func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if len(h) < 8 || h[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok := h[7:]
		_, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil })
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}

	r.GET("/api/employees", auth, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}, "total": 0})
	})

	return httptest.NewServer(r)
}

func TestAuthIntegration(t *testing.T) {
	ts := startAuthServer()
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
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if body.Token == "" {
		t.Fatalf("empty token returned")
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/employees", nil)
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
