package api

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/ronaldpalay/hris/src/services"
)

func RegisterPayrollRoutes(rg *gin.RouterGroup, repo services.PayrollRepo) {
    rg.POST("/payroll", func(c *gin.Context) {
        var in map[string]interface{}
        if err := c.BindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
            return
        }
        if in["id"] == nil || in["id"] == "" {
            in["id"] = fmt.Sprintf("pay-%d", time.Now().UnixNano())
        }
        doc := in
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        out, err := repo.Create(ctx, doc)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
            return
        }
        c.JSON(http.StatusCreated, out)
    })

    rg.GET("/payroll/employee/:id", func(c *gin.Context) {
        id := c.Param("id")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        items, err := repo.ListByEmployee(ctx, id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
    })
}
