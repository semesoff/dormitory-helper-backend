package laundryRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

type Booking struct {
	ID        int
	UserID    int
	StartTime time.Time
	EndTime   time.Time
}

// CreateLaundryBooking создает запись на стирку
func (r *Repository) CreateLaundryBooking(ctx context.Context, conn *pgx.Conn, userID int, startTime, endTime time.Time) (int, error) {
	// Проверка на пересечение с существующими записями
	var count int
	err := conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM laundry_bookings
		WHERE (start_time, end_time) OVERLAPS ($1, $2)
	`, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to check booking overlap: %w", err)
	}
	if count > 0 {
		return 0, fmt.Errorf("time slot is already booked")
	}

	// Создание записи
	var bookingID int
	err = conn.QueryRow(ctx, `
		INSERT INTO laundry_bookings (user_id, start_time, end_time)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, startTime, endTime).Scan(&bookingID)
	if err != nil {
		return 0, fmt.Errorf("failed to create laundry booking: %w", err)
	}

	return bookingID, nil
}

// GetLaundryBookings получает все записи на стирку в заданном диапазоне времени
func (r *Repository) GetLaundryBookings(ctx context.Context, conn *pgx.Conn, startTime, endTime *time.Time) ([]Booking, error) {
	query := `SELECT id, user_id, start_time, end_time FROM laundry_bookings`
	args := []interface{}{}

	if startTime != nil && endTime != nil {
		query += ` WHERE start_time >= $1 AND end_time <= $2`
		args = append(args, *startTime, *endTime)
	} else if startTime != nil {
		query += ` WHERE start_time >= $1`
		args = append(args, *startTime)
	} else if endTime != nil {
		query += ` WHERE end_time <= $1`
		args = append(args, *endTime)
	}

	query += ` ORDER BY start_time`

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query laundry bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.StartTime, &b.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan laundry booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

// GetUserLaundryBookings получает все записи пользователя на стирку
func (r *Repository) GetUserLaundryBookings(ctx context.Context, conn *pgx.Conn, userID int) ([]Booking, error) {
	rows, err := conn.Query(ctx, `
		SELECT id, user_id, start_time, end_time 
		FROM laundry_bookings
		WHERE user_id = $1
		ORDER BY start_time
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user laundry bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.StartTime, &b.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan user laundry booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

// DeleteLaundryBooking удаляет запись на стирку
func (r *Repository) DeleteLaundryBooking(ctx context.Context, conn *pgx.Conn, bookingID, userID int) error {
	result, err := conn.Exec(ctx, `
		DELETE FROM laundry_bookings
		WHERE id = $1 AND user_id = $2
	`, bookingID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete laundry booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking not found or user is not the owner")
	}

	return nil
}

// CreateKitchenBooking создает запись на кухню
func (r *Repository) CreateKitchenBooking(ctx context.Context, conn *pgx.Conn, userID int, startTime, endTime time.Time) (int, error) {
	// Проверка на пересечение с существующими записями
	var count int
	err := conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM kitchen_bookings
		WHERE (start_time, end_time) OVERLAPS ($1, $2)
	`, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to check booking overlap: %w", err)
	}
	if count > 0 {
		return 0, fmt.Errorf("time slot is already booked")
	}

	// Создание записи
	var bookingID int
	err = conn.QueryRow(ctx, `
		INSERT INTO kitchen_bookings (user_id, start_time, end_time)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, startTime, endTime).Scan(&bookingID)
	if err != nil {
		return 0, fmt.Errorf("failed to create kitchen booking: %w", err)
	}

	return bookingID, nil
}

// GetKitchenBookings получает все записи на кухню в заданном диапазоне времени
func (r *Repository) GetKitchenBookings(ctx context.Context, conn *pgx.Conn, startTime, endTime *time.Time) ([]Booking, error) {
	query := `SELECT id, user_id, start_time, end_time FROM kitchen_bookings`
	args := []interface{}{}

	if startTime != nil && endTime != nil {
		query += ` WHERE start_time >= $1 AND end_time <= $2`
		args = append(args, *startTime, *endTime)
	} else if startTime != nil {
		query += ` WHERE start_time >= $1`
		args = append(args, *startTime)
	} else if endTime != nil {
		query += ` WHERE end_time <= $1`
		args = append(args, *endTime)
	}

	query += ` ORDER BY start_time`

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query kitchen bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.StartTime, &b.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan kitchen booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

// GetUserKitchenBookings получает все записи пользователя на кухню
func (r *Repository) GetUserKitchenBookings(ctx context.Context, conn *pgx.Conn, userID int) ([]Booking, error) {
	rows, err := conn.Query(ctx, `
		SELECT id, user_id, start_time, end_time 
		FROM kitchen_bookings
		WHERE user_id = $1
		ORDER BY start_time
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user kitchen bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.StartTime, &b.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan user kitchen booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

// DeleteKitchenBooking удаляет запись на кухню
func (r *Repository) DeleteKitchenBooking(ctx context.Context, conn *pgx.Conn, bookingID, userID int) error {
	result, err := conn.Exec(ctx, `
		DELETE FROM kitchen_bookings
		WHERE id = $1 AND user_id = $2
	`, bookingID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete kitchen booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking not found or user is not the owner")
	}

	return nil
}
