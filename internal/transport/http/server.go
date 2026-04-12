package httptransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dormitory-helper-backend/internal/service"
	"dormitory-helper-backend/internal/storage"
)

type Server struct {
	svc         *service.Services
	corsEnabled bool
	mux         *http.ServeMux
}

func NewServer(svc *service.Services, corsEnabled bool) *Server {
	s := &Server{
		svc:         svc,
		corsEnabled: corsEnabled,
		mux:         http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	var handler http.Handler = s.mux
	handler = s.cors(handler)
	return handler
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/v1/auth/session", s.handleSession)
	s.mux.Handle("/api/v1/auth/me", s.auth(http.HandlerFunc(s.handleMe)))

	s.mux.Handle("/api/v1/bookings", s.auth(http.HandlerFunc(s.handleBookings)))
	s.mux.Handle("/api/v1/bookings/my", s.auth(http.HandlerFunc(s.handleMyBookings)))
	s.mux.Handle("/api/v1/bookings/", s.auth(http.HandlerFunc(s.handleBookingByID)))

	s.mux.Handle("/api/v1/repairs", s.auth(http.HandlerFunc(s.handleRepairs)))
	s.mux.Handle("/api/v1/repairs/my", s.auth(http.HandlerFunc(s.handleMyRepairs)))
	s.mux.Handle("/api/v1/repairs/", s.auth(http.HandlerFunc(s.handleRepairByID)))
	s.mux.Handle("/api/v1/admin/repairs/", s.auth(http.HandlerFunc(s.handleAdminRepairStatus)))

	s.mux.Handle("/api/v1/exchange", s.auth(http.HandlerFunc(s.handleExchange)))
	s.mux.Handle("/api/v1/exchange/", s.auth(http.HandlerFunc(s.handleExchangeByID)))

	s.mux.Handle("/api/v1/announcements", s.auth(http.HandlerFunc(s.handleAnnouncements)))
	s.mux.Handle("/api/v1/admin/announcements", s.auth(http.HandlerFunc(s.handleAdminAnnouncements)))

	s.mux.Handle("/api/v1/polls", s.auth(http.HandlerFunc(s.handlePolls)))
	s.mux.Handle("/api/v1/polls/", s.auth(http.HandlerFunc(s.handlePollByID)))
	s.mux.Handle("/api/v1/admin/polls", s.auth(http.HandlerFunc(s.handleAdminPolls)))
	s.mux.Handle("/api/v1/admin/polls/", s.auth(http.HandlerFunc(s.handleAdminPollByID)))

	s.mux.Handle("/api/v1/faq/search", s.auth(http.HandlerFunc(s.handleFAQSearch)))
	s.mux.Handle("/api/v1/faq/keywords", s.auth(http.HandlerFunc(s.handleFAQKeywords)))
	s.mux.Handle("/api/v1/faq/topics/", s.auth(http.HandlerFunc(s.handleFAQByTopic)))
	s.mux.Handle("/api/v1/admin/faq", s.auth(http.HandlerFunc(s.handleAdminFAQ)))
	s.mux.Handle("/api/v1/admin/faq/", s.auth(http.HandlerFunc(s.handleAdminFAQByID)))

	s.mux.Handle("/api/v1/notifications", s.auth(http.HandlerFunc(s.handleNotifications)))
	s.mux.Handle("/api/v1/notifications/", s.auth(http.HandlerFunc(s.handleNotificationByID)))
	s.mux.Handle("/api/v1/notification-settings", s.auth(http.HandlerFunc(s.handleNotificationSettings)))

	s.mux.Handle("/api/v1/admin/analytics/summary", s.auth(http.HandlerFunc(s.handleAnalyticsSummary)))
	s.mux.Handle("/api/v1/admin/analytics/charts", s.auth(http.HandlerFunc(s.handleAnalyticsCharts)))
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		user, err := s.svc.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(service.ContextWithUser(r.Context(), user)))
	})
}

func (s *Server) cors(next http.Handler) http.Handler {
	if !s.corsEnabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authUser(r *http.Request) service.AuthUser {
	user, _ := service.UserFromContext(r.Context())
	return user
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrBadRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, storage.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func parseIDPath(path, prefix string) (int64, error) {
	raw := strings.TrimPrefix(path, prefix)
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return 0, fmt.Errorf("missing id")
	}
	parts := strings.Split(raw, "/")
	return strconv.ParseInt(parts[0], 10, 64)
}

func parseLastPathPart(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionRequest struct {
	DeviceID string `json:"device_id"`
}

type bookingRequest struct {
	ResourceType string    `json:"resource_type"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
}

type repairRequest struct {
	Location    string `json:"location"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type repairStatusRequest struct {
	Status string `json:"status"`
}

type exchangeRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Contact     string `json:"contact"`
}

type announcementRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Priority string `json:"priority"`
}

type pollRequest struct {
	Title     string    `json:"title"`
	Anonymous bool      `json:"anonymous"`
	EndAt     time.Time `json:"end_at"`
	Options   []string  `json:"options"`
}

type voteRequest struct {
	OptionID int64 `json:"option_id"`
}

type faqRequest struct {
	Topic    string   `json:"topic"`
	Keywords []string `json:"keywords"`
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
}
