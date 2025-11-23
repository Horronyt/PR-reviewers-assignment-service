package handler

import (
	"encoding/json"
	"net/http"

	"github.com/horronyt/pr-reviewers-service/internal/domain"
	"github.com/horronyt/pr-reviewers-service/internal/service"
)

// PRHandler обработчик для операций с PR
type PRHandler struct {
	prService   *service.PRService
	userService *service.UserService
}

// NewPRHandler создает новый handler
func NewPRHandler(prService *service.PRService, userService *service.UserService) *PRHandler {
	return &PRHandler{
		prService:   prService,
		userService: userService,
	}
}

// CreatePR обработчик POST /pullRequest/create
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if domErr, ok := err.(domain.DomainError); ok {
			w.Header().Set("Content-Type", "application/json")
			statusCode := http.StatusBadRequest
			if domErr.Code == domain.ErrorCodeNotFound {
				statusCode = http.StatusNotFound
			} else if domErr.Code == domain.ErrorCodePRExists {
				statusCode = http.StatusConflict
			}
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{Code: string(domErr.Code), Message: domErr.Message},
			})
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": map[string]interface{}{
			"pull_request_id":    pr.PullRequestID,
			"pull_request_name":  pr.PullRequestName,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": pr.AssignedReviewers,
			"createdAt":          pr.CreatedAt,
			"mergedAt":           pr.MergedAt,
		},
	})
}

// MergePR обработчик POST /pullRequest/merge
func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		if domErr, ok := err.(domain.DomainError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{Code: string(domErr.Code), Message: domErr.Message},
			})
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": map[string]interface{}{
			"pull_request_id":    pr.PullRequestID,
			"pull_request_name":  pr.PullRequestName,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": pr.AssignedReviewers,
			"createdAt":          pr.CreatedAt,
			"mergedAt":           pr.MergedAt,
		},
	})
}

// ReassignReviewer обработчик POST /pullRequest/reassign
func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if domErr, ok := err.(domain.DomainError); ok {
			w.Header().Set("Content-Type", "application/json")
			statusCode := http.StatusBadRequest
			if domErr.Code == domain.ErrorCodeNotFound {
				statusCode = http.StatusNotFound
			} else if domErr.Code == domain.ErrorCodePRMerged ||
				domErr.Code == domain.ErrorCodeNotAssigned ||
				domErr.Code == domain.ErrorCodeNoCandidate {
				statusCode = http.StatusConflict
			}
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{Code: string(domErr.Code), Message: domErr.Message},
			})
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": map[string]interface{}{
			"pull_request_id":    pr.PullRequestID,
			"pull_request_name":  pr.PullRequestName,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": pr.AssignedReviewers,
			"createdAt":          pr.CreatedAt,
			"mergedAt":           pr.MergedAt,
		},
		"replaced_by": newReviewerID,
	})
}
