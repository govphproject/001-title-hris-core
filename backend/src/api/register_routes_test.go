package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ronaldpalay/hris/src/services"
)

// ensure Register* functions wire routes (existence check: response != 404)
func TestRegisterAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api")
	authStore := services.NewInMemoryUserStore()
	RegisterAuthRoutes(g, authStore, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatalf("auth login route not registered; got 404")
	}
}

func TestRegisterEmployeeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api")
	empRepo := services.NewInMemoryEmployeeRepo()
	RegisterEmployeeRoutes(g, empRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatalf("employees route not registered; got 404")
	}
}

func TestRegisterPayrollRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api")
	pRepo := services.NewInMemoryPayrollRepo()
	RegisterPayrollRoutes(g, pRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/payroll", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatalf("payroll route not registered; got 404")
	}
}
