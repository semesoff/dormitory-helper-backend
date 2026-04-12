package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dormitory-helper-backend/internal/models"
	"dormitory-helper-backend/internal/storage"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
)

type Services struct {
	repo   *storage.Repository
	tokens *TokenManager
}

func New(repo *storage.Repository, tokens *TokenManager) *Services {
	return &Services{repo: repo, tokens: tokens}
}

func (s *Services) Login(ctx context.Context, username, password string) (models.User, string, error) {
	user, passwordHash, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return models.User{}, "", err
	}
	if HashPassword(password) != passwordHash {
		return models.User{}, "", ErrUnauthorized
	}
	token, err := s.tokens.Generate(user)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

func (s *Services) EnsureStudentSession(ctx context.Context, deviceID string) (models.User, string, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return models.User{}, "", fmt.Errorf("%w: device_id is required", ErrBadRequest)
	}

	user, err := s.repo.GetUserByDeviceID(ctx, deviceID)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return models.User{}, "", err
		}
		username := "student-" + deviceID
		if len(username) > 64 {
			username = username[:64]
		}
		user, err = s.repo.CreateGuestUser(ctx, username, deviceID)
		if err != nil {
			return models.User{}, "", err
		}
	}

	token, err := s.tokens.Generate(user)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

func (s *Services) Me(ctx context.Context, authUser AuthUser) (models.User, error) {
	return s.repo.GetUserByID(ctx, authUser.ID)
}

func (s *Services) ParseToken(token string) (AuthUser, error) {
	return s.tokens.Parse(token)
}

func (s *Services) ListBookings(ctx context.Context, resourceType string, authUser AuthUser, mine bool) ([]models.Booking, error) {
	if resourceType != models.ResourceLaundry && resourceType != models.ResourceKitchen {
		return nil, fmt.Errorf("%w: invalid resource_type", ErrBadRequest)
	}
	if mine {
		return s.repo.ListBookings(ctx, resourceType, &authUser.ID)
	}
	return s.repo.ListBookings(ctx, resourceType, nil)
}

func (s *Services) CreateBooking(ctx context.Context, authUser AuthUser, resourceType string, startAt, endAt time.Time) (models.Booking, error) {
	if resourceType != models.ResourceLaundry && resourceType != models.ResourceKitchen {
		return models.Booking{}, fmt.Errorf("%w: invalid resource_type", ErrBadRequest)
	}
	if !startAt.Before(endAt) {
		return models.Booking{}, fmt.Errorf("%w: start_at must be before end_at", ErrBadRequest)
	}
	duration := endAt.Sub(startAt)
	maxDuration := 2 * time.Hour
	if resourceType == models.ResourceKitchen {
		maxDuration = 3 * time.Hour
	}
	if duration > maxDuration {
		return models.Booking{}, fmt.Errorf("%w: booking duration exceeds limit", ErrBadRequest)
	}
	conflict, err := s.repo.HasBookingConflict(ctx, resourceType, startAt, endAt)
	if err != nil {
		return models.Booking{}, err
	}
	if conflict {
		return models.Booking{}, fmt.Errorf("%w: booking slot conflicts with existing booking", ErrBadRequest)
	}
	return s.repo.CreateBooking(ctx, models.Booking{
		UserID:       authUser.ID,
		ResourceType: resourceType,
		StartAt:      startAt,
		EndAt:        endAt,
	})
}

func (s *Services) DeleteBooking(ctx context.Context, authUser AuthUser, id int64) error {
	ownerID, err := s.repo.GetBookingOwner(ctx, id)
	if err != nil {
		return err
	}
	if authUser.Role != models.RoleAdmin && ownerID != authUser.ID {
		return ErrForbidden
	}
	return s.repo.DeleteBooking(ctx, id)
}

func (s *Services) ListRepairs(ctx context.Context, authUser AuthUser, mine bool) ([]models.Repair, error) {
	if mine {
		return s.repo.ListRepairs(ctx, &authUser.ID)
	}
	return s.repo.ListRepairs(ctx, nil)
}

func (s *Services) CreateRepair(ctx context.Context, authUser AuthUser, location, category, description string) (models.Repair, error) {
	if strings.TrimSpace(location) == "" || strings.TrimSpace(category) == "" || strings.TrimSpace(description) == "" {
		return models.Repair{}, fmt.Errorf("%w: location, category and description are required", ErrBadRequest)
	}
	return s.repo.CreateRepair(ctx, models.Repair{
		UserID:      authUser.ID,
		Location:    location,
		Category:    category,
		Description: description,
		Status:      models.RepairPending,
	})
}

func (s *Services) DeleteRepair(ctx context.Context, authUser AuthUser, id int64) error {
	repair, err := s.repo.GetRepairByID(ctx, id)
	if err != nil {
		return err
	}
	if authUser.Role != models.RoleAdmin && repair.UserID != authUser.ID {
		return ErrForbidden
	}
	return s.repo.DeleteRepair(ctx, id)
}

func (s *Services) UpdateRepairStatus(ctx context.Context, authUser AuthUser, id int64, status string) (models.Repair, error) {
	if authUser.Role != models.RoleAdmin {
		return models.Repair{}, ErrForbidden
	}
	switch status {
	case models.RepairPending, models.RepairInProgress, models.RepairCompleted, models.RepairRejected:
	default:
		return models.Repair{}, fmt.Errorf("%w: invalid repair status", ErrBadRequest)
	}
	repair, err := s.repo.UpdateRepairStatus(ctx, id, status)
	if err != nil {
		return models.Repair{}, err
	}
	settings, err := s.repo.NotificationSettings(ctx, repair.UserID)
	if err == nil && settings.RepairsEnabled {
		_ = s.repo.CreateNotification(ctx, repair.UserID, "repair", "Статус заявки обновлён", fmt.Sprintf("Заявка для %s переведена в статус %s", repair.Location, status), fmt.Sprintf("repair-status:%d:%s", repair.ID, status))
	}
	return repair, nil
}

func (s *Services) ListExchangeItems(ctx context.Context) ([]models.ExchangeItem, error) {
	return s.repo.ListExchangeItems(ctx)
}

func (s *Services) CreateExchangeItem(ctx context.Context, authUser AuthUser, title, category, itemType, description, contact string) (models.ExchangeItem, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(category) == "" || strings.TrimSpace(itemType) == "" || strings.TrimSpace(description) == "" {
		return models.ExchangeItem{}, fmt.Errorf("%w: title, category, type and description are required", ErrBadRequest)
	}
	switch itemType {
	case "give", "exchange", "sell":
	default:
		return models.ExchangeItem{}, fmt.Errorf("%w: invalid exchange type", ErrBadRequest)
	}
	return s.repo.CreateExchangeItem(ctx, models.ExchangeItem{
		UserID:      authUser.ID,
		Title:       title,
		Category:    category,
		Type:        itemType,
		Description: description,
		Contact:     contact,
	})
}

func (s *Services) DeleteExchangeItem(ctx context.Context, authUser AuthUser, id int64) error {
	ownerID, err := s.repo.GetExchangeOwner(ctx, id)
	if err != nil {
		return err
	}
	if authUser.Role != models.RoleAdmin && ownerID != authUser.ID {
		return ErrForbidden
	}
	return s.repo.DeleteExchangeItem(ctx, id)
}

func (s *Services) ListAnnouncements(ctx context.Context) ([]models.Announcement, error) {
	return s.repo.ListAnnouncements(ctx)
}

func (s *Services) CreateAnnouncement(ctx context.Context, authUser AuthUser, title, content, priority string) (models.Announcement, error) {
	if authUser.Role != models.RoleAdmin {
		return models.Announcement{}, ErrForbidden
	}
	switch priority {
	case "high", "medium", "low":
	default:
		return models.Announcement{}, fmt.Errorf("%w: invalid priority", ErrBadRequest)
	}
	return s.repo.CreateAnnouncement(ctx, models.Announcement{Title: title, Content: content, Priority: priority})
}

func (s *Services) ListPolls(ctx context.Context, authUser AuthUser) ([]models.Poll, error) {
	return s.repo.ListPolls(ctx, authUser.ID)
}

func (s *Services) CreatePoll(ctx context.Context, authUser AuthUser, title string, anonymous bool, endAt time.Time, options []string) (models.Poll, error) {
	if authUser.Role != models.RoleAdmin {
		return models.Poll{}, ErrForbidden
	}
	if strings.TrimSpace(title) == "" || len(options) < 2 {
		return models.Poll{}, fmt.Errorf("%w: title and at least two options are required", ErrBadRequest)
	}
	for i, option := range options {
		options[i] = strings.TrimSpace(option)
		if options[i] == "" {
			return models.Poll{}, fmt.Errorf("%w: options must not be empty", ErrBadRequest)
		}
	}
	if !endAt.After(time.Now()) {
		return models.Poll{}, fmt.Errorf("%w: end_at must be in the future", ErrBadRequest)
	}
	createdBy := authUser.ID
	return s.repo.CreatePoll(ctx, models.Poll{
		Title:     title,
		Anonymous: anonymous,
		CreatedBy: &createdBy,
		EndAt:     endAt,
		Status:    models.PollActive,
	}, options)
}

func (s *Services) VotePoll(ctx context.Context, authUser AuthUser, pollID, optionID int64) error {
	if err := s.repo.VotePoll(ctx, pollID, optionID, authUser.ID); err != nil {
		if strings.Contains(err.Error(), "already voted") || strings.Contains(err.Error(), "closed") {
			return fmt.Errorf("%w: %s", ErrBadRequest, err.Error())
		}
		return err
	}
	return nil
}

func (s *Services) DeletePoll(ctx context.Context, authUser AuthUser, pollID int64) error {
	if authUser.Role != models.RoleAdmin {
		return ErrForbidden
	}
	return s.repo.DeletePoll(ctx, pollID)
}

func (s *Services) SearchFAQ(ctx context.Context, query string) ([]models.FAQEntry, error) {
	return s.repo.SearchFAQ(ctx, query)
}

func (s *Services) FAQByTopic(ctx context.Context, topic string) ([]models.FAQEntry, error) {
	return s.repo.FAQByTopic(ctx, topic)
}

func (s *Services) FAQKeywords(ctx context.Context) ([]models.FAQKeyword, error) {
	return s.repo.FAQKeywords(ctx)
}

func (s *Services) ListFAQAdmin(ctx context.Context, authUser AuthUser) ([]models.FAQEntry, error) {
	if authUser.Role != models.RoleAdmin {
		return nil, ErrForbidden
	}
	return s.repo.ListFAQEntries(ctx)
}

func (s *Services) CreateFAQEntry(ctx context.Context, authUser AuthUser, topic string, keywords []string, question, answer string) (models.FAQEntry, error) {
	if authUser.Role != models.RoleAdmin {
		return models.FAQEntry{}, ErrForbidden
	}
	return s.repo.CreateFAQEntry(ctx, models.FAQEntry{Topic: topic, Keywords: keywords, Question: question, Answer: answer})
}

func (s *Services) DeleteFAQEntry(ctx context.Context, authUser AuthUser, id int64) error {
	if authUser.Role != models.RoleAdmin {
		return ErrForbidden
	}
	return s.repo.DeleteFAQEntry(ctx, id)
}

func (s *Services) NotificationSettings(ctx context.Context, authUser AuthUser) (models.NotificationSettings, error) {
	return s.repo.NotificationSettings(ctx, authUser.ID)
}

func (s *Services) UpdateNotificationSettings(ctx context.Context, authUser AuthUser, item models.NotificationSettings) (models.NotificationSettings, error) {
	item.UserID = authUser.ID
	return s.repo.UpsertNotificationSettings(ctx, item)
}

func (s *Services) ListNotifications(ctx context.Context, authUser AuthUser) ([]models.Notification, error) {
	return s.repo.ListNotifications(ctx, authUser.ID)
}

func (s *Services) MarkNotificationRead(ctx context.Context, authUser AuthUser, id int64) error {
	return s.repo.MarkNotificationRead(ctx, id, authUser.ID)
}

func (s *Services) AnalyticsSummary(ctx context.Context, authUser AuthUser) (models.AnalyticsSummary, error) {
	if authUser.Role != models.RoleAdmin {
		return models.AnalyticsSummary{}, ErrForbidden
	}
	return s.repo.AnalyticsSummary(ctx)
}

func (s *Services) AnalyticsCharts(ctx context.Context, authUser AuthUser) (models.AnalyticsCharts, error) {
	if authUser.Role != models.RoleAdmin {
		return models.AnalyticsCharts{}, ErrForbidden
	}
	return s.repo.AnalyticsCharts(ctx)
}

func (s *Services) RunWorkerCycle(ctx context.Context) error {
	if _, err := s.repo.CloseExpiredPolls(ctx); err != nil {
		return err
	}
	bookings, err := s.repo.UpcomingBookings(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, booking := range bookings {
		settings, err := s.repo.NotificationSettings(ctx, booking.UserID)
		if err != nil || !settings.BookingsEnabled {
			continue
		}
		startDiff := booking.StartAt.Sub(now)
		if startDiff > 0 && startDiff <= time.Hour {
			title := "Скоро начнётся бронирование"
			message := fmt.Sprintf("Бронирование %s начнётся в %d мин.", booking.ResourceType, int(startDiff.Minutes()+0.5))
			_ = s.repo.CreateNotification(ctx, booking.UserID, "booking", title, message, fmt.Sprintf("booking-start:%d", booking.ID))
		}
		endDiff := booking.EndAt.Sub(now)
		if endDiff > 0 && endDiff <= 15*time.Minute {
			title := "Скоро закончится бронирование"
			message := fmt.Sprintf("Бронирование %s закончится в %d мин.", booking.ResourceType, int(endDiff.Minutes()+0.5))
			_ = s.repo.CreateNotification(ctx, booking.UserID, "booking", title, message, fmt.Sprintf("booking-end:%d", booking.ID))
		}
	}
	polls, err := s.repo.ActivePollReminders(ctx)
	if err != nil {
		return err
	}
	for _, poll := range polls {
		hoursLeft := int(time.Until(poll.EndAt).Hours()) + 1
		message := fmt.Sprintf("%s — осталось менее %d ч.", poll.Title, hoursLeft)
		_ = s.repo.CreateNotification(ctx, poll.UserID, "poll", "Голосование скоро завершится", message, fmt.Sprintf("poll-ending:%d:%d", poll.PollID, poll.UserID))
	}
	return nil
}
