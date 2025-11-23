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
	defer it.Close()
	it.Cleanup(t)

	// Создаём команду с 5 участниками
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

	// Создаём PR (автор не должен попасть в ревьюверы)
	resp := it.Post(t, "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-reassign-1",
		"pull_request_name": "Refactoring",
		"author_id":         "author",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Run("Successfully reassign reviewer", func(t *testing.T) {
		// Предположим, изначально назначены r1 и r2
		body := map[string]any{
			"pull_request_id": "pr-reassign-1",
			"old_user_id":     "r1", // заменяем r1
		}

		resp := it.Post(t, "/pullRequest/reassign", body)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
			ReplacedBy string `json:"replaced_by"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

		// r1 должен быть заменён на кого-то из оставшихся активных (r3 или r4)
		assert.NotContains(t, result.PR.AssignedReviewers, "r1")
		assert.Contains(t, []string{"r3", "r4"}, result.ReplacedBy)
		assert.Len(t, result.PR.AssignedReviewers, 2)
		assert.NotContains(t, result.PR.AssignedReviewers, "author")
	})

	t.Run("Reassign on merged PR → 409 PR_MERGED", func(t *testing.T) {
		// Сначала мержим PR
		it.Post(t, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-reassign-1"})

		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-reassign-1",
			"old_user_id":     "r2",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var errResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Equal(t, "PR_MERGED", errResp.Error.Code)
	})

	t.Run("Reassign non-assigned reviewer → 409 NOT_ASSIGNED", func(t *testing.T) {
		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-reassign-1",
			"old_user_id":     "r4", // r4 не был назначен
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
		it.Post(t, "/pullRequest/create", map[string]any{
			"pull_request_id":   "pr-no-candidate",
			"pull_request_name": "Hotfix",
			"author_id":         "author",
		})

		// Деактивируем всех, кроме автора и одного ревьювера
		for _, id := range []string{"r2", "r3", "r4"} {
			it.Post(t, "/users/setIsActive", map[string]any{"user_id": id, "is_active": false})
		}

		resp := it.Post(t, "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-no-candidate",
			"old_user_id":     "r1", // пытаемся заменить последнего активного
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
