package repo

import (
	"context"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
	"time"
)

// PRRepository интерфейс для работы с PR
type PRRepository interface {
	// CreatePR создает новый PR
	CreatePR(ctx context.Context, pr *domain.PullRequest) error

	// GetPRByID получает PR по ID
	GetPRByID(ctx context.Context, prID string) (*domain.PullRequest, error)

	// UpdateReviewers обновляет список ревьюверов
	UpdateReviewers(ctx context.Context, prID string, reviewers []string) error

	// UpdatePRStatus обновляет статус PR
	UpdatePRStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error

	// GetPRsByReviewer получает PR, где пользователь назначен ревьювером
	GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)

	// PRExists проверяет существование PR
	PRExists(ctx context.Context, prID string) (bool, error)
}
