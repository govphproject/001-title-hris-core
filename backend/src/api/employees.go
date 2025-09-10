package api

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/ronaldpalay/hris/src/services"
)

func RegisterEmployeeRoutes(rg *gin.RouterGroup, repo services.EmployeeRepo) {
    rg.GET("/employees", func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        items, err := repo.List(ctx)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
    })

    rg.POST("/employees", func(c *gin.Context) {
        var in map[string]interface{}
        if err := c.BindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
            return
        }
        if in["employee_id"] == nil {
            in["employee_id"] = fmt.Sprintf("emp-%d", time.Now().UnixNano())
        }
        if in["version"] == nil {
            in["version"] = 1
        }
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        doc, err := repo.Create(ctx, in)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
            return
        }
        c.JSON(http.StatusCreated, doc)
    })

    rg.GET("/employees/:id", func(c *gin.Context) {
        id := c.Param("id")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        doc, err := repo.Get(ctx, id)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
            return
        }
        c.JSON(http.StatusOK, doc)
    })

    rg.PUT("/employees/:id", func(c *gin.Context) {
        id := c.Param("id")
        var in map[string]interface{}
        if err := c.BindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
            return
        }
        var expected *int
        if v, ok := in["version"].(float64); ok {
            vv := int(v)
            expected = &vv
        }
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        doc, err := repo.Update(ctx, id, in, expected)
        if err != nil {
            if err.Error() == "version mismatch" {
                c.JSON(http.StatusConflict, gin.H{"error": "version mismatch"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db update failed"})
            return
        }
        c.JSON(http.StatusOK, doc)
    })

    rg.DELETE("/employees/:id", func(c *gin.Context) {
        id := c.Param("id")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := repo.Delete(ctx, id); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
            return
        }
        c.Status(http.StatusNoContent)
    })
}
