package repository

import (
	"context"
	"errors"
	"fmt"

	"cafe-scheduling-api/internal/model"

	"github.com/jackc/pgx/v5"
)

var ErrShiftTypeNotFound = errors.New("shift type not found")

func (r *Repository) CreateShiftType(ctx context.Context, shiftType model.ShiftType) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO shift_types (name, color, start_time, end_time)
		VALUES ($1, $2, $3::time, $4::time)
		RETURNING id`,
		shiftType.Name, shiftType.Color, shiftType.StartTime, shiftType.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create shift type: %w", err)
	}
	return id, nil
}

func (r *Repository) GetShiftTypeByID(ctx context.Context, id int64) (model.ShiftType, error) {
	return getShiftType(ctx, r.db, `SELECT id, name, color, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at, updated_at
		FROM shift_types WHERE id = $1`, id)
}

func (r *Repository) ListShiftTypes(ctx context.Context) ([]model.ShiftType, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, color, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at, updated_at
		FROM shift_types
		ORDER BY start_time, id`)
	if err != nil {
		return nil, fmt.Errorf("list shift types: %w", err)
	}
	defer rows.Close()

	shiftTypes := make([]model.ShiftType, 0)
	for rows.Next() {
		shiftType, err := scanShiftType(rows)
		if err != nil {
			return nil, err
		}
		shiftTypes = append(shiftTypes, shiftType)
	}
	return shiftTypes, rows.Err()
}

func (r *Repository) FirstShiftType(ctx context.Context) (model.ShiftType, error) {
	return getShiftType(ctx, r.db, `SELECT id, name, color, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at, updated_at
		FROM shift_types
		ORDER BY start_time, id
		LIMIT 1`)
}

func getShiftType(ctx context.Context, db DBTX, query string, args ...any) (model.ShiftType, error) {
	shiftType, err := scanShiftType(db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ShiftType{}, ErrShiftTypeNotFound
	}
	return shiftType, err
}

func scanShiftType(row rowScanner) (model.ShiftType, error) {
	var shiftType model.ShiftType
	err := row.Scan(
		&shiftType.ID,
		&shiftType.Name,
		&shiftType.Color,
		&shiftType.StartTime,
		&shiftType.EndTime,
		&shiftType.CreatedAt,
		&shiftType.UpdatedAt,
	)
	if err != nil {
		return model.ShiftType{}, fmt.Errorf("scan shift type: %w", err)
	}
	return shiftType, nil
}
