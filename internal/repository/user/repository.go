package userRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

// CreateUser - создание пользователя
func (r *Repository) CreateUser(ctx context.Context, tx *pgx.Conn, username string, TTL time.Duration) (int, error) {
	// Вставка пользователя
	var userId int
	if err := tx.QueryRow(ctx, `
		INSERT INTO public.users (username) VALUES ($1)
		RETURNING id
	`, username).Scan(&userId); err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// Вычисляем время истечения
	expirationTime := time.Now().Add(TTL)

	// Вставка времени жизни пользователя
	if _, err := tx.Exec(ctx, `
		INSERT INTO public.users_time_live (user_id, time_live) VALUES ($1, $2)
	`, userId, expirationTime); err != nil {
		return 0, fmt.Errorf("failed to create user's live time: %w", err)
	}
	return userId, nil
}

// GetExpiredUsers Получение пользователей с истекшим временем жизни.
// Возвращает срез id пользователей, у которых time_live <= NOW()
func (r *Repository) GetExpiredUsers(ctx context.Context, conn *pgx.Conn) ([]int, error) {
	rows, err := conn.Query(ctx, `
		SELECT u.id FROM public.users u
		JOIN public.users_time_live utl ON u.id = utl.user_id
		WHERE utl.time_live <= NOW()
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired users: %w", err)
	}
	defer rows.Close()

	ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return nil, fmt.Errorf("failed to collect expired user ids: %w", err)
	}
	return ids, nil
}

// DeleteUser - удаление пользователя
// Удаление производим в рамках переданного соединения/транзакции.
func (r *Repository) DeleteUser(ctx context.Context, tx *pgx.Conn, userId int) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM public.users WHERE id = $1
	`, userId); err != nil {
		return fmt.Errorf("failed to delete user id %d: %w", userId, err)
	}
	return nil
}

// GetUserByID возвращает информацию о пользователе по ID
func (r *Repository) GetUserByID(ctx context.Context, conn *pgx.Conn, userId int) (username string, err error) {
	err = conn.QueryRow(ctx, `
		SELECT username FROM public.users WHERE id = $1
	`, userId).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("failed to get user by id %d: %w", userId, err)
	}
	return username, nil
}

// CheckUserExpired проверяет, истек ли срок жизни пользователя
func (r *Repository) CheckUserExpired(ctx context.Context, conn *pgx.Conn, userId int) (bool, error) {
	var expired bool
	err := conn.QueryRow(ctx, `
		SELECT time_live <= NOW() 
		FROM public.users_time_live 
		WHERE user_id = $1
	`, userId).Scan(&expired)
	if err != nil {
		return false, fmt.Errorf("failed to check user expiration for id %d: %w", userId, err)
	}
	return expired, nil
}

// HasActiveBookings проверяет, есть ли у пользователя активные букинги
func (r *Repository) HasActiveBookings(ctx context.Context, conn *pgx.Conn, userId int) (bool, error) {
	var count int
	err := conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT 1 FROM public.laundry_bookings WHERE user_id = $1
			UNION ALL
			SELECT 1 FROM public.kitchen_bookings WHERE user_id = $1
		) as bookings
	`, userId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check bookings for user %d: %w", userId, err)
	}
	return count > 0, nil
}
