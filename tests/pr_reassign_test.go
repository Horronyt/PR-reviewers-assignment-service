// tests/pr_reassign_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRReassign(t *testing.T) {
	it := New(t)

	// Создаём команду
	it.Post(t, "/team/add", map[string]any{
		"team_name": "core",
		"members": []map[string]any{
			{"user_id": "author", "username": "Author", "is_active": true},
			{"user_id": "r1", "username": "Alice", "is_active": true},
			{"user_id": "r2", "username": "Bob", "is_active": true},
			{"user_id": "r3", "username": "Charlie", "is_active": true},
			{"user_id": "r4", "username": "Dave", "is_active": true},
		},
	})

	// === Создаём PR и сразу получаем назначенных ревьюверов ===
	var initialPR struct {
		PR struct {
			AssignedReviewers []string `json:"assigned_reviewers"`
		} `json:"pr"`
	}

	resp := it.Post(t, "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-reassign-1",
		"pull_request_name": "Refactoring",
		"author_id":         "author",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&initialPR))

	initialReviewers := initialPR.PR.AssignedReviewers
	require.Len(t, initialReviewers, 2, "при создании PR должно быть назначено 2 ревьювера")
	require.NotContains(t, initialReviewers, "author", "автор не должен быть ревьювером")

	t.Run("Successfully reassign reviewer", func(t *testing.T) {
		// Берём любого из текущих ревьюверов
		oldReviewerID := initialReviewers[0]

		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-reassign-1",
			"old_user_id":     oldReviewerID,
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
			ReplacedBy string `json:"replaced_by"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

		newReviewers := result.PR.AssignedReviewers

		assert.NotContains(t, newReviewers, oldReviewerID, "старый ревьювер должен исчезнуть")
		assert.Contains(t, newReviewers, result.ReplacedBy, "новый ревьювер должен быть в списке")
		assert.Len(t, newReviewers, 2, "количество ревьюверов должно остаться 2")
		assert.NotContains(t, newReviewers, "author", "автор не должен попасть в ревьюверы")
		assert.NotEqual(t, oldReviewerID, result.ReplacedBy, "новый ≠ старый")
	})

	t.Run("Reassign on merged PR → 409 PR_MERGED", func(t *testing.T) {
		it.Post(t, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-reassign-1"})

		oldReviewerID := initialReviewers[0] // всё ещё валиден

		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-reassign-1",
			"old_user_id":     oldReviewerID,
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var errResp struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Equal(t, "PR_MERGED", errResp.Error.Code)
	})

	t.Run("Reassign non-assigned reviewer → 409 NOT_ASSIGNED", func(t *testing.T) {
		// Создаём новый PR специально для этого кейса
		respCreate := it.Post(t, "/pullRequest/create", map[string]any{
			"pull_request_id":   "pr-not-assigned-test",
			"pull_request_name": "Test PR",
			"author_id":         "author",
		})
		defer respCreate.Body.Close()

		var tempPR struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
		}
		require.NoError(t, json.NewDecoder(respCreate.Body).Decode(&tempPR))
		current := tempPR.PR.AssignedReviewers

		// Находим НЕ назначенного
		all := map[string]bool{"r1": true, "r2": true, "r3": true, "r4": true}
		for _, r := range current {
			delete(all, r)
		}
		var nonAssigned string
		for id := range all {
			nonAssigned = id
			break
		}
		require.NotEmpty(t, nonAssigned)

		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-not-assigned-test",
			"old_user_id":     nonAssigned,
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var errResp struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Equal(t, "NOT_ASSIGNED", errResp.Error.Code)
	})

	t.Run("No active candidate → 409 NO_CANDIDATE", func(t *testing.T) {
		// Создаём новый PR
		var prResp struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
		}

		resp := it.Post(t, "/pullRequest/create", map[string]any{
			"pull_request_id":   "pr-no-candidate",
			"pull_request_name": "Hotfix",
			"author_id":         "author",
		})
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&prResp))

		currentReviewers := prResp.PR.AssignedReviewers
		require.Len(t, currentReviewers, 2)

		// Деактивируем ВСЕХ, кроме ОДНОГО из текущих ревьюверов
		// Например, оставляем активным только первого назначенного
		keepActive := currentReviewers[0]

		// Деактивируем всех остальных из команды (r1–r4), кроме keepActive и author
		for _, id := range []string{"r1", "r2", "r3", "r4"} {
			if id != keepActive {
				it.Post(t, "/users/setIsActive", map[string]any{
					"user_id":   id,
					"is_active": false,
				})
			}
		}

		// Теперь пытаемся заменить единственного активного ревьювера
		resp = it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-no-candidate",
			"old_user_id":     keepActive, // ← 100% назначен и активен
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var errResp struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Equal(t, "NO_CANDIDATE", errResp.Error.Code)
	})

	t.Run("PR not found → 404", func(t *testing.T) {
		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-ghost",
			"old_user_id":     "r1",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
