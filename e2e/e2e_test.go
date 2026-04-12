package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"dormitory-helper-backend/internal/service"
	"dormitory-helper-backend/internal/storage"
	httptransport "dormitory-helper-backend/internal/transport/http"
)

func TestE2EFlows(t *testing.T) {
	databaseURL := testDatabaseURL(t)
	ctx := context.Background()

	db, err := storage.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db, "../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	resetTables(t, db)

	repo := storage.NewRepository(db)
	svc := service.New(repo, service.NewTokenManager("test-secret", 24*time.Hour))
	ts := &http.Server{Handler: httptransport.NewServer(svc, true).Handler()}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	go ts.Serve(listener)
	defer ts.Shutdown(context.Background())

	baseURL := "http://" + listener.Addr().String()
	adminToken := login(t, baseURL, "admin", "admin123")
	studentToken := login(t, baseURL, "student", "student123")

	bookingID := createBooking(t, baseURL, studentToken)
	assertStatus(t, doRequest(t, "GET", baseURL+"/api/v1/bookings/my?resource_type=laundry", studentToken, nil), http.StatusOK)
	assertStatus(t, doRequest(t, "POST", baseURL+"/api/v1/bookings", studentToken, map[string]any{
		"resource_type": "laundry",
		"start_at":      time.Now().Add(31 * time.Minute).UTC(),
		"end_at":        time.Now().Add(90 * time.Minute).UTC(),
	}), http.StatusBadRequest)

	repairID := createRepair(t, baseURL, studentToken)
	assertStatus(t, doRequest(t, "PATCH", fmt.Sprintf("%s/api/v1/admin/repairs/%d/status", baseURL, repairID), adminToken, map[string]any{"status": "completed"}), http.StatusOK)

	exchangeID := createExchange(t, baseURL, studentToken)
	assertStatus(t, doRequest(t, "DELETE", fmt.Sprintf("%s/api/v1/exchange/%d", baseURL, exchangeID), studentToken, nil), http.StatusOK)

	assertStatus(t, doRequest(t, "POST", baseURL+"/api/v1/admin/announcements", adminToken, map[string]any{
		"title":    "Проверка пожарной сигнализации",
		"content":  "Завтра в 12:00",
		"priority": "high",
	}), http.StatusCreated)

	pollID, optionID := createPoll(t, baseURL, adminToken)
	assertStatus(t, doRequest(t, "POST", fmt.Sprintf("%s/api/v1/polls/%d/vote", baseURL, pollID), studentToken, map[string]any{"option_id": optionID}), http.StatusOK)
	assertStatus(t, doRequest(t, "POST", fmt.Sprintf("%s/api/v1/polls/%d/vote", baseURL, pollID), studentToken, map[string]any{"option_id": optionID}), http.StatusBadRequest)

	assertStatus(t, doRequest(t, "GET", baseURL+"/api/v1/faq/search?q=ремонт", studentToken, nil), http.StatusOK)
	assertStatus(t, doRequest(t, "GET", baseURL+"/api/v1/admin/analytics/summary", adminToken, nil), http.StatusOK)

	if err := svc.RunWorkerCycle(context.Background()); err != nil {
		t.Fatalf("worker cycle: %v", err)
	}
	assertStatus(t, doRequest(t, "GET", baseURL+"/api/v1/notifications", studentToken, nil), http.StatusOK)
	assertStatus(t, doRequest(t, "DELETE", fmt.Sprintf("%s/api/v1/bookings/%d", baseURL, bookingID), studentToken, nil), http.StatusOK)
}

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	if url := getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	t.Skip("TEST_DATABASE_URL is not set")
	return ""
}

func getenv(key string) string {
	return os.Getenv(key)
}

func resetTables(t *testing.T, db *sql.DB) {
	t.Helper()
	queries := []string{
		"TRUNCATE notifications, poll_votes, poll_options, polls, announcements, exchange_items, repairs, bookings RESTART IDENTITY CASCADE",
		"DELETE FROM faq_entries",
		"INSERT INTO faq_entries (topic, keywords, question, answer) VALUES ('repairs', ARRAY['ремонт'], 'Как подать заявку?', 'Через приложение')",
		"DELETE FROM notification_settings",
		"INSERT INTO notification_settings (user_id) SELECT id FROM users ON CONFLICT DO NOTHING",
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("reset tables: %v", err)
		}
	}
}

func login(t *testing.T, baseURL, username, password string) string {
	t.Helper()
	resp := doRequest(t, "POST", baseURL+"/api/v1/auth/login", "", map[string]string{"username": username, "password": password})
	expectStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	return payload.AccessToken
}

func createBooking(t *testing.T, baseURL, token string) int64 {
	t.Helper()
	resp := doRequest(t, "POST", baseURL+"/api/v1/bookings", token, map[string]any{
		"resource_type": "laundry",
		"start_at":      time.Now().Add(30 * time.Minute).UTC(),
		"end_at":        time.Now().Add(90 * time.Minute).UTC(),
	})
	expectStatus(t, resp, http.StatusCreated)
	defer resp.Body.Close()
	var payload struct {
		ID int64 `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return payload.ID
}

func createRepair(t *testing.T, baseURL, token string) int64 {
	t.Helper()
	resp := doRequest(t, "POST", baseURL+"/api/v1/repairs", token, map[string]string{
		"location":    "Комната 305",
		"category":    "electrical",
		"description": "Не работает розетка",
	})
	expectStatus(t, resp, http.StatusCreated)
	defer resp.Body.Close()
	var payload struct {
		ID int64 `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return payload.ID
}

func createExchange(t *testing.T, baseURL, token string) int64 {
	t.Helper()
	resp := doRequest(t, "POST", baseURL+"/api/v1/exchange", token, map[string]string{
		"title":       "Лампа",
		"category":    "electronics",
		"type":        "give",
		"description": "Рабочая лампа",
		"contact":     "@student",
	})
	expectStatus(t, resp, http.StatusCreated)
	defer resp.Body.Close()
	var payload struct {
		ID int64 `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return payload.ID
}

func createPoll(t *testing.T, baseURL, token string) (int64, int64) {
	t.Helper()
	resp := doRequest(t, "POST", baseURL+"/api/v1/admin/polls", token, map[string]any{
		"title":     "Нужен ли второй холодильник?",
		"anonymous": true,
		"end_at":    time.Now().Add(2 * time.Hour).UTC(),
		"options":   []string{"Да", "Нет"},
	})
	expectStatus(t, resp, http.StatusCreated)
	defer resp.Body.Close()
	var payload struct {
		ID      int64 `json:"id"`
		Options []struct {
			ID int64 `json:"id"`
		} `json:"options"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return payload.ID, payload.Options[0].ID
}

func doRequest(t *testing.T, method, url, token string, body any) *http.Response {
	t.Helper()
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d, expected %d, body: %s", resp.StatusCode, expected, string(body))
	}
}

func expectStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("unexpected status %d, expected %d, body: %s", resp.StatusCode, expected, string(body))
	}
}
