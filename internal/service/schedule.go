package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cafe-scheduling-api/internal/model"
	"cafe-scheduling-api/internal/repository"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidDateRange      = errors.New("invalid date range")
	ErrInvalidShiftTime      = errors.New("invalid shift time")
	ErrShiftOutsideBusiness  = errors.New("shift is outside business hours")
	ErrScheduleConflict      = errors.New("shift overlaps an existing shift for the same employee")
	ErrBusinessHoursNotFound = errors.New("business hours are not configured for this day")
)

type ScheduleService struct {
	repo *repository.Repository
}

type GenerateScheduleRequest struct {
	WeekStart   string  `json:"week_start" binding:"required"`
	EmployeeIDs []int64 `json:"employee_ids"`
	ShiftTypeID int64   `json:"shift_type_id"`
}

type GenerateScheduleResult struct {
	WeekStart   string         `json:"week_start"`
	WeekEnd     string         `json:"week_end"`
	ShiftTypeID int64          `json:"shift_type_id"`
	Created     []model.Shift  `json:"created"`
	Skipped     []ScheduleSkip `json:"skipped"`
}

type ScheduleSkip struct {
	EmployeeID int64  `json:"employee_id"`
	WorkDate   string `json:"work_date"`
	Reason     string `json:"reason"`
}

type ShiftConflict struct {
	EmployeeID   int64         `json:"employee_id"`
	EmployeeName string        `json:"employee_name"`
	WorkDate     string        `json:"work_date"`
	Shifts       []model.Shift `json:"shifts"`
}

type EmployeeHours struct {
	EmployeeID   int64   `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	ShiftCount   int     `json:"shift_count"`
	TotalHours   float64 `json:"total_hours"`
}

type AdjustShiftRequest struct {
	EmployeeID  int64  `json:"employee_id" binding:"required"`
	ShiftTypeID int64  `json:"shift_type_id" binding:"required"`
	WorkDate    string `json:"work_date" binding:"required"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
}

func NewScheduleService(repo *repository.Repository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

func (s *ScheduleService) GenerateSchedule(ctx context.Context, req GenerateScheduleRequest) (GenerateScheduleResult, error) {
	weekStart, weekEnd, err := parseWeekRange(req.WeekStart)
	if err != nil {
		return GenerateScheduleResult{}, err
	}

	shiftType, err := s.resolveShiftType(ctx, req.ShiftTypeID)
	if err != nil {
		return GenerateScheduleResult{}, err
	}

	employees, err := s.resolveEmployees(ctx, req.EmployeeIDs)
	if err != nil {
		return GenerateScheduleResult{}, err
	}

	result := GenerateScheduleResult{
		WeekStart:   weekStart.Format("2006-01-02"),
		WeekEnd:     weekEnd.Format("2006-01-02"),
		ShiftTypeID: shiftType.ID,
		Created:     make([]model.Shift, 0),
		Skipped:     make([]ScheduleSkip, 0),
	}

	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		for day := 0; day < 7; day++ {
			date := weekStart.AddDate(0, 0, day)
			dateString := date.Format("2006-01-02")
			dayOfWeek := weekdayIndex(date)

			hours, hourErr := s.repo.GetBusinessHourByDay(ctx, dayOfWeek)
			if hourErr != nil {
				return fmt.Errorf("%w: %s", ErrBusinessHoursNotFound, dateString)
			}
			if !shiftFits(shiftType, hours) {
				return fmt.Errorf("%w: %s %s-%s is not inside %s-%s", ErrShiftOutsideBusiness, dateString, shiftType.StartTime, shiftType.EndTime, hours.OpenTime, hours.CloseTime)
			}

			for _, employee := range employees {
				existing, listErr := s.repo.ListEmployeeShiftsByDateTx(ctx, tx, employee.ID, dateString)
				if listErr != nil {
					return listErr
				}

				candidate := model.Shift{
					EmployeeID:  employee.ID,
					ShiftTypeID: shiftType.ID,
					WorkDate:    dateString,
					StartTime:   shiftType.StartTime,
					EndTime:     shiftType.EndTime,
				}
				if overlapsAny(candidate, existing) {
					result.Skipped = append(result.Skipped, ScheduleSkip{
						EmployeeID: employee.ID,
						WorkDate:   dateString,
						Reason:     ErrScheduleConflict.Error(),
					})
					continue
				}

				id, createErr := s.repo.CreateShiftTx(ctx, tx, candidate)
				if createErr != nil {
					return createErr
				}
				candidate.ID = id
				candidate.EmployeeName = employee.Name
				candidate.ShiftTypeName = shiftType.Name
				result.Created = append(result.Created, candidate)
			}
		}
		return nil
	})
	if err != nil {
		return GenerateScheduleResult{}, err
	}
	return result, nil
}

func (s *ScheduleService) ListWeek(ctx context.Context, start string) ([]model.Shift, error) {
	weekStart, weekEnd, err := parseWeekRange(start)
	if err != nil {
		return nil, err
	}
	return s.repo.ListShiftsByRange(ctx, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
}

func (s *ScheduleService) ListEmployeeWeek(ctx context.Context, employeeID int64, start string) ([]model.Shift, error) {
	weekStart, weekEnd, err := parseWeekRange(start)
	if err != nil {
		return nil, err
	}
	return s.repo.ListEmployeeShiftsByRange(ctx, employeeID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
}

func (s *ScheduleService) DetectConflicts(ctx context.Context, start string) ([]ShiftConflict, error) {
	shifts, err := s.ListWeek(ctx, start)
	if err != nil {
		return nil, err
	}
	return detectOverlaps(shifts), nil
}

func (s *ScheduleService) WeeklyHours(ctx context.Context, start string) ([]EmployeeHours, error) {
	shifts, err := s.ListWeek(ctx, start)
	if err != nil {
		return nil, err
	}
	return calculateWeeklyHours(shifts), nil
}

func (s *ScheduleService) EmployeeWeeklyHours(ctx context.Context, employeeID int64, start string) ([]EmployeeHours, error) {
	shifts, err := s.ListEmployeeWeek(ctx, employeeID, start)
	if err != nil {
		return nil, err
	}
	return calculateWeeklyHours(shifts), nil
}

func (s *ScheduleService) AdjustShift(ctx context.Context, id int64, req AdjustShiftRequest) (model.Shift, error) {
	if _, _, err := parseClockRange(req.StartTime, req.EndTime); err != nil {
		return model.Shift{}, err
	}
	if _, err := time.Parse("2006-01-02", req.WorkDate); err != nil {
		return model.Shift{}, ErrInvalidDateRange
	}

	var updated model.Shift
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := s.repo.GetShiftByIDTx(ctx, tx, id)
		if err != nil {
			return err
		}
		employee, err := s.repo.GetEmployeeByID(ctx, req.EmployeeID)
		if err != nil {
			return err
		}
		shiftType, err := s.repo.GetShiftTypeByID(ctx, req.ShiftTypeID)
		if err != nil {
			return err
		}
		hours, err := s.repo.GetBusinessHourByDay(ctx, weekdayIndex(mustParseDate(req.WorkDate)))
		if err != nil {
			return fmt.Errorf("%w: %s", ErrBusinessHoursNotFound, req.WorkDate)
		}
		if !clockRangeFits(req.StartTime, req.EndTime, hours) {
			return fmt.Errorf("%w: %s %s-%s is not inside %s-%s", ErrShiftOutsideBusiness, req.WorkDate, req.StartTime, req.EndTime, hours.OpenTime, hours.CloseTime)
		}

		existingOnDate, err := s.repo.ListEmployeeShiftsByDateTx(ctx, tx, req.EmployeeID, req.WorkDate)
		if err != nil {
			return err
		}
		candidate := model.Shift{
			ID:            id,
			EmployeeID:    req.EmployeeID,
			ShiftTypeID:   req.ShiftTypeID,
			WorkDate:      req.WorkDate,
			StartTime:     req.StartTime,
			EndTime:       req.EndTime,
			EmployeeName:  employee.Name,
			ShiftTypeName: shiftType.Name,
		}
		for _, shift := range existingOnDate {
			if shift.ID == id {
				continue
			}
			if overlaps(candidate, shift) {
				return ErrScheduleConflict
			}
		}

		if err := s.repo.UpdateShiftTx(ctx, tx, candidate); err != nil {
			return err
		}
		updated = candidate
		return nil
	})
	if err != nil {
		return model.Shift{}, err
	}
	return updated, nil
}

func (s *ScheduleService) resolveShiftType(ctx context.Context, id int64) (model.ShiftType, error) {
	if id > 0 {
		return s.repo.GetShiftTypeByID(ctx, id)
	}
	return s.repo.FirstShiftType(ctx)
}

func (s *ScheduleService) resolveEmployees(ctx context.Context, ids []int64) ([]model.Employee, error) {
	if len(ids) == 0 {
		return s.repo.ListActiveEmployees(ctx)
	}

	employees := make([]model.Employee, 0, len(ids))
	for _, id := range ids {
		employee, err := s.repo.GetEmployeeByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if !employee.Active {
			return nil, fmt.Errorf("employee %d is inactive", id)
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

func parseWeekRange(raw string) (time.Time, time.Time, error) {
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	weekStart := normalizeWeekStart(date)
	return weekStart, weekStart.AddDate(0, 0, 6), nil
}

func normalizeWeekStart(date time.Time) time.Time {
	offset := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -offset)
}

func weekdayIndex(date time.Time) int {
	return (int(date.Weekday()) + 6) % 7
}

func mustParseDate(raw string) time.Time {
	date, _ := time.Parse("2006-01-02", raw)
	return date
}

func parseClockRange(start, end string) (int, int, error) {
	startMinutes, err := parseClock(start)
	if err != nil {
		return 0, 0, err
	}
	endMinutes, err := parseClock(end)
	if err != nil {
		return 0, 0, err
	}
	if endMinutes <= startMinutes {
		return 0, 0, ErrInvalidShiftTime
	}
	return startMinutes, endMinutes, nil
}

func parseClock(raw string) (int, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return 0, ErrInvalidShiftTime
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, ErrInvalidShiftTime
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, ErrInvalidShiftTime
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 {
		return 0, ErrInvalidShiftTime
	}
	return hour*60 + minute, nil
}

func shiftFits(shiftType model.ShiftType, hours model.BusinessHour) bool {
	shiftStart, err := parseClock(shiftType.StartTime)
	if err != nil {
		return false
	}
	shiftEnd, err := parseClock(shiftType.EndTime)
	if err != nil {
		return false
	}
	open, err := parseClock(hours.OpenTime)
	if err != nil {
		return false
	}
	close, err := parseClock(hours.CloseTime)
	if err != nil {
		return false
	}
	return shiftStart >= open && shiftEnd <= close
}

func clockRangeFits(start, end string, hours model.BusinessHour) bool {
	shiftStart, startErr := parseClock(start)
	shiftEnd, endErr := parseClock(end)
	open, openErr := parseClock(hours.OpenTime)
	close, closeErr := parseClock(hours.CloseTime)
	if startErr != nil || endErr != nil || openErr != nil || closeErr != nil {
		return false
	}
	return shiftStart >= open && shiftEnd <= close && shiftEnd > shiftStart
}

func overlaps(a, b model.Shift) bool {
	aStart, aErr := parseClock(a.StartTime)
	aEnd, aErr2 := parseClock(a.EndTime)
	bStart, bErr := parseClock(b.StartTime)
	bEnd, bErr2 := parseClock(b.EndTime)
	if aErr != nil || aErr2 != nil || bErr != nil || bErr2 != nil {
		return false
	}
	return aStart < bEnd && bStart < aEnd
}

func overlapsAny(candidate model.Shift, existing []model.Shift) bool {
	for _, shift := range existing {
		if candidate.WorkDate == shift.WorkDate && candidate.EmployeeID == shift.EmployeeID && overlaps(candidate, shift) {
			return true
		}
	}
	return false
}

func detectOverlaps(shifts []model.Shift) []ShiftConflict {
	grouped := make(map[string][]model.Shift)
	for _, shift := range shifts {
		key := fmt.Sprintf("%d:%s", shift.EmployeeID, shift.WorkDate)
		grouped[key] = append(grouped[key], shift)
	}

	conflicts := make([]ShiftConflict, 0)
	for _, group := range grouped {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].StartTime < group[j].StartTime
		})
		conflictingIDs := make(map[int64]bool)
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				if overlaps(group[i], group[j]) {
					conflictingIDs[group[i].ID] = true
					conflictingIDs[group[j].ID] = true
				}
			}
		}
		if len(conflictingIDs) > 0 {
			involved := make([]model.Shift, 0)
			for _, shift := range group {
				if conflictingIDs[shift.ID] {
					involved = append(involved, shift)
				}
			}
			conflicts = append(conflicts, ShiftConflict{
				EmployeeID:   group[0].EmployeeID,
				EmployeeName: group[0].EmployeeName,
				WorkDate:     group[0].WorkDate,
				Shifts:       involved,
			})
		}
	}
	return conflicts
}

func calculateWeeklyHours(shifts []model.Shift) []EmployeeHours {
	type stats struct {
		name  string
		count int
		total float64
	}
	byEmployee := make(map[int64]*stats)
	for _, shift := range shifts {
		start, startErr := parseClock(shift.StartTime)
		end, endErr := parseClock(shift.EndTime)
		if startErr != nil || endErr != nil {
			continue
		}
		entry := byEmployee[shift.EmployeeID]
		if entry == nil {
			entry = &stats{name: shift.EmployeeName}
			byEmployee[shift.EmployeeID] = entry
		}
		entry.count++
		entry.total += float64(end-start) / 60
	}

	hours := make([]EmployeeHours, 0, len(byEmployee))
	for employeeID, entry := range byEmployee {
		hours = append(hours, EmployeeHours{
			EmployeeID:   employeeID,
			EmployeeName: entry.name,
			ShiftCount:   entry.count,
			TotalHours:   entry.total,
		})
	}
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].EmployeeID < hours[j].EmployeeID
	})
	return hours
}
