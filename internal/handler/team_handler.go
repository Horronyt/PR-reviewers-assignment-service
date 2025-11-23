package handler

import (
	"encoding/json"
	"net/http"

	"github.com/horronyt/pr-reviewers-service/internal/domain"
	"github.com/horronyt/pr-reviewers-service/internal/service"
)

// ErrorResponse структура для ошибки
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TeamHandler обработчик для операций с командами
type TeamHandler struct {
	teamService *service.TeamService
}

// NewTeamHandler создает новый handler
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// AddTeam обработчик POST /team/add
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Конвертируем в domain model
	team := &domain.Team{
		TeamName: req.TeamName,
		Members:  make([]domain.User, len(req.Members)),
	}

	for i, member := range req.Members {
		team.Members[i] = domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	result, err := h.teamService.CreateTeam(r.Context(), team)
	if err != nil {
		if domErr, ok := err.(domain.DomainError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
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
		"team": map[string]interface{}{
			"team_name": result.TeamName,
			"members": func() []interface{} {
				members := make([]interface{}, len(result.Members))
				for i, m := range result.Members {
					members[i] = map[string]interface{}{
						"user_id":   m.UserID,
						"username":  m.Username,
						"is_active": m.IsActive,
					}
				}
				return members
			}(),
		},
	})
}

// GetTeam обработчик GET /team/get
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name parameter required", http.StatusBadRequest)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
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
		"team_name": team.TeamName,
		"members": func() []interface{} {
			members := make([]interface{}, len(team.Members))
			for i, m := range team.Members {
				members[i] = map[string]interface{}{
					"user_id":   m.UserID,
					"username":  m.Username,
					"is_active": m.IsActive,
				}
			}
			return members
		}(),
	})
}
