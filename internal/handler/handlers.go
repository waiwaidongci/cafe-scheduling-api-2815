package handler

import (
	"cafe-scheduling-api/internal/config"
	"cafe-scheduling-api/internal/repository"
	"cafe-scheduling-api/internal/service"
)

type Handlers struct {
	Auth          *AuthHandler
	Employees     *EmployeeHandler
	ShiftTypes    *ShiftTypeHandler
	BusinessHours *BusinessHourHandler
	Schedules     *ScheduleHandler
}

func NewHandlers(
	cfg config.Config,
	authService *service.AuthService,
	scheduleService *service.ScheduleService,
	repo *repository.Repository,
) *Handlers {
	return &Handlers{
		Auth:          NewAuthHandler(cfg, authService),
		Employees:     NewEmployeeHandler(repo),
		ShiftTypes:    NewShiftTypeHandler(repo),
		BusinessHours: NewBusinessHourHandler(repo),
		Schedules:     NewScheduleHandler(scheduleService),
	}
}
