package handler

import (
	"time"

	"cafe-scheduling-api/internal/model"
	"cafe-scheduling-api/internal/pkg/response"
	"cafe-scheduling-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type ShiftTypeHandler struct {
	repo *repository.Repository
}

func NewShiftTypeHandler(repo *repository.Repository) *ShiftTypeHandler {
	return &ShiftTypeHandler{repo: repo}
}

type CreateShiftTypeRequest struct {
	Name      string `json:"name" binding:"required"`
	Color     string `json:"color"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

func (h *ShiftTypeHandler) Create(c *gin.Context) {
	var req CreateShiftTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid shift type data")
		return
	}
	if !validClock(req.StartTime) || !validClock(req.EndTime) {
		response.BadRequest(c, "times must use HH:MM format")
		return
	}

	shiftType := model.ShiftType{
		Name:      req.Name,
		Color:     req.Color,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	id, err := h.repo.CreateShiftType(c.Request.Context(), shiftType)
	if err != nil {
		response.InternalError(c, "failed to create shift type")
		return
	}
	shiftType.ID = id
	response.Created(c, shiftType)
}

func (h *ShiftTypeHandler) List(c *gin.Context) {
	shiftTypes, err := h.repo.ListShiftTypes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list shift types")
		return
	}
	response.OK(c, shiftTypes)
}

func validClock(raw string) bool {
	if len(raw) != 5 || raw[2] != ':' {
		return false
	}
	_, err := time.Parse("15:04", raw)
	return err == nil
}
