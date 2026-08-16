package repository

import (
	"context"
	"errors"
	"fmt"

	"cafe-scheduling-api/internal/model"

	"github.com/jackc/pgx/v5"
)

var ErrEmployeeNotFound = errors.New("employee not found")

func (r *Repository) CreateEmployee(ctx context.Context, employee model.Employee) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO employees (username, name, role, password_hash, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		employee.Username, employee.Name, employee.Role, employee.PasswordHash, employee.Active,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create employee: %w", err)
	}
	return id, nil
}

func (r *Repository) GetEmployeeByUsername(ctx context.Context, username string) (model.Employee, error) {
	return getEmployee(ctx, r.db, `SELECT id, username, name, role, password_hash, active, created_at, updated_at
		FROM employees WHERE username = $1`, username)
}

func (r *Repository) GetEmployeeByID(ctx context.Context, id int64) (model.Employee, error) {
	return getEmployee(ctx, r.db, `SELECT id, username, name, role, password_hash, active, created_at, updated_at
		FROM employees WHERE id = $1`, id)
}

func (r *Repository) ListActiveEmployees(ctx context.Context) ([]model.Employee, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, username, name, role, password_hash, active, created_at, updated_at
		FROM employees
		WHERE active = true
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list active employees: %w", err)
	}
	defer rows.Close()

	var employees []model.Employee
	for rows.Next() {
		employee, err := scanEmployee(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	return employees, rows.Err()
}

func (r *Repository) ListEmployees(ctx context.Context, limit, offset int) ([]model.Employee, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, username, name, role, password_hash, active, created_at, updated_at
		FROM employees
		ORDER BY id
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	employees := make([]model.Employee, 0)
	for rows.Next() {
		employee, err := scanEmployee(rows)
		if err != nil {
			return nil, 0, err
		}
		employees = append(employees, employee)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM employees`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count employees: %w", err)
	}
	return employees, total, nil
}

func getEmployee(ctx context.Context, db DBTX, query string, args ...any) (model.Employee, error) {
	employee, err := scanEmployee(db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Employee{}, ErrEmployeeNotFound
	}
	return employee, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEmployee(row rowScanner) (model.Employee, error) {
	var employee model.Employee
	err := row.Scan(
		&employee.ID,
		&employee.Username,
		&employee.Name,
		&employee.Role,
		&employee.PasswordHash,
		&employee.Active,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)
	if err != nil {
		return model.Employee{}, fmt.Errorf("scan employee: %w", err)
	}
	return employee, nil
}
