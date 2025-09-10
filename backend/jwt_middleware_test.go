package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func genToken(t *testing.T, roles interface{}, tamperAlg bool) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": "tester", "roles": roles, "exp": time.Now().Add(1 * time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if tamperAlg {
		// change header alg to simulate unexpected signing method
		tok.Header["alg"] = "RS256"
	}
	s, err := tok.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("signed token: %v", err)
	}
	return s
}

func TestAuthMiddleware_AllowsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ok", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	token := genToken(t, []string{"admin"}, false)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestRequireRole_VariousRoleTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", RequireRole("admin"), func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		name  string
		roles interface{}
	}{
		{"slice-string", []string{"admin"}},
		{"slice-interface", []interface{}{"admin"}},
		{"single-string", "admin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			token := genToken(t, tc.roles, false)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("%s: expected 200 got %d", tc.name, w.Code)
			}
		})
	}
}

func TestRequireRole_RejectsWrongSigningMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", RequireRole("admin"), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	// craft token that is signed with HS256 but has header alg tampered to RS256 to trigger rejection
	token := genToken(t, []string{"admin"}, true)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong signing method, got %d", w.Code)
	}
}
