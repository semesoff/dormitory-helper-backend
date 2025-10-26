package laundryService

import (
	"context"
	laundryRepository "dormitory-helper-service/internal/repository/laundry"
	utilsService "dormitory-helper-service/internal/service/utils"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LaundryRepository interface {
	CreateLaundryBooking(ctx context.Context, conn *pgx.Conn, userID int, startTime, endTime time.Time) (int, error)
	GetLaundryBookings(ctx context.Context, conn *pgx.Conn, startTime, endTime *time.Time) ([]laundryRepository.Booking, error)
	GetUserLaundryBookings(ctx context.Context, conn *pgx.Conn, userID int) ([]laundryRepository.Booking, error)
	DeleteLaundryBooking(ctx context.Context, conn *pgx.Conn, bookingID, userID int) error

	CreateKitchenBooking(ctx context.Context, conn *pgx.Conn, userID int, startTime, endTime time.Time) (int, error)
	GetKitchenBookings(ctx context.Context, conn *pgx.Conn, startTime, endTime *time.Time) ([]laundryRepository.Booking, error)
	GetUserKitchenBookings(ctx context.Context, conn *pgx.Conn, userID int) ([]laundryRepository.Booking, error)
	DeleteKitchenBooking(ctx context.Context, conn *pgx.Conn, bookingID, userID int) error
}

type Service struct {
	repo LaundryRepository
	db   *pgxpool.Pool
}

func NewService(repo LaundryRepository, db *pgxpool.Pool) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

// CreateLaundryBooking создает запись на стирку
func (s *Service) CreateLaundryBooking(ctx context.Context, userID int, startTime, endTime time.Time) (int, error) {
	// Проверка длительности (максимум 2 часа)
	if endTime.Sub(startTime) > 2*time.Hour {
		return 0, fmt.Errorf("laundry booking duration cannot exceed 2 hours")
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return 0, fmt.Errorf("end time must be after start time")
	}

	var bookingID int
	err := utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		var err error
		bookingID, err = s.repo.CreateLaundryBooking(ctx, conn.Conn(), userID, startTime, endTime)
		return err
	})

	return bookingID, err
}

// GetLaundryBookings получает все записи на стирку
func (s *Service) GetLaundryBookings(ctx context.Context, startTime, endTime *time.Time) ([]laundryRepository.Booking, error) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	bookings, err := s.repo.GetLaundryBookings(ctx, conn.Conn(), startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get laundry bookings: %w", err)
	}

	return bookings, nil
}

// GetUserLaundryBookings получает все записи пользователя на стирку
func (s *Service) GetUserLaundryBookings(ctx context.Context, userID int) ([]laundryRepository.Booking, error) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	bookings, err := s.repo.GetUserLaundryBookings(ctx, conn.Conn(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user laundry bookings: %w", err)
	}

	return bookings, nil
}

// DeleteLaundryBooking удаляет запись на стирку
func (s *Service) DeleteLaundryBooking(ctx context.Context, bookingID, userID int) error {
	return utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		return s.repo.DeleteLaundryBooking(ctx, conn.Conn(), bookingID, userID)
	})
}

// CreateKitchenBooking создает запись на кухню
func (s *Service) CreateKitchenBooking(ctx context.Context, userID int, startTime, endTime time.Time) (int, error) {
	// Проверка длительности (максимум 3 часа)
	if endTime.Sub(startTime) > 3*time.Hour {
		return 0, fmt.Errorf("kitchen booking duration cannot exceed 3 hours")
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return 0, fmt.Errorf("end time must be after start time")
	}

	var bookingID int
	err := utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		var err error
		bookingID, err = s.repo.CreateKitchenBooking(ctx, conn.Conn(), userID, startTime, endTime)
		return err
	})

	return bookingID, err
}

// GetKitchenBookings получает все записи на кухню
func (s *Service) GetKitchenBookings(ctx context.Context, startTime, endTime *time.Time) ([]laundryRepository.Booking, error) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	bookings, err := s.repo.GetKitchenBookings(ctx, conn.Conn(), startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get kitchen bookings: %w", err)
	}

	return bookings, nil
}

// GetUserKitchenBookings получает все записи пользователя на кухню
func (s *Service) GetUserKitchenBookings(ctx context.Context, userID int) ([]laundryRepository.Booking, error) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	bookings, err := s.repo.GetUserKitchenBookings(ctx, conn.Conn(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user kitchen bookings: %w", err)
	}

	return bookings, nil
}

// DeleteKitchenBooking удаляет запись на кухню
func (s *Service) DeleteKitchenBooking(ctx context.Context, bookingID, userID int) error {
	return utilsService.WithTx(ctx, s.db, func(conn *pgxpool.Conn) error {
		return s.repo.DeleteKitchenBooking(ctx, conn.Conn(), bookingID, userID)
	})
}
