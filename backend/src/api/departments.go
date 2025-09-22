package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ronaldpalay/hris/src/services"
)

func RegisterDepartmentRoutes(rg *gin.RouterGroup, repo services.DepartmentRepo) {
	rg.GET("/departments", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		items, err := repo.List(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
	})

	rg.POST("/departments", func(c *gin.Context) {
		var in map[string]interface{}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		if in["id"] == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
			return
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

	rg.GET("/departments/:id", func(c *gin.Context) {
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

	rg.PUT("/departments/:id", func(c *gin.Context) {
		id := c.Param("id")
		var in map[string]interface{}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		doc, err := repo.Update(ctx, id, in, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db update failed"})
			return
		}
		c.JSON(http.StatusOK, doc)
	})

	rg.DELETE("/departments/:id", func(c *gin.Context) {
		id := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := repo.Delete(ctx, id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// linking endpoints
	rg.POST("/departments/:id/employees", func(c *gin.Context) {
		id := c.Param("id")
		var in struct {
			EmployeeID string `json:"employee_id"`
		}
		if err := c.BindJSON(&in); err != nil || in.EmployeeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id required"})
			return
		}
		// use service for linking logic
		// lightweight: create a temporary service using repo
		svc := services.NewDepartmentService(repo)
		doc, err := svc.AddEmployee(c.Request.Context(), id, in.EmployeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "link failed"})
			return
		}
		c.JSON(http.StatusOK, doc)
	})

	rg.DELETE("/departments/:id/employees/:emp", func(c *gin.Context) {
		id := c.Param("id")
		emp := c.Param("emp")
		svc := services.NewDepartmentService(repo)
		doc, err := svc.RemoveEmployee(c.Request.Context(), id, emp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unlink failed"})
			return
		}
		c.JSON(http.StatusOK, doc)
	})
}
