package middleware

import (
    "os"

    "github.com/gin-gonic/gin"
)

// small helper to read env with fallback
func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

var defaultJWTSecret = []byte(getEnv("HRIS_JWT_SECRET", "dev-jwt-secret"))

// AuthMiddlewareDefault uses the environment-configured JWT secret.
func AuthMiddlewareDefault() gin.HandlerFunc {
    return AuthMiddleware(defaultJWTSecret)
}

// RequireRoleDefault uses the environment-configured JWT secret.
func RequireRoleDefault(role string) gin.HandlerFunc {
    return RequireRole(role, defaultJWTSecret)
}
