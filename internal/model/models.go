package model

import "time"

type Role string

const (
	RoleManager  Role = "manager"
	RoleEmployee Role = "employee"
)

type Employee struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ShiftType struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BusinessHour struct {
	ID        int64     `json:"id"`
	DayOfWeek int       `json:"day_of_week"`
	OpenTime  string    `json:"open_time"`
	CloseTime string    `json:"close_time"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Shift struct {
	ID            int64     `json:"id"`
	EmployeeID    int64     `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	ShiftTypeID   int64     `json:"shift_type_id"`
	ShiftTypeName string    `json:"shift_type_name"`
	WorkDate      string    `json:"work_date"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
