package repo

import (
	"context"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	// CreateOrUpdateUser создает или обновляет пользователя
	CreateOrUpdateUser(ctx context.Context, user *domain.User) error

	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)

	// GetUsersByTeam получает всех пользователей команды
	GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)

	// GetActiveUsers получает активных пользователей команды
	GetActiveUsers(ctx context.Context, teamName string) ([]domain.User, error)

	// SetUserActive устанавливает флаг активности
	SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)

	// GetAllUsersByIDs получает пользователей по списку ID
	GetAllUsersByIDs(ctx context.Context, userIDs []string) ([]domain.User, error)
}
