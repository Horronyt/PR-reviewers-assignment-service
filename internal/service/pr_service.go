package service

import (
	"context"
	"fmt"
	"time"

	"github.com/horronyt/pr-reviewers-service/internal/domain"
	"github.com/horronyt/pr-reviewers-service/internal/repo"
)

// PRService сервис для работы с PR
type PRService struct {
	prRepo        repo.PRRepository
	userRepo      repo.UserRepository
	assignmentSvc *ReviewerAssignmentService
}

// NewPRService создает новый сервис PR
func NewPRService(
	prRepo repo.PRRepository,
	userRepo repo.UserRepository,
	assignmentSvc *ReviewerAssignmentService,
) *PRService {
	return &PRService{
		prRepo:        prRepo,
		userRepo:      userRepo,
		assignmentSvc: assignmentSvc,
	}
}

// CreatePR создает новый PR и назначает ревьюверов
func (s *PRService) CreatePR(ctx context.Context, prID, name, authorID string) (*domain.PullRequest, error) {
	// Проверяем существование PR
	exists, err := s.prRepo.PRExists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewError(domain.ErrorCodePRExists, "PR id already exists")
	}

	// Проверяем существование автора
	_, err = s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, domain.NewError(domain.ErrorCodeNotFound, "author not found")
	}

	// Создаем PR
	pr := &domain.PullRequest{
		PullRequestID:   prID,
		PullRequestName: name,
		AuthorID:        authorID,
		Status:          domain.PRStatusOpen,
		CreatedAt:       time.Now(),
	}

	// Назначаем ревьюверов
	reviewers, err := s.assignmentSvc.AssignReviewers(ctx, pr)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	// Сохраняем PR
	if err := s.prRepo.CreatePR(ctx, pr); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

// GetPR получает PR по ID
func (s *PRService) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.prRepo.GetPRByID(ctx, prID)
}

// MergePR помечает PR как MERGED (идемпотентно)
func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, domain.NewError(domain.ErrorCodeNotFound, "PR not found")
	}

	// Если уже merged - просто возвращаем
	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	// Обновляем статус
	now := time.Now()
	if err := s.prRepo.UpdatePRStatus(ctx, prID, domain.PRStatusMerged, &now); err != nil {
		return nil, err
	}

	pr.Status = domain.PRStatusMerged
	pr.MergedAt = &now
	return pr, nil
}

// ReassignReviewer переназначает ревьювера на PR
func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	newReviewerID, err := s.assignmentSvc.ReassignReviewer(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

// GetReviewsForUser получает PR, где пользователь назначен ревьювером
func (s *PRService) GetReviewsForUser(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	return s.prRepo.GetPRByReviewer(ctx, userID)
}
