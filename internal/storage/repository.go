package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dormitory-helper-backend/internal/models"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (models.User, string, error) {
	const query = `
SELECT id, username, password_hash, role, created_at
FROM users
WHERE username = $1`
	var user models.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &passwordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, "", ErrNotFound
	}
	return user, passwordHash, err
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	const query = `
SELECT id, username, role, COALESCE(device_id, ''), created_at
FROM users
WHERE id = $1`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Role, &user.DeviceID, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return user, err
}

func (r *Repository) GetUserByDeviceID(ctx context.Context, deviceID string) (models.User, error) {
	const query = `
SELECT id, username, role, COALESCE(device_id, ''), created_at
FROM users
WHERE device_id = $1`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(&user.ID, &user.Username, &user.Role, &user.DeviceID, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return user, err
}

func (r *Repository) CreateGuestUser(ctx context.Context, username, deviceID string) (models.User, error) {
	const query = `
INSERT INTO users (username, password_hash, role, device_id)
VALUES ($1, '', 'student', $2)
RETURNING id, username, role, COALESCE(device_id, ''), created_at`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, username, deviceID).Scan(&user.ID, &user.Username, &user.Role, &user.DeviceID, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	_, _ = r.db.ExecContext(ctx, "INSERT INTO notification_settings (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", user.ID)
	return user, nil
}

func (r *Repository) ListBookings(ctx context.Context, resourceType string, onlyUserID *int64) ([]models.Booking, error) {
	query := `
SELECT id, user_id, resource_type, start_at, end_at, created_at
FROM bookings
WHERE resource_type = $1`
	args := []any{resourceType}
	if onlyUserID != nil {
		query += " AND user_id = $2"
		args = append(args, *onlyUserID)
	}
	query += " ORDER BY start_at ASC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Booking
	for rows.Next() {
		var item models.Booking
		if err := rows.Scan(&item.ID, &item.UserID, &item.ResourceType, &item.StartAt, &item.EndAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) HasBookingConflict(ctx context.Context, resourceType string, startAt, endAt time.Time) (bool, error) {
	const query = `
SELECT EXISTS (
    SELECT 1
    FROM bookings
    WHERE resource_type = $1
      AND start_at < $3
      AND end_at > $2
)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, resourceType, startAt, endAt).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateBooking(ctx context.Context, booking models.Booking) (models.Booking, error) {
	const query = `
INSERT INTO bookings (user_id, resource_type, start_at, end_at)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, booking.UserID, booking.ResourceType, booking.StartAt, booking.EndAt).Scan(&booking.ID, &booking.CreatedAt)
	return booking, err
}

func (r *Repository) GetBookingOwner(ctx context.Context, id int64) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM bookings WHERE id = $1", id).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return userID, err
}

func (r *Repository) DeleteBooking(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM bookings WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListRepairs(ctx context.Context, onlyUserID *int64) ([]models.Repair, error) {
	query := `
SELECT id, user_id, location, category, description, status, created_at, updated_at
FROM repairs`
	args := []any{}
	if onlyUserID != nil {
		query += " WHERE user_id = $1"
		args = append(args, *onlyUserID)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Repair
	for rows.Next() {
		var item models.Repair
		if err := rows.Scan(&item.ID, &item.UserID, &item.Location, &item.Category, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateRepair(ctx context.Context, repair models.Repair) (models.Repair, error) {
	const query = `
INSERT INTO repairs (user_id, location, category, description, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, repair.UserID, repair.Location, repair.Category, repair.Description, repair.Status).Scan(&repair.ID, &repair.CreatedAt, &repair.UpdatedAt)
	return repair, err
}

func (r *Repository) GetRepairByID(ctx context.Context, id int64) (models.Repair, error) {
	const query = `
SELECT id, user_id, location, category, description, status, created_at, updated_at
FROM repairs
WHERE id = $1`
	var item models.Repair
	err := r.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.UserID, &item.Location, &item.Category, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Repair{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) DeleteRepair(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM repairs WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateRepairStatus(ctx context.Context, id int64, status string) (models.Repair, error) {
	const query = `
UPDATE repairs
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, location, category, description, status, created_at, updated_at`
	var item models.Repair
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(&item.ID, &item.UserID, &item.Location, &item.Category, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Repair{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) ListExchangeItems(ctx context.Context) ([]models.ExchangeItem, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, title, category, type, description, contact, created_at
FROM exchange_items
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.ExchangeItem
	for rows.Next() {
		var item models.ExchangeItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Title, &item.Category, &item.Type, &item.Description, &item.Contact, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateExchangeItem(ctx context.Context, item models.ExchangeItem) (models.ExchangeItem, error) {
	err := r.db.QueryRowContext(ctx, `
INSERT INTO exchange_items (user_id, title, category, type, description, contact)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`,
		item.UserID, item.Title, item.Category, item.Type, item.Description, item.Contact,
	).Scan(&item.ID, &item.CreatedAt)
	return item, err
}

func (r *Repository) GetExchangeOwner(ctx context.Context, id int64) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM exchange_items WHERE id = $1", id).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return userID, err
}

func (r *Repository) DeleteExchangeItem(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM exchange_items WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListAnnouncements(ctx context.Context) ([]models.Announcement, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, title, content, priority, created_at
FROM announcements
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Announcement
	for rows.Next() {
		var item models.Announcement
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.Priority, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateAnnouncement(ctx context.Context, item models.Announcement) (models.Announcement, error) {
	err := r.db.QueryRowContext(ctx, `
INSERT INTO announcements (title, content, priority)
VALUES ($1, $2, $3)
RETURNING id, created_at`, item.Title, item.Content, item.Priority).Scan(&item.ID, &item.CreatedAt)
	return item, err
}

func (r *Repository) ListPolls(ctx context.Context, userID int64) ([]models.Poll, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT p.id, p.title, p.anonymous, p.created_by, p.end_at, p.status, p.created_at,
       EXISTS (SELECT 1 FROM poll_votes pv WHERE pv.poll_id = p.id AND pv.user_id = $1) AS has_voted
FROM polls p
ORDER BY CASE WHEN p.status = 'active' THEN 0 ELSE 1 END, p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var polls []models.Poll
	for rows.Next() {
		var poll models.Poll
		var createdBy sql.NullInt64
		if err := rows.Scan(&poll.ID, &poll.Title, &poll.Anonymous, &createdBy, &poll.EndAt, &poll.Status, &poll.CreatedAt, &poll.HasVoted); err != nil {
			return nil, err
		}
		if createdBy.Valid {
			poll.CreatedBy = &createdBy.Int64
		}
		options, err := r.listPollOptions(ctx, poll.ID)
		if err != nil {
			return nil, err
		}
		poll.Options = options
		for _, opt := range options {
			poll.TotalVotes += opt.Votes
		}
		polls = append(polls, poll)
	}
	return polls, rows.Err()
}

func (r *Repository) listPollOptions(ctx context.Context, pollID int64) ([]models.PollOption, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, option_text, votes
FROM poll_options
WHERE poll_id = $1
ORDER BY id ASC`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.PollOption
	for rows.Next() {
		var item models.PollOption
		if err := rows.Scan(&item.ID, &item.Text, &item.Votes); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreatePoll(ctx context.Context, poll models.Poll, options []string) (models.Poll, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Poll{}, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `
INSERT INTO polls (title, anonymous, created_by, end_at, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`, poll.Title, poll.Anonymous, poll.CreatedBy, poll.EndAt, poll.Status).Scan(&poll.ID, &poll.CreatedAt)
	if err != nil {
		return models.Poll{}, err
	}
	for _, option := range options {
		if _, err := tx.ExecContext(ctx, `INSERT INTO poll_options (poll_id, option_text) VALUES ($1, $2)`, poll.ID, option); err != nil {
			return models.Poll{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return models.Poll{}, err
	}
	polls, err := r.ListPolls(ctx, 0)
	if err != nil {
		return models.Poll{}, err
	}
	for _, item := range polls {
		if item.ID == poll.ID {
			return item, nil
		}
	}
	return poll, nil
}

func (r *Repository) VotePoll(ctx context.Context, pollID, optionID, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	if err := tx.QueryRowContext(ctx, "SELECT status FROM polls WHERE id = $1", pollID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status != models.PollActive {
		return fmt.Errorf("poll is closed")
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO poll_votes (poll_id, user_id, option_id)
VALUES ($1, $2, $3)`, pollID, userID, optionID); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("user already voted")
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE poll_options SET votes = votes + 1 WHERE id = $1 AND poll_id = $2", optionID, pollID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeletePoll(ctx context.Context, pollID int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM polls WHERE id = $1", pollID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CloseExpiredPolls(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE polls
SET status = 'closed'
WHERE status = 'active' AND end_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) ListFAQEntries(ctx context.Context) ([]models.FAQEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, topic, keywords, question, answer, created_at
FROM faq_entries
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFAQRows(rows)
}

func (r *Repository) SearchFAQ(ctx context.Context, query string) ([]models.FAQEntry, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, `
SELECT id, topic, keywords, question, answer, created_at
FROM faq_entries
WHERE LOWER(question) LIKE $1
   OR LOWER(answer) LIKE $1
   OR EXISTS (
       SELECT 1
       FROM unnest(keywords) kw
       WHERE LOWER(kw) LIKE $1
   )
ORDER BY created_at DESC`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFAQRows(rows)
}

func (r *Repository) FAQByTopic(ctx context.Context, topic string) ([]models.FAQEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, topic, keywords, question, answer, created_at
FROM faq_entries
WHERE topic = $1
ORDER BY created_at DESC`, topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFAQRows(rows)
}

func (r *Repository) FAQKeywords(ctx context.Context) ([]models.FAQKeyword, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT DISTINCT ON (topic)
       topic,
       COALESCE(NULLIF(keywords[1], ''), topic) AS keyword
FROM faq_entries
ORDER BY topic, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.FAQKeyword
	for rows.Next() {
		var item models.FAQKeyword
		if err := rows.Scan(&item.Topic, &item.Keyword); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func scanFAQRows(rows *sql.Rows) ([]models.FAQEntry, error) {
	var result []models.FAQEntry
	for rows.Next() {
		var item models.FAQEntry
		var keywords pq.StringArray
		if err := rows.Scan(&item.ID, &item.Topic, &keywords, &item.Question, &item.Answer, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Keywords = []string(keywords)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateFAQEntry(ctx context.Context, item models.FAQEntry) (models.FAQEntry, error) {
	var keywords pq.StringArray = item.Keywords
	err := r.db.QueryRowContext(ctx, `
INSERT INTO faq_entries (topic, keywords, question, answer)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at`, item.Topic, keywords, item.Question, item.Answer).Scan(&item.ID, &item.CreatedAt)
	return item, err
}

func (r *Repository) DeleteFAQEntry(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM faq_entries WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) NotificationSettings(ctx context.Context, userID int64) (models.NotificationSettings, error) {
	var item models.NotificationSettings
	err := r.db.QueryRowContext(ctx, `
SELECT user_id, bookings_enabled, repairs_enabled, polls_enabled, updated_at
FROM notification_settings
WHERE user_id = $1`, userID).Scan(&item.UserID, &item.BookingsEnabled, &item.RepairsEnabled, &item.PollsEnabled, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.NotificationSettings{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) UpsertNotificationSettings(ctx context.Context, item models.NotificationSettings) (models.NotificationSettings, error) {
	err := r.db.QueryRowContext(ctx, `
INSERT INTO notification_settings (user_id, bookings_enabled, repairs_enabled, polls_enabled, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (user_id) DO UPDATE
SET bookings_enabled = EXCLUDED.bookings_enabled,
    repairs_enabled = EXCLUDED.repairs_enabled,
    polls_enabled = EXCLUDED.polls_enabled,
    updated_at = NOW()
RETURNING user_id, bookings_enabled, repairs_enabled, polls_enabled, updated_at`,
		item.UserID, item.BookingsEnabled, item.RepairsEnabled, item.PollsEnabled).Scan(
		&item.UserID, &item.BookingsEnabled, &item.RepairsEnabled, &item.PollsEnabled, &item.UpdatedAt,
	)
	return item, err
}

func (r *Repository) ListNotifications(ctx context.Context, userID int64) ([]models.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, type, title, message, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Notification
	for rows.Next() {
		var item models.Notification
		if err := rows.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Message, &item.IsRead, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreateNotification(ctx context.Context, userID int64, notificationType, title, message, dedupeKey string) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO notifications (user_id, type, title, message, dedupe_key)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (dedupe_key) DO NOTHING`, userID, notificationType, title, message, dedupeKey)
	return err
}

func (r *Repository) AnalyticsSummary(ctx context.Context) (models.AnalyticsSummary, error) {
	var item models.AnalyticsSummary
	err := r.db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM bookings WHERE resource_type = 'laundry'),
  (SELECT COUNT(*) FROM bookings WHERE resource_type = 'kitchen'),
  (SELECT COUNT(*) FROM repairs),
  (SELECT COUNT(*) FROM exchange_items),
  (SELECT COUNT(*) FROM polls),
  (SELECT COUNT(*) FROM users)`).Scan(
		&item.LaundryBookings, &item.KitchenBookings, &item.Repairs, &item.ExchangeItems, &item.Polls, &item.Users,
	)
	return item, err
}

func (r *Repository) chartQuery(ctx context.Context, query string, args ...any) ([]models.ChartValue, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.ChartValue
	for rows.Next() {
		var item models.ChartValue
		if err := rows.Scan(&item.Label, &item.Value); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) AnalyticsCharts(ctx context.Context) (models.AnalyticsCharts, error) {
	var charts models.AnalyticsCharts
	var err error
	charts.RepairStatuses, err = r.chartQuery(ctx, `
SELECT status, COUNT(*)::BIGINT
FROM repairs
GROUP BY status
ORDER BY status`)
	if err != nil {
		return charts, err
	}
	charts.ExchangeTypes, err = r.chartQuery(ctx, `
SELECT type, COUNT(*)::BIGINT
FROM exchange_items
GROUP BY type
ORDER BY type`)
	if err != nil {
		return charts, err
	}
	charts.WeeklyActivity, err = r.chartQuery(ctx, `
SELECT TO_CHAR(day_bucket, 'YYYY-MM-DD') AS label, COUNT(*)::BIGINT
FROM (
    SELECT DATE_TRUNC('day', created_at) AS day_bucket FROM bookings
    UNION ALL
    SELECT DATE_TRUNC('day', created_at) AS day_bucket FROM repairs
    UNION ALL
    SELECT DATE_TRUNC('day', created_at) AS day_bucket FROM exchange_items
    UNION ALL
    SELECT DATE_TRUNC('day', created_at) AS day_bucket FROM polls
) t
WHERE day_bucket >= DATE_TRUNC('day', NOW()) - INTERVAL '6 day'
GROUP BY day_bucket
ORDER BY day_bucket`)
	if err != nil {
		return charts, err
	}
	charts.ActivePolls, err = r.chartQuery(ctx, `
SELECT p.title, COALESCE(SUM(po.votes), 0)::BIGINT
FROM polls p
LEFT JOIN poll_options po ON po.poll_id = p.id
WHERE p.status = 'active'
GROUP BY p.id, p.title
ORDER BY p.created_at DESC
LIMIT 10`)
	return charts, err
}

func (r *Repository) UpcomingBookings(ctx context.Context) ([]models.Booking, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, resource_type, start_at, end_at, created_at
FROM bookings
WHERE end_at > NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Booking
	for rows.Next() {
		var item models.Booking
		if err := rows.Scan(&item.ID, &item.UserID, &item.ResourceType, &item.StartAt, &item.EndAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ActivePollReminders(ctx context.Context) ([]struct {
	PollID int64
	UserID int64
	Title  string
	EndAt  time.Time
}, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT p.id, u.id, p.title, p.end_at
FROM polls p
JOIN users u ON u.role = 'student'
JOIN notification_settings ns ON ns.user_id = u.id AND ns.polls_enabled = TRUE
LEFT JOIN poll_votes pv ON pv.poll_id = p.id AND pv.user_id = u.id
WHERE p.status = 'active'
  AND p.end_at > NOW()
  AND p.end_at <= NOW() + INTERVAL '24 hour'
  AND pv.id IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []struct {
		PollID int64
		UserID int64
		Title  string
		EndAt  time.Time
	}
	for rows.Next() {
		var item struct {
			PollID int64
			UserID int64
			Title  string
			EndAt  time.Time
		}
		if err := rows.Scan(&item.PollID, &item.UserID, &item.Title, &item.EndAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
