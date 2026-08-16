package handler

import (
	"cafe-scheduling-api/internal/model"
	"cafe-scheduling-api/internal/pkg/response"
	"cafe-scheduling-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type BusinessHourHandler struct {
	repo *repository.Repository
}

func NewBusinessHourHandler(repo *repository.Repository) *BusinessHourHandler {
	return &BusinessHourHandler{repo: repo}
}

type UpsertBusinessHoursRequest struct {
	Hours []model.BusinessHour `json:"hours" binding:"required,dive"`
}

func (h *BusinessHourHandler) Upsert(c *gin.Context) {
	var req UpsertBusinessHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid business hours data")
		return
	}

	saved := make([]model.BusinessHour, 0, len(req.Hours))
	for _, hour := range req.Hours {
		if hour.DayOfWeek < 0 || hour.DayOfWeek > 6 {
			response.BadRequest(c, "day_of_week must be between 0 and 6")
			return
		}
		if !validClock(hour.OpenTime) || !validClock(hour.CloseTime) {
			response.BadRequest(c, "times must use HH:MM format")
			return
		}
		id, err := h.repo.UpsertBusinessHour(c.Request.Context(), hour)
		if err != nil {
			response.InternalError(c, "failed to upsert business hours")
			return
		}
		hour.ID = id
		saved = append(saved, hour)
	}
	response.OK(c, saved)
}

func (h *BusinessHourHandler) List(c *gin.Context) {
	hours, err := h.repo.ListBusinessHours(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list business hours")
		return
	}
	response.OK(c, hours)
}
