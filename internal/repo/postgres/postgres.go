// internal/repo/postgres/postgres.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ======================== USER REPOSITORY ========================

func (r *Repository) CreateOrUpdateUser(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id) DO UPDATE
        SET username = $2, team_name = $3, is_active = $4, updated_at = $6
    `
	now := time.Now()
	_, err := r.db.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive, now, now)
	return err
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at FROM users WHERE user_id = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&u.UserID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at FROM users WHERE team_name = $1 ORDER BY user_id`
	return r.scanUsers(ctx, query, teamName)
}

func (r *Repository) GetActiveUsers(ctx context.Context, teamName string) ([]domain.User, error) {
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at FROM users WHERE team_name = $1 AND is_active = true ORDER BY user_id`
	return r.scanUsers(ctx, query, teamName)
}

func (r *Repository) SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	query := `
        UPDATE users
        SET is_active = $1, updated_at = $2
        WHERE user_id = $3
        RETURNING user_id, username, team_name, is_active, created_at, updated_at
    `
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, isActive, time.Now(), userID).Scan(
		&u.UserID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *Repository) GetAllUsersByIDs(ctx context.Context, userIDs []string) ([]domain.User, error) {
	if len(userIDs) == 0 {
		return []domain.User{}, nil
	}
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at FROM users WHERE user_id = ANY($1) ORDER BY user_id`
	return r.scanUsers(ctx, query, userIDs)
}

// ======================== TEAM REPOSITORY ========================

func (r *Repository) CreateTeam(ctx context.Context, team *domain.Team) error {
	query := `INSERT INTO teams (team_name, created_at, updated_at) VALUES ($1, $2, $3) ON CONFLICT (team_name) DO NOTHING`
	_, err := r.db.Exec(ctx, query, team.TeamName, time.Now(), time.Now())
	if err != nil {
		return err
	}
	var exists bool
	err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, team.TeamName).Scan(&exists)
	if err == nil && !exists {
		return domain.NewError(domain.ErrorCodeTeamExists, "team already exists")
	}
	return nil
}

func (r *Repository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	query := `SELECT team_name, created_at, updated_at FROM teams WHERE team_name = $1`
	t := &domain.Team{}
	err := r.db.QueryRow(ctx, query, teamName).Scan(&t.TeamName, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}
	t.Members, err = r.GetUsersByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, teamName).Scan(&exists)
	return exists, err
}

func (r *Repository) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	return r.GetUsersByTeam(ctx, teamName)
}

// ======================== PR REPOSITORY ========================

func (r *Repository) CreatePR(ctx context.Context, pr *domain.PullRequest) error {
	query := `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (pull_request_id) DO NOTHING
    `
	_, err := r.db.Exec(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, domain.PRStatusOpen, time.Now())
	if err != nil {
		return err
	}
	if len(pr.AssignedReviewers) > 0 {
		return r.UpdateReviewers(ctx, pr.PullRequestID, pr.AssignedReviewers)
	}
	return nil
}

func (r *Repository) GetPRByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests WHERE pull_request_id = $1
    `
	pr := &domain.PullRequest{}
	err := r.db.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("PR not found: %w", err)
	}

	rows, err := r.db.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY reviewer_id`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}
	return pr, rows.Err()
}

func (r *Repository) UpdateReviewers(ctx context.Context, prID string, reviewers []string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM pr_reviewers WHERE pull_request_id = $1`, prID)
	if err != nil {
		return err
	}
	if len(reviewers) == 0 {
		return nil
	}
	query := `INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at) VALUES ($1, $2, $3)`
	now := time.Now()
	for _, reviewerID := range reviewers {
		if _, err := r.db.Exec(ctx, query, prID, reviewerID, now); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpdatePRStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error {
	query := `UPDATE pull_requests SET status = $1, merged_at = $2 WHERE pull_request_id = $3`
	_, err := r.db.Exec(ctx, query, status, mergedAt, prID)
	return err
}

func (r *Repository) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	query := `
        SELECT DISTINCT
            pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.reviewer_id = $1
        ORDER BY pr.created_at DESC
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		pr := domain.PullRequest{}
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (r *Repository) PRExists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`, prID).Scan(&exists)
	return exists, err
}

// ======================== STATS REPOSITORY ========================

func (r *Repository) GetReviewerStats(ctx context.Context) ([]repo.ReviewerStats, error) {
	query := `SELECT reviewer_id, COUNT(*) FROM pr_reviewers GROUP BY reviewer_id ORDER BY COUNT(*) DESC, reviewer_id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []repo.ReviewerStats
	for rows.Next() {
		s := repo.ReviewerStats{}
		if err := rows.Scan(&s.UserID, &s.AssignmentCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *Repository) GetPRStats(ctx context.Context) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM pull_requests GROUP BY status`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

// ======================== ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ========================

func (r *Repository) scanUsers(ctx context.Context, query string, args ...interface{}) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u := domain.User{}
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
