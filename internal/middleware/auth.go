package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"cafe-scheduling-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextEmployeeKey = "employee"
	ContextTokenKey    = "token"
)

type EmployeeClaims struct {
	UserID   int64      `json:"user_id"`
	Username string     `json:"username"`
	Role     model.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, ttl time.Duration, employee model.Employee) (string, error) {
	now := time.Now()
	claims := EmployeeClaims{
		UserID:   employee.ID,
		Username: employee.Username,
		Role:     employee.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   employee.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Authenticate(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		claims := &EmployeeClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextTokenKey, token)
		c.Set(ContextEmployeeKey, model.Employee{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		})
		c.Next()
	}
}

func RequireManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		employee, ok := CurrentEmployee(c)
		if !ok || employee.Role != model.RoleManager {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "manager role required"})
			return
		}
		c.Next()
	}
}

func CurrentEmployee(c *gin.Context) (model.Employee, bool) {
	value, exists := c.Get(ContextEmployeeKey)
	if !exists {
		return model.Employee{}, false
	}
	employee, ok := value.(model.Employee)
	return employee, ok
}

func CurrentEmployeeFromContext(ctx context.Context) (model.Employee, bool) {
	value := ctx.Value(ContextEmployeeKey)
	if value == nil {
		return model.Employee{}, false
	}
	employee, ok := value.(model.Employee)
	return employee, ok
}

func extractToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
