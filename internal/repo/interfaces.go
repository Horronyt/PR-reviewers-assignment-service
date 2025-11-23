package repo

import (
	"context"
	"github.com/horronyt/pr-reviewers-service/internal/domain"
	"time"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	// CreateOrUpdate создает или обновляет пользователя
	CreateOrUpdate(ctx context.Context, user *domain.User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// GetByTeam получает всех пользователей команды
	GetByTeam(ctx context.Context, teamName string) ([]domain.User, error)

	// GetActive получает активных пользователей команды
	GetActive(ctx context.Context, teamName string) ([]domain.User, error)

	// SetActive устанавливает флаг активности
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)

	// GetAllByIDs получает пользователей по списку ID
	GetAllByIDs(ctx context.Context, userIDs []string) ([]domain.User, error)
}

// TeamRepository интерфейс для работы с командами
type TeamRepository interface {
	// Create создает новую команду
	Create(ctx context.Context, team *domain.Team) error

	// GetByName получает команду по имени
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)

	// Exists проверяет существование команды
	Exists(ctx context.Context, teamName string) (bool, error)

	// GetMembers получает членов команды
	GetMembers(ctx context.Context, teamName string) ([]domain.User, error)
}

// PRRepository интерфейс для работы с PR
type PRRepository interface {
	// Create создает новый PR
	Create(ctx context.Context, pr *domain.PullRequest) error

	// GetByID получает PR по ID
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)

	// UpdateReviewers обновляет список ревьюверов
	UpdateReviewers(ctx context.Context, prID string, reviewers []string) error

	// UpdateStatus обновляет статус PR
	UpdateStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error

	// GetByReviewer получает PR, где пользователь назначен ревьювером
	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)

	// Exists проверяет существование PR
	Exists(ctx context.Context, prID string) (bool, error)
}

// Статистика
type ReviewerStats struct {
	UserID          string
	AssignmentCount int
}

type StatsRepository interface {
	// GetReviewerStats получает статистику по ревьюверам
	GetReviewerStats(ctx context.Context) ([]ReviewerStats, error)

	// GetPRStats получает статистику по PR
	GetPRStats(ctx context.Context) (map[string]int, error)
}
