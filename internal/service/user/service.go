package userService

import (
	"context"
	utilsService "dormitory-helper-service/internal/service/utils"
	jwtUtils "dormitory-helper-service/internal/utils/jwt"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, tx *pgx.Conn, username string, TTL time.Duration) (int, error)
	GetExpiredUsers(ctx context.Context, conn *pgx.Conn) ([]int, error)
	DeleteUser(ctx context.Context, tx *pgx.Conn, userId int) error
	GetUserByID(ctx context.Context, conn *pgx.Conn, userId int) (username string, err error)
	CheckUserExpired(ctx context.Context, conn *pgx.Conn, userId int) (bool, error)
	HasActiveBookings(ctx context.Context, conn *pgx.Conn, userId int) (bool, error)
}

type Service struct {
	repo      UserRepository
	db        *pgxpool.Pool
	jwtSecret []byte
}

func NewService(repo UserRepository, db *pgxpool.Pool, jwtSecret []byte) *Service {
	return &Service{
		repo:      repo,
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// CreateUser создает нового пользователя с указанным временем жизни
func (s *Service) CreateUser(ctx context.Context, username string, TTL time.Duration) (int, string, error) {
	var userId int
	var token string
	err := utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		var err error
		userId, err = s.repo.CreateUser(ctx, conn.Conn(), username, TTL)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Генерация JWT-токена
		token, err = jwtUtils.GenerateToken(userId, username, s.jwtSecret)
		if err != nil {
			return fmt.Errorf("failed to generate token: %w", err)
		}

		return nil
	})
	return userId, token, err
}

// CheckAuthentication проверяет аутентификацию пользователя
// Логика:
// 1. Если токен пустой - создаем нового пользователя
// 2. Если токен есть и валидный:
//   - Если время жизни истекло И нет букингов - создаем нового пользователя
//   - Если есть букинги - возвращаем существующего пользователя с текущим токеном
func (s *Service) CheckAuthentication(ctx context.Context, token string) (userId int, username string, resultToken string, err error) {
	// Если токен пустой - создаем нового пользователя
	if token == "" {
		return s.createNewUser(ctx)
	}

	// Валидация токена
	claims, err := jwtUtils.ValidateToken(token, s.jwtSecret)
	if err != nil {
		// Токен невалидный - создаем нового пользователя
		return s.createNewUser(ctx)
	}

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Проверяем, истек ли срок жизни пользователя
	expired, err := s.repo.CheckUserExpired(ctx, conn.Conn(), claims.UserID)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to check user expiration: %w", err)
	}

	// Если время жизни не истекло - возвращаем текущего пользователя
	if !expired {
		return claims.UserID, claims.Username, token, nil
	}

	// Время жизни истекло - проверяем букинги
	hasBookings, err := s.repo.HasActiveBookings(ctx, conn.Conn(), claims.UserID)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to check bookings: %w", err)
	}

	// Если есть букинги - возвращаем текущего пользователя с текущим токеном
	if hasBookings {
		return claims.UserID, claims.Username, token, nil
	}

	// Нет букингов - удаляем старого и создаем нового пользователя
	if err := s.DeleteUser(ctx, claims.UserID); err != nil {
		return 0, "", "", fmt.Errorf("failed to delete expired user: %w", err)
	}

	return s.createNewUser(ctx)
}

// createNewUser создает нового пользователя с автоматически сгенерированным именем
func (s *Service) createNewUser(ctx context.Context) (userId int, username string, token string, err error) {
	username = fmt.Sprintf("user_%d", time.Now().UnixNano())
	TTL := 7 * 24 * time.Hour // 7 дней

	userId, token, err = s.CreateUser(ctx, username, TTL)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to create new user: %w", err)
	}

	return userId, username, token, nil
}

// GetExpiredUsers возвращает список ID пользователей с истекшим временем жизни
func (s *Service) GetExpiredUsers(ctx context.Context) ([]int, error) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	ids, err := s.repo.GetExpiredUsers(ctx, conn.Conn())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired users: %w", err)
	}

	return ids, nil
}

// DeleteUser удаляет пользователя по ID
func (s *Service) DeleteUser(ctx context.Context, userId int) error {
	return utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		if err := s.repo.DeleteUser(ctx, conn.Conn(), userId); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		return nil
	})
}

// CleanupExpiredUsers удаляет всех пользователей с истекшим временем жизни.
// Возвращает количество удаленных пользователей
func (s *Service) CleanupExpiredUsers(ctx context.Context) (int, error) {
	expiredUsers, err := s.GetExpiredUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired users: %w", err)
	}

	deletedCount := 0
	for _, userId := range expiredUsers {
		if err := s.DeleteUser(ctx, userId); err != nil {
			// Логируем ошибку, но продолжаем удаление остальных
			fmt.Printf("failed to delete user %d: %v\n", userId, err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}
