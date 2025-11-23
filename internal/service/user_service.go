package service

import (
	"context"

	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/repo"
)

// UserService сервис для работы с пользователями
type UserService struct {
	userRepo  repo.UserRepository
	statsRepo repo.StatsRepository
}

// NewUserService создает новый сервис пользователей
func NewUserService(userRepo repo.UserRepository, statsRepo repo.StatsRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		statsRepo: statsRepo,
	}
}

// SetActive устанавливает флаг активности пользователя
func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetUserActive(ctx, userID, isActive)
	if err != nil {
		return nil, domain.NewError(domain.ErrorCodeNotFound, "user not found")
	}
	return user, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

// GetReviewerStats получает статистику ревьюверов
func (s *UserService) GetReviewerStats(ctx context.Context) ([]repo.ReviewerStats, error) {
	return s.statsRepo.GetReviewerStats(ctx)
}

// GetPRStats получает статистику PR
func (s *UserService) GetPRStats(ctx context.Context) (map[string]int, error) {
	return s.statsRepo.GetPRStats(ctx)
}
