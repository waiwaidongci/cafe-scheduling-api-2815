package router

import (
	"net/http"

	"cafe-scheduling-api/internal/config"
	"cafe-scheduling-api/internal/handler"
	"cafe-scheduling-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, handlers *handler.Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.POST("/auth/login", handlers.Auth.Login)

	authed := api.Group("")
	authed.Use(middleware.Authenticate(cfg.JWTSecret))

	authed.GET("/my/schedules", handlers.Schedules.MyWeek)
	authed.GET("/my/statistics/hours", handlers.Schedules.MyWeeklyHours)

	manager := authed.Group("")
	manager.Use(middleware.RequireManager())
	manager.POST("/employees", handlers.Employees.Create)
	manager.GET("/employees", handlers.Employees.List)
	manager.POST("/shift-types", handlers.ShiftTypes.Create)
	manager.GET("/shift-types", handlers.ShiftTypes.List)
	manager.PUT("/business-hours", handlers.BusinessHours.Upsert)
	manager.GET("/business-hours", handlers.BusinessHours.List)
	manager.POST("/schedules/generate", handlers.Schedules.Generate)
	manager.GET("/schedules/week", handlers.Schedules.Week)
	manager.GET("/schedules/conflicts", handlers.Schedules.Conflicts)
	manager.PUT("/schedules/:id", handlers.Schedules.Adjust)
	manager.GET("/statistics/hours", handlers.Schedules.WeeklyHours)

	return r
}
