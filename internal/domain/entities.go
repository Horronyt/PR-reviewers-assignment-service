package domain

import "time"

type User struct {
	UserID    string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Team struct {
	TeamName  string
	Members   []User
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

// NewError создает новую доменную ошибку
func NewError(code ErrorCode, message string) DomainError {
	return DomainError{Code: code, Message: message}
}
