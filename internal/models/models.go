package models

import "time"

const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

const (
	ResourceLaundry = "laundry"
	ResourceKitchen = "kitchen"
)

const (
	RepairPending    = "pending"
	RepairInProgress = "in_progress"
	RepairCompleted  = "completed"
	RepairRejected   = "rejected"
)

const (
	PollActive = "active"
	PollClosed = "closed"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	DeviceID  string    `json:"device_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Booking struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	ResourceType string    `json:"resource_type"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repair struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Location    string    `json:"location"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExchangeItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Contact     string    `json:"contact"`
	CreatedAt   time.Time `json:"created_at"`
}

type Announcement struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

type PollOption struct {
	ID    int64  `json:"id"`
	Text  string `json:"text"`
	Votes int64  `json:"votes"`
}

type Poll struct {
	ID         int64        `json:"id"`
	Title      string       `json:"title"`
	Anonymous  bool         `json:"anonymous"`
	CreatedBy  *int64       `json:"created_by,omitempty"`
	EndAt      time.Time    `json:"end_at"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	Options    []PollOption `json:"options"`
	HasVoted   bool         `json:"has_voted"`
	TotalVotes int64        `json:"total_votes"`
}

type FAQEntry struct {
	ID        int64     `json:"id"`
	Topic     string    `json:"topic"`
	Keywords  []string  `json:"keywords"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

type FAQKeyword struct {
	Topic   string `json:"topic"`
	Keyword string `json:"keyword"`
}

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationSettings struct {
	UserID          int64     `json:"user_id"`
	BookingsEnabled bool      `json:"bookings_enabled"`
	RepairsEnabled  bool      `json:"repairs_enabled"`
	PollsEnabled    bool      `json:"polls_enabled"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AnalyticsSummary struct {
	LaundryBookings int64 `json:"laundry_bookings"`
	KitchenBookings int64 `json:"kitchen_bookings"`
	Repairs         int64 `json:"repairs"`
	ExchangeItems   int64 `json:"exchange_items"`
	Polls           int64 `json:"polls"`
	Users           int64 `json:"users"`
}

type ChartValue struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type AnalyticsCharts struct {
	RepairStatuses []ChartValue `json:"repair_statuses"`
	ExchangeTypes  []ChartValue `json:"exchange_types"`
	WeeklyActivity []ChartValue `json:"weekly_activity"`
	ActivePolls    []ChartValue `json:"active_polls"`
}
