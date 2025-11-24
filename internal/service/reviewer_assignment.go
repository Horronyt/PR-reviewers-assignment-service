package service

import (
	"context"
	"fmt"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/repo"
	"math/rand"

	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
)

// ReviewerAssignmentService сервис назначения ревьюверов
type ReviewerAssignmentService struct {
	userRepo repo.UserRepository
	teamRepo repo.TeamRepository
	prRepo   repo.PRRepository
}

// NewReviewerAssignmentService создает новый сервис
func NewReviewerAssignmentService(
	userRepo repo.UserRepository,
	teamRepo repo.TeamRepository,
	prRepo repo.PRRepository,
) *ReviewerAssignmentService {
	return &ReviewerAssignmentService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		prRepo:   prRepo,
	}
}

// AssignReviewers назначает до 2 ревьюверов на PR
func (s *ReviewerAssignmentService) AssignReviewers(ctx context.Context, pr *domain.PullRequest) ([]string, error) {
	// Получаем информацию об авторе
	author, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	// Получаем активных членов команды автора, исключая автора
	candidates, err := s.userRepo.GetActiveUsers(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	// Фильтруем: исключаем автора
	var availableCandidates []domain.User
	for _, candidate := range candidates {
		if candidate.UserID != pr.AuthorID {
			availableCandidates = append(availableCandidates, candidate)
		}
	}

	// Выбираем до 2 ревьюверов
	reviewersCount := 2
	if len(availableCandidates) < 2 {
		reviewersCount = len(availableCandidates)
	}

	// Перемешиваем и выбираем
	rand.Shuffle(len(availableCandidates), func(i, j int) {
		availableCandidates[i], availableCandidates[j] = availableCandidates[j], availableCandidates[i]
	})

	var assignedReviewers []string
	for i := 0; i < reviewersCount; i++ {
		assignedReviewers = append(assignedReviewers, availableCandidates[i].UserID)
	}

	return assignedReviewers, nil
}

// ReassignReviewer переназначает ревьювера
func (s *ReviewerAssignmentService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (string, error) {
	// Проверяем что PR существует и не merged
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return "", domain.NewError(domain.ErrorCodeNotFound, "PR not found")
	}

	if pr.Status == domain.PRStatusMerged {
		return "", domain.NewError(domain.ErrorCodePRMerged, "cannot reassign on merged PR")
	}

	// Проверяем что старый ревьювер назначен
	found := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return "", domain.NewError(domain.ErrorCodeNotAssigned, "reviewer is not assigned to this PR")
	}

	// Получаем команду старого ревьювера
	oldReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		return "", domain.NewError(domain.ErrorCodeNotFound, "old reviewer not found")
	}

	// Получаем активных членов его команды, исключая его самого
	candidates, err := s.userRepo.GetActiveUsers(ctx, oldReviewer.TeamName)
	if err != nil {
		return "", err
	}

	excluded := map[string]bool{
		oldReviewerID: true,
		pr.AuthorID:   true,
	}
	for _, r := range pr.AssignedReviewers {
		excluded[r] = true
	}

	var availableCandidates []domain.User
	for _, candidate := range candidates {
		if !excluded[candidate.UserID] && candidate.IsActive {
			availableCandidates = append(availableCandidates, candidate)
		}
	}

	if len(availableCandidates) == 0 {
		return "", domain.NewError(domain.ErrorCodeNoCandidate, "no active replacement candidate in team")
	}

	// Выбираем случайного кандидата
	newReviewerID := availableCandidates[rand.Intn(len(availableCandidates))].UserID
	// Обновляем список ревьюверов
	newReviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			newReviewers = append(newReviewers, newReviewerID)
		} else {
			newReviewers = append(newReviewers, reviewer)
		}
	}

	if err := s.prRepo.UpdateReviewers(ctx, prID, newReviewers); err != nil {
		return "", err
	}

	return newReviewerID, nil
}

// PickRandomReviewers выбирает N случайных активных членов команды
func (s *ReviewerAssignmentService) PickRandomReviewers(candidates []domain.User, count int) []domain.User {
	if len(candidates) <= count {
		return candidates
	}

	result := make([]domain.User, count)
	perm := rand.Perm(len(candidates))
	for i := 0; i < count; i++ {
		result[i] = candidates[perm[i]]
	}
	return result
}
