package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Horronyt/PR-reviewers-assignment-service/internal/domain"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/service"
)

// UserHandler обработчик для операций с пользователями
type UserHandler struct {
	userService *service.UserService
	prService   *service.PRService
}

// NewUserHandler создает новый handler
func NewUserHandler(userService *service.UserService, prService *service.PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

// SetActive обработчик POST /users/setIsActive
func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.SetActive(r.Context(), req.UserID, req.IsActive)
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
		"user": map[string]interface{}{
			"user_id":   user.UserID,
			"username":  user.Username,
			"team_name": user.TeamName,
			"is_active": user.IsActive,
		},
	})
}

// GetReview обработчик GET /users/getReview
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	prs, err := h.prService.GetReviewsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	prList := make([]interface{}, len(prs))
	for i, pr := range prs {
		prList[i] = map[string]interface{}{
			"pull_request_id":   pr.PullRequestID,
			"pull_request_name": pr.PullRequestName,
			"author_id":         pr.AuthorID,
			"status":            pr.Status,
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prList,
	})
}
