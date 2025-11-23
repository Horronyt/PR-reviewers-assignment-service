package service

import (
	"context"
	"fmt"

	"github.com/horronyt/pr-reviewers-service/internal/domain"
	"github.com/horronyt/pr-reviewers-service/internal/repo"
)

// TeamService сервис для работы с командами
type TeamService struct {
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
}

// NewTeamService создает новый сервис команд
func NewTeamService(teamRepo repo.TeamRepository, userRepo repo.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam создает команду с участниками
func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	// Проверяем существование команды
	exists, err := s.teamRepo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewError(domain.ErrorCodeTeamExists, "team already exists")
	}

	// Создаем команду
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	// Создаем/обновляем пользователей
	for i := range team.Members {
		team.Members[i].TeamName = team.TeamName
		if err := s.userRepo.CreateOrUpdate(ctx, &team.Members[i]); err != nil {
			return nil, fmt.Errorf("failed to create team member: %w", err)
		}
	}

	return team, nil
}

// GetTeam получает команду с членами
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, domain.NewError(domain.ErrorCodeNotFound, "team not found")
	}
	return team, nil
}

// GetTeamMembers получает членов команды
func (s *TeamService) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	return s.teamRepo.GetTeamMembers(ctx, teamName)
}
