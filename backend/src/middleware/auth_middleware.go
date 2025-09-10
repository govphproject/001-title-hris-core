package middleware

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// parseToken parses the JWT token string and returns the claims if valid.
func parseToken(tok string, secret []byte) (jwt.MapClaims, error) {
    p, err := jwt.ParseWithClaims(tok, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return secret, nil
    })
    if err != nil || p == nil || !p.Valid {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    claims, ok := p.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid claims")
    }
    return claims, nil
}

// AuthMiddleware validates JWT from the Authorization header using the provided secret.
func AuthMiddleware(secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.GetHeader("Authorization")
        if !strings.HasPrefix(h, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        tok := strings.TrimPrefix(h, "Bearer ")
        _, err := parseToken(tok, secret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Next()
    }
}

// RequireRole checks that the JWT contains the required role
func RequireRole(role string, secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.GetHeader("Authorization")
        if !strings.HasPrefix(h, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        tok := strings.TrimPrefix(h, "Bearer ")
        claims, err := parseToken(tok, secret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        // roles can be []interface{}, []string, or string
        switch rs := claims["roles"].(type) {
        case []interface{}:
            for _, r := range rs {
                if s, ok := r.(string); ok && s == role {
                    c.Next()
                    return
                }
            }
        case []string:
            for _, s := range rs {
                if s == role {
                    c.Next()
                    return
                }
            }
        case string:
            if rs == role {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
    }
}
