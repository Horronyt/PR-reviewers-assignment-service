package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Horronyt/PR-reviewers-assignment-service/internal/service"
)

// StatsHandler обработчик для получения статистики
type StatsHandler struct {
	userService *service.UserService
}

// NewStatsHandler создает новый handler для статистики
func NewStatsHandler(userService *service.UserService) *StatsHandler {
	return &StatsHandler{
		userService: userService,
	}
}

// GetStats обработчик GET /stats
// Возвращает статистику по ревьюверам и PR
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем статистику по ревьюверам
	reviewerStats, err := h.userService.GetReviewerStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get reviewer stats", http.StatusInternalServerError)
		return
	}

	// Получаем статистику по PR
	prStats, err := h.userService.GetPRStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get PR stats", http.StatusInternalServerError)
		return
	}

	// Преобразуем reviewer stats в JSON-friendly формат
	reviewerStatsData := make([]map[string]interface{}, len(reviewerStats))
	for i, stat := range reviewerStats {
		reviewerStatsData[i] = map[string]interface{}{
			"user_id":          stat.UserID,
			"assignment_count": stat.AssignmentCount,
		}
	}

	// Формируем ответ
	response := map[string]interface{}{
		"reviewer_stats": reviewerStatsData,
		"pr_stats":       prStats,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetReviewerStats вспомогательный метод для получения статистики ревьюверов
// Может быть использован для более детальной информации
func (h *StatsHandler) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.userService.GetReviewerStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get reviewer stats", http.StatusInternalServerError)
		return
	}

	// Преобразуем в JSON-friendly формат
	statsData := make([]map[string]interface{}, len(stats))
	for i, stat := range stats {
		statsData[i] = map[string]interface{}{
			"user_id":          stat.UserID,
			"assignment_count": stat.AssignmentCount,
		}
	}

	response := map[string]interface{}{
		"reviewer_stats":  statsData,
		"total_reviewers": len(stats),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetPRStats вспомогательный метод для получения статистики PR
func (h *StatsHandler) GetPRStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.userService.GetPRStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get PR stats", http.StatusInternalServerError)
		return
	}

	// Вычисляем общее количество PR
	totalPRs := 0
	for _, count := range stats {
		totalPRs += count
	}

	response := map[string]interface{}{
		"pr_stats":   stats,
		"total_prs":  totalPRs,
		"open_prs":   stats["OPEN"],
		"merged_prs": stats["MERGED"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
