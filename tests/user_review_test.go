// tests/user_review_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserReviews(t *testing.T) {
	it := New(t)
	defer it.Close()

	it.Post(t, "/team/add", map[string]any{
		"team_name": "mobile",
		"members": []map[string]any{
			{"user_id": "author", "username": "A", "is_active": true},
			{"user_id": "reviewer", "username": "R", "is_active": true},
		},
	})

	it.Post(t, "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-x",
		"pull_request_name": "Feature",
		"author_id":         "author",
	})

	t.Run("User sees assigned PRs", func(t *testing.T) {
		resp := it.Get(t, "/users/getReview?user_id=reviewer")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			UserID       string `json:"user_id"`
			PullRequests []struct {
				PullRequestID string `json:"pull_request_id"`
			} `json:"pull_requests"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, "reviewer", result.UserID)
		assert.Len(t, result.PullRequests, 1)
		assert.Equal(t, "pr-x", result.PullRequests[0].PullRequestID)
	})
}
