package repository

import (
	"context"
	"errors"
	"fmt"

	"cafe-scheduling-api/internal/model"

	"github.com/jackc/pgx/v5"
)

var ErrBusinessHourNotFound = errors.New("business hour not found")

func (r *Repository) UpsertBusinessHour(ctx context.Context, hour model.BusinessHour) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO business_hours (day_of_week, open_time, close_time)
		VALUES ($1, $2::time, $3::time)
		ON CONFLICT (day_of_week)
		DO UPDATE SET open_time = EXCLUDED.open_time, close_time = EXCLUDED.close_time, updated_at = now()
		RETURNING id`,
		hour.DayOfWeek, hour.OpenTime, hour.CloseTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert business hour: %w", err)
	}
	return id, nil
}

func (r *Repository) GetBusinessHourByDay(ctx context.Context, dayOfWeek int) (model.BusinessHour, error) {
	hour, err := scanBusinessHour(r.db.QueryRow(ctx, `
		SELECT id, day_of_week, to_char(open_time, 'HH24:MI'), to_char(close_time, 'HH24:MI'), updated_at
		FROM business_hours
		WHERE day_of_week = $1`, dayOfWeek))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.BusinessHour{}, ErrBusinessHourNotFound
	}
	return hour, err
}

func (r *Repository) ListBusinessHours(ctx context.Context) ([]model.BusinessHour, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, day_of_week, to_char(open_time, 'HH24:MI'), to_char(close_time, 'HH24:MI'), updated_at
		FROM business_hours
		ORDER BY day_of_week`)
	if err != nil {
		return nil, fmt.Errorf("list business hours: %w", err)
	}
	defer rows.Close()

	hours := make([]model.BusinessHour, 0)
	for rows.Next() {
		hour, err := scanBusinessHour(rows)
		if err != nil {
			return nil, err
		}
		hours = append(hours, hour)
	}
	return hours, rows.Err()
}

func scanBusinessHour(row rowScanner) (model.BusinessHour, error) {
	var hour model.BusinessHour
	err := row.Scan(
		&hour.ID,
		&hour.DayOfWeek,
		&hour.OpenTime,
		&hour.CloseTime,
		&hour.UpdatedAt,
	)
	if err != nil {
		return model.BusinessHour{}, fmt.Errorf("scan business hour: %w", err)
	}
	return hour, nil
}
