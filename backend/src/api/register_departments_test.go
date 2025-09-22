package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ronaldpalay/hris/src/services"
)

func TestRegisterDepartmentRoutes(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    g := r.Group("/api")
    repo := services.NewInMemoryDepartmentRepo()
    RegisterDepartmentRoutes(g, repo)

    // list
    req := httptest.NewRequest(http.MethodGet, "/api/departments", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code == http.StatusNotFound {
        t.Fatalf("departments list route not registered; got 404")
    }
}
