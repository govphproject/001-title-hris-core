package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ronaldpalay/hris/src/models"
	"github.com/ronaldpalay/hris/src/services"
)

func RegisterRecruitmentRoutes(rg *gin.RouterGroup, repo services.RecruitmentRepo) {
	svc := services.NewRecruitmentService(repo)

	rg.GET("/jobs", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		items, err := svc.ListJobs(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
	})

	rg.POST("/jobs", func(c *gin.Context) {
		var in models.JobPosting
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := svc.PostJob(ctx, &in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, in)
	})

	// applicants
	rg.GET("/applicants", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		items, err := svc.ListCandidates(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
	})

	rg.POST("/applicants", func(c *gin.Context) {
		var in models.Candidate
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := svc.ApplyCandidate(ctx, &in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, in)
	})
}
