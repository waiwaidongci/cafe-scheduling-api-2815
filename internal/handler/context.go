package handler

import (
	"cafe-scheduling-api/internal/middleware"
	"cafe-scheduling-api/internal/model"

	"github.com/gin-gonic/gin"
)

type ginContext = gin.Context

func currentEmployee(c *gin.Context) (model.Employee, bool) {
	return middleware.CurrentEmployee(c)
}
