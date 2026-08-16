package handler

import (
	"errors"
	"strconv"

	"cafe-scheduling-api/internal/pkg/response"
	"cafe-scheduling-api/internal/repository"
	"cafe-scheduling-api/internal/service"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	service *service.ScheduleService
}

func NewScheduleHandler(service *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) Generate(c *gin.Context) {
	var req service.GenerateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "week_start is required")
		return
	}
	result, err := h.service.GenerateSchedule(c.Request.Context(), req)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.Created(c, result)
}

func (h *ScheduleHandler) Week(c *gin.Context) {
	start := c.Query("start")
	if start == "" {
		response.BadRequest(c, "start date is required")
		return
	}
	shifts, err := h.service.ListWeek(c.Request.Context(), start)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, shifts)
}

func (h *ScheduleHandler) Conflicts(c *gin.Context) {
	start := c.Query("start")
	if start == "" {
		response.BadRequest(c, "start date is required")
		return
	}
	conflicts, err := h.service.DetectConflicts(c.Request.Context(), start)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, conflicts)
}

func (h *ScheduleHandler) WeeklyHours(c *gin.Context) {
	start := c.Query("start")
	if start == "" {
		response.BadRequest(c, "start date is required")
		return
	}
	stats, err := h.service.WeeklyHours(c.Request.Context(), start)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, stats)
}

func (h *ScheduleHandler) Adjust(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid shift id")
		return
	}
	var req service.AdjustShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid shift data")
		return
	}
	shift, err := h.service.AdjustShift(c.Request.Context(), id, req)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, shift)
}

func (h *ScheduleHandler) MyWeek(c *gin.Context) {
	employeeID := currentEmployeeID(c)
	start := c.Query("start")
	if start == "" {
		response.BadRequest(c, "start date is required")
		return
	}
	shifts, err := h.service.ListEmployeeWeek(c.Request.Context(), employeeID, start)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, shifts)
}

func (h *ScheduleHandler) MyWeeklyHours(c *gin.Context) {
	employeeID := currentEmployeeID(c)
	start := c.Query("start")
	if start == "" {
		response.BadRequest(c, "start date is required")
		return
	}
	stats, err := h.service.EmployeeWeeklyHours(c.Request.Context(), employeeID, start)
	if err != nil {
		writeScheduleError(c, err)
		return
	}
	response.OK(c, stats)
}

func currentEmployeeID(c *gin.Context) int64 {
	employee, _ := currentEmployee(c)
	return employee.ID
}

func writeScheduleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidDateRange):
		response.BadRequest(c, "invalid date range")
	case errors.Is(err, service.ErrInvalidShiftTime):
		response.BadRequest(c, "invalid shift time")
	case errors.Is(err, service.ErrShiftOutsideBusiness):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrBusinessHoursNotFound):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrScheduleConflict):
		response.JSON(c, 409, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrShiftNotFound):
		response.NotFound(c, "shift not found")
	default:
		response.InternalError(c, "schedule operation failed")
	}
}
