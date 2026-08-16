package handler

import (
	"errors"

	"cafe-scheduling-api/internal/config"
	"cafe-scheduling-api/internal/middleware"
	"cafe-scheduling-api/internal/pkg/response"
	"cafe-scheduling-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg         config.Config
	authService *service.AuthService
}

func NewAuthHandler(cfg config.Config, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "username and password are required")
		return
	}

	employee, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid username or password")
			return
		}
		response.InternalError(c, "login failed")
		return
	}

	token, err := middleware.GenerateToken(h.cfg.JWTSecret, h.cfg.TokenTTL, employee)
	if err != nil {
		response.InternalError(c, "failed to create token")
		return
	}

	response.OK(c, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"employee":   employee,
	})
}
