package repository

import (
	"context"
	"errors"
	"fmt"

	"cafe-scheduling-api/internal/model"

	"github.com/jackc/pgx/v5"
)

var ErrShiftNotFound = errors.New("shift not found")

const shiftSelect = `
	SELECT s.id, s.employee_id, e.name, s.shift_type_id, st.name,
	       to_char(s.work_date, 'YYYY-MM-DD'),
	       to_char(s.start_time, 'HH24:MI'),
	       to_char(s.end_time, 'HH24:MI'),
	       s.created_at, s.updated_at
	FROM shifts s
	JOIN employees e ON e.id = s.employee_id
	JOIN shift_types st ON st.id = s.shift_type_id`

func (r *Repository) CreateShiftTx(ctx context.Context, db DBTX, shift model.Shift) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, `
		INSERT INTO shifts (employee_id, shift_type_id, work_date, start_time, end_time)
		VALUES ($1, $2, $3::date, $4::time, $5::time)
		RETURNING id`,
		shift.EmployeeID, shift.ShiftTypeID, shift.WorkDate, shift.StartTime, shift.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create shift: %w", err)
	}
	return id, nil
}

func (r *Repository) UpdateShiftTx(ctx context.Context, db DBTX, shift model.Shift) error {
	commandTag, err := db.Exec(ctx, `
		UPDATE shifts
		SET employee_id = $1, shift_type_id = $2, work_date = $3::date,
		    start_time = $4::time, end_time = $5::time, updated_at = now()
		WHERE id = $6`,
		shift.EmployeeID, shift.ShiftTypeID, shift.WorkDate, shift.StartTime, shift.EndTime, shift.ID,
	)
	if err != nil {
		return fmt.Errorf("update shift: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrShiftNotFound
	}
	return nil
}

func (r *Repository) GetShiftByID(ctx context.Context, id int64) (model.Shift, error) {
	return getShift(ctx, r.db, shiftSelect+` WHERE s.id = $1`, id)
}

func (r *Repository) GetShiftByIDTx(ctx context.Context, db DBTX, id int64) (model.Shift, error) {
	return getShift(ctx, db, shiftSelect+` WHERE s.id = $1`, id)
}

func (r *Repository) ListShiftsByRange(ctx context.Context, start, end string) ([]model.Shift, error) {
	return listShifts(ctx, r.db, shiftSelect+` WHERE s.work_date BETWEEN $1::date AND $2::date ORDER BY s.work_date, s.start_time, e.name`, start, end)
}

func (r *Repository) ListEmployeeShiftsByRange(ctx context.Context, employeeID int64, start, end string) ([]model.Shift, error) {
	return listShifts(ctx, r.db, shiftSelect+` WHERE s.employee_id = $1 AND s.work_date BETWEEN $2::date AND $3::date ORDER BY s.work_date, s.start_time`, employeeID, start, end)
}

func (r *Repository) ListEmployeeShiftsByDateTx(ctx context.Context, db DBTX, employeeID int64, workDate string) ([]model.Shift, error) {
	return listShifts(ctx, db, shiftSelect+` WHERE s.employee_id = $1 AND s.work_date = $2::date ORDER BY s.start_time`, employeeID, workDate)
}

func getShift(ctx context.Context, db DBTX, query string, args ...any) (model.Shift, error) {
	shift, err := scanShift(db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Shift{}, ErrShiftNotFound
	}
	return shift, err
}

func listShifts(ctx context.Context, db DBTX, query string, args ...any) ([]model.Shift, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list shifts: %w", err)
	}
	defer rows.Close()

	shifts := make([]model.Shift, 0)
	for rows.Next() {
		shift, err := scanShift(rows)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}
	return shifts, rows.Err()
}

func scanShift(row rowScanner) (model.Shift, error) {
	var shift model.Shift
	err := row.Scan(
		&shift.ID,
		&shift.EmployeeID,
		&shift.EmployeeName,
		&shift.ShiftTypeID,
		&shift.ShiftTypeName,
		&shift.WorkDate,
		&shift.StartTime,
		&shift.EndTime,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)
	if err != nil {
		return model.Shift{}, fmt.Errorf("scan shift: %w", err)
	}
	return shift, nil
}
