package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IntegrationTest содержит helper'ы для интеграционных тестов
type IntegrationTest struct {
	db      *pgxpool.Pool
	client  *http.Client
	baseURL string
}

func NewIntegrationTest(t *testing.T) *IntegrationTest {
	// Подключение к тестовой БД
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/pr_reviewers_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return &IntegrationTest{
		db:      db,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "http://localhost:8080",
	}
}

func (it *IntegrationTest) Close() error {
	it.db.Close()
	return nil
}

func (it *IntegrationTest) CleanupDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		if _, err := it.db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("Failed to cleanup table %s: %v", table, err)
		}
	}
}

func (it *IntegrationTest) Post(endpoint string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return it.client.Post(
		fmt.Sprintf("%s%s", it.baseURL, endpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
}

func (it *IntegrationTest) Get(endpoint string) (*http.Response, error) {
	return it.client.Get(fmt.Sprintf("%s%s", it.baseURL, endpoint))
}

// Test Functions

func TestCreateTeam(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	body := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}

	resp, err := it.Post("/team/add", body)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	team, ok := result["team"].(map[string]interface{})
	if !ok {
		t.Fatal("Team not found in response")
	}

	if team["team_name"] != "backend" {
		t.Errorf("Expected team_name 'backend', got %v", team["team_name"])
	}
}

func TestGetTeam(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	// Create team first
	createBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}
	it.Post("/team/add", createBody)

	// Get team
	resp, err := it.Get("/team/get?team_name=backend")
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreatePR(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	// Create team
	teamBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	it.Post("/team/add", teamBody)

	// Create PR
	prBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Feature X",
		"author_id":         "u1",
	}

	resp, err := it.Post("/pullRequest/create", prBody)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("PR not found in response")
	}

	// Check reviewers
	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	if !ok {
		t.Fatal("assigned_reviewers not found")
	}

	if len(reviewers) != 2 {
		t.Errorf("Expected 2 reviewers, got %d", len(reviewers))
	}
}

func TestMergePR(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	// Setup: Create team and PR
	teamBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	it.Post("/team/add", teamBody)

	prBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Feature X",
		"author_id":         "u1",
	}
	it.Post("/pullRequest/create", prBody)

	// Merge PR
	mergeBody := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	resp, err := it.Post("/pullRequest/merge", mergeBody)
	if err != nil {
		t.Fatalf("Failed to merge PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("PR not found in response")
	}

	if pr["status"] != "MERGED" {
		t.Errorf("Expected status MERGED, got %v", pr["status"])
	}
}

func TestSetUserActive(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	// Create team
	teamBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}
	it.Post("/team/add", teamBody)

	// Set user inactive
	userBody := map[string]interface{}{
		"user_id":   "u1",
		"is_active": false,
	}

	resp, err := it.Post("/users/setIsActive", userBody)
	if err != nil {
		t.Fatalf("Failed to set user active: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestIdempotentMerge(t *testing.T) {
	it := NewIntegrationTest(t)
	defer it.Close()
	defer it.CleanupDatabase(t)

	// Setup
	teamBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	it.Post("/team/add", teamBody)

	prBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Feature X",
		"author_id":         "u1",
	}
	it.Post("/pullRequest/create", prBody)

	// Merge twice
	mergeBody := map[string]interface{}{"pull_request_id": "pr-1"}
	resp1, _ := it.Post("/pullRequest/merge", mergeBody)
	resp1.Body.Close()

	resp2, _ := it.Post("/pullRequest/merge", mergeBody)
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Second merge should succeed, got %d", resp2.StatusCode)
	}
}
