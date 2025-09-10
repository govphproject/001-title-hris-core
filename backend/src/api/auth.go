package api

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/ronaldpalay/hris/src/services"
)

// RegisterAuthRoutes registers auth routes on the given router group.
func RegisterAuthRoutes(rg *gin.RouterGroup, authStore services.AuthStore, jwtSecret []byte) {
    rg.POST("/auth/login", func(c *gin.Context) {
        var creds struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        if err := c.BindJSON(&creds); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
            return
        }
        ok, roles, err := authStore.ValidateCredentials(context.Background(), creds.Username, creds.Password)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "auth error"})
            return
        }
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
            return
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "sub":   creds.Username,
            "roles": roles,
            "exp":   time.Now().Add(1 * time.Hour).Unix(),
        })
        s, err := token.SignedString(jwtSecret)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"token": s})
    })

    rg.POST("/users", func(c *gin.Context) {
        var in struct {
            Username string   `json:"username"`
            Password string   `json:"password"`
            Roles    []string `json:"roles"`
        }
        if err := c.BindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
            return
        }
        if in.Username == "" || in.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
            return
        }
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := authStore.CreateUser(ctx, in.Username, in.Password, in.Roles); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
            return
        }
        c.JSON(http.StatusCreated, gin.H{"username": in.Username, "roles": in.Roles})
    })
}

