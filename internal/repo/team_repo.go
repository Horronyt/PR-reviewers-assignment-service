package repo

import (
	"context"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
)

// TeamRepository интерфейс для работы с командами
type TeamRepository interface {
	// CreateTeam создает новую команду
	CreateTeam(ctx context.Context, team *domain.Team) error

	// GetTeamByName получает команду по имени
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)

	// TeamExists проверяет существование команды
	TeamExists(ctx context.Context, teamName string) (bool, error)

	// GetTeamMembers получает членов команды
	GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error)
}
