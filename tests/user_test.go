// tests/user_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUserActive(t *testing.T) {
	it := New(t)
	defer it.Close()

	it.Post(t, "/team/add", map[string]any{
		"team_name": "frontend",
		"members":   []map[string]any{{"user_id": "u1", "username": "Alice", "is_active": true}},
	})

	t.Run("Deactivate user", func(t *testing.T) {
		resp := it.Post(t, "/users/setIsActive", map[string]any{"user_id": "u1", "is_active": false})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			User struct {
				IsActive bool `json:"is_active"`
			} `json:"user"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.False(t, result.User.IsActive)
	})

	t.Run("Non-existing user → 404", func(t *testing.T) {
		resp := it.Post(t, "/users/setIsActive", map[string]any{"user_id": "ghost", "is_active": true})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
