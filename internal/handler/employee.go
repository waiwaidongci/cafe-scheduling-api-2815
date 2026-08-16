package handler

import (
	"cafe-scheduling-api/internal/model"
	"cafe-scheduling-api/internal/pkg/pagination"
	"cafe-scheduling-api/internal/pkg/response"
	"cafe-scheduling-api/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type EmployeeHandler struct {
	repo *repository.Repository
}

func NewEmployeeHandler(repo *repository.Repository) *EmployeeHandler {
	return &EmployeeHandler{repo: repo}
}

type CreateEmployeeRequest struct {
	Username string     `json:"username" binding:"required"`
	Name     string     `json:"name" binding:"required"`
	Password string     `json:"password" binding:"required,min=6"`
	Role     model.Role `json:"role" binding:"required,oneof=manager employee"`
	Active   *bool      `json:"active"`
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid employee data")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	employee := model.Employee{
		Username:     req.Username,
		Name:         req.Name,
		Role:         req.Role,
		PasswordHash: string(hash),
		Active:       active,
	}

	id, err := h.repo.CreateEmployee(c.Request.Context(), employee)
	if err != nil {
		response.InternalError(c, "failed to create employee")
		return
	}
	employee.ID = id
	employee.PasswordHash = ""
	response.Created(c, employee)
}

func (h *EmployeeHandler) List(c *gin.Context) {
	params := pagination.Parse(c)
	employees, total, err := h.repo.ListEmployees(c.Request.Context(), params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		response.InternalError(c, "failed to list employees")
		return
	}
	response.OK(c, pagination.Page[model.Employee]{
		Items:    employees,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	})
}
