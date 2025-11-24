// tests/pr_merge_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRMerge(t *testing.T) {
	it := New(t)

	it.Post(t, "/team/add", map[string]any{
		"team_name": "infra",
		"members":   []map[string]any{{"user_id": "u1", "username": "Alice", "is_active": true}},
	})
	it.Post(t, "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-999",
		"pull_request_name": "CI fix",
		"author_id":         "u1",
	})

	t.Run("Merge PR → MERGED", func(t *testing.T) {
		resp := it.Post(t, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-999"})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			PR struct {
				Status   string  `json:"status"`
				MergedAt *string `json:"mergedAt"`
			} `json:"pr"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, "MERGED", result.PR.Status)
		assert.NotNil(t, result.PR.MergedAt)
	})

	t.Run("Idempotent merge → still 200", func(t *testing.T) {
		resp := it.Post(t, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-999"})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
