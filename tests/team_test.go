// tests/team_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamFlow(t *testing.T) {
	it := New(t)

	t.Run("Create team successfully", func(t *testing.T) {
		body := map[string]any{
			"team_name": "payments",
			"members": []map[string]any{
				{"user_id": "u1", "username": "Alice", "is_active": true},
				{"user_id": "u2", "username": "Bob", "is_active": true},
			},
		}

		resp := it.Post(t, "/team/add", body)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result struct {
			Team struct {
				TeamName string `json:"team_name"`
				Members  []struct {
					UserID   string `json:"user_id"`
					Username string `json:"username"`
				} `json:"members"`
			} `json:"team"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, "payments", result.Team.TeamName)
		assert.Len(t, result.Team.Members, 2)
	})

	t.Run("Get existing team", func(t *testing.T) {
		it.Post(t, "/team/add", map[string]any{
			"team_name": "mobile",
			"members":   []map[string]any{{"user_id": "u10", "username": "Mike", "is_active": true}},
		})

		resp := it.Get(t, "/team/get?team_name=mobile")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Create duplicate team → 400", func(t *testing.T) {
		it.Post(t, "/team/add", map[string]any{"team_name": "dup", "members": []map[string]any{{"user_id": "u1", "username": "X", "is_active": true}}})

		resp := it.Post(t, "/team/add", map[string]any{"team_name": "dup", "members": []map[string]any{{"user_id": "u2", "username": "Y", "is_active": true}}})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
