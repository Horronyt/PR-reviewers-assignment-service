// tests/pr_create_test.go
package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// Обернем общую логику настройки в отдельную функцию, чтобы не дублировать код
func setupTest(t *testing.T, it *IntegrationTest) {
	it.Cleanup(t) // Очистка перед каждым тестом/подтестом

	// Общая настройка данных, необходимая для обоих подтестов
	it.Post(t, "/team/add", map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "author", "username": "Author", "is_active": true},
			{"user_id": "r1", "username": "R1", "is_active": true},
			{"user_id": "r2", "username": "R2", "is_active": true},
			{"user_id": "r3", "username": "R3", "is_active": true},
		},
	})
}

func TestPRCreation(t *testing.T) {
	// New и Close вызываются только один раз для всего набора тестов/подтестов
	it := New(t)
	defer it.Close()

	t.Run("Creates PR and assigns 2 reviewers", func(t *testing.T) {
		// Настройка базы данных только для этого подтеста
		setupTest(t, it)

		resp := it.Post(t, "/pullRequest/create", map[string]any{
			"pull_request_id":   "pr-1001",
			"pull_request_name": " Dark mode",
			"author_id":         "author",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
				Status            string   `json:"status"`
			} `json:"pr"`
		}

		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Len(t, result.PR.AssignedReviewers, 2)
		assert.NotEqual(t, "author", result.PR.AssignedReviewers[0])
		assert.Equal(t, "OPEN", result.PR.Status)
	})
}
