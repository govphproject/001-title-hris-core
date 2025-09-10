package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ronaldpalay/hris/src/middleware"
)

// Backwards-compatible wrappers used by tests and earlier code that expect
// AuthMiddleware() and RequireRole(role) without passing the secret.
// These now delegate to `src/middleware` helpers which read the default
// JWT secret from the environment.
func AuthMiddleware() gin.HandlerFunc {
	return middleware.AuthMiddlewareDefault()
}

func RequireRole(role string) gin.HandlerFunc {
	return middleware.RequireRoleDefault(role)
}
