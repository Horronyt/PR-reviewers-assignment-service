// internal/domain/entities.go
package domain

import (
	"time"
)

// User — пользователь системы
type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TeamName  string    `json:"team_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Team — команда (с загруженными участниками, если нужно)
type Team struct {
	TeamName  string    `json:"team_name"`
	Members   []User    `json:"members,omitempty"` // опционально, если запрашиваем с участниками
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// PullRequest — полный объект PR для внешнего API
type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`   // только ID ревьюверов
	CreatedAt         time.Time  `json:"created_at,omitempty"` // теперь единообразно: snake_case + omitempty
	MergedAt          *time.Time `json:"merged_at,omitempty"`
}

// PullRequestShort — укороченная версия (например, для списка у ревьювера)
type PullRequestShort struct {
	PullRequestID   string    `json:"pull_request_id"`
	PullRequestName string    `json:"pull_request_name"`
	AuthorID        string    `json:"author_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

// Константы статуса PR
const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

// === Доменные ошибки ===
type ErrorCode string

const (
	ErrorCodeTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists     ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged     ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

type DomainError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e DomainError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func NewError(code ErrorCode, message string) error {
	return DomainError{Code: code, Message: message}
}

// === Вспомогательная структура (не JSON, только внутри приложения) ===

// PRReviewerAssignment — если вдруг понадобится хранить время назначения
// (пока не используется в API, но может быть полезно в статистике)
type PRReviewerAssignment struct {
	PullRequestID string
	ReviewerID    string
	AssignedAt    time.Time
}
