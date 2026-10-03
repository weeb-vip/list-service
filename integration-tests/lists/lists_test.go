//go:build integration

package lists

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gqlResponse struct {
	Data   json.RawMessage   `json:"data"`
	Errors []json.RawMessage `json:"errors"`
}

func query(t *testing.T, userID, operation string, variables map[string]any) *gqlResponse {
	t.Helper()
	body, err := json.Marshal(map[string]any{"query": operation, "variables": variables})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/graphql", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("x-user-id", userID)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var out gqlResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))

	return &out
}

func mustData(t *testing.T, userID, operation string, variables map[string]any, into any) {
	t.Helper()
	resp := query(t, userID, operation, variables)
	require.Empty(t, resp.Errors, "unexpected GraphQL errors: %s", resp.Errors)
	if into != nil {
		require.NoError(t, json.Unmarshal(resp.Data, into))
	}
}

type activity struct {
	Type           string   `json:"type"`
	UserID         string   `json:"user_id"`
	AnimeID        *string  `json:"anime_id"`
	WorkID         *string  `json:"work_id"`
	Status         *string  `json:"status"`
	PreviousStatus *string  `json:"previous_status"`
	Score          *float64 `json:"score"`
	ID             string   `json:"id"`
	OccurredAt     string   `json:"occurred_at"`
}

// outboxFor returns the activity events written for a user, oldest first.
func outboxFor(t *testing.T, userID string) []activity {
	t.Helper()
	var rows []struct {
		ID      string
		Payload []byte
	}
	require.NoError(t, database.Raw(`
		SELECT id::text AS id, payload::text AS payload FROM outbox_events
		WHERE subject = 'user-activity' AND payload->>'user_id' = ? AND published_at IS NULL
		ORDER BY created_at, id`, userID).Scan(&rows).Error)
	out := make([]activity, 0, len(rows))
	for _, row := range rows {
		var a activity
		require.NoError(t, json.Unmarshal(row.Payload, &a))
		assert.Equal(t, row.ID, a.ID, "outbox row id and payload id must match")
		out = append(out, a)
	}

	return out
}

func newUserID(t *testing.T) string {
	id := fmt.Sprintf("user_e2e_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		database.Exec("DELETE FROM user_anime WHERE user_id = ?", id)
		database.Exec("DELETE FROM user_work WHERE user_id = ?", id)
		database.Exec("DELETE FROM outbox_events WHERE payload->>'user_id' = ?", id)
	})

	return id
}

const addAnime = `mutation($input: UserAnimeInput!) { AddAnime(input: $input) { id status episodes score } }`

func TestAnimeListChangesWriteActivityEvents(t *testing.T) {
	user := newUserID(t)
	anime := uuid.NewString()

	// Add as plan-to-watch.
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "PLANTOWATCH"}}, nil)
	// Start watching.
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "WATCHING", "episodes": 1}}, nil)
	// Progress only: the detail page saves episodes through the same mutation.
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "WATCHING", "episodes": 2}}, nil)
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "WATCHING", "episodes": 3, "tags": []string{"hype"}}}, nil)
	// Score it.
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "WATCHING", "episodes": 3, "score": 8.5}}, nil)
	// Finish it, same score.
	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "COMPLETED", "episodes": 12, "score": 8.5}}, nil)
	// Remove it.
	var del struct{ DeleteAnime bool }
	mustData(t, user, `mutation($id: ID!) { DeleteAnime(id: $id) }`, map[string]any{"id": anime}, &del)
	assert.True(t, del.DeleteAnime)

	got := outboxFor(t, user)
	types := make([]string, 0, len(got))
	for _, a := range got {
		types = append(types, a.Type)
		assert.Equal(t, user, a.UserID)
		require.NotNil(t, a.AnimeID)
		assert.Equal(t, anime, *a.AnimeID)
		assert.Nil(t, a.WorkID)
		assert.NotEmpty(t, a.OccurredAt)
	}
	assert.Equal(t, []string{"anime.added", "anime.status_changed", "anime.scored", "anime.status_changed", "anime.removed"}, types)

	assert.Equal(t, "PLANTOWATCH", *got[0].Status)
	assert.Equal(t, "PLANTOWATCH", *got[1].PreviousStatus)
	assert.Equal(t, "WATCHING", *got[1].Status)
	assert.Equal(t, 8.5, *got[2].Score)
	assert.Equal(t, "WATCHING", *got[3].PreviousStatus)
	assert.Equal(t, "COMPLETED", *got[3].Status)
	assert.Equal(t, "COMPLETED", *got[4].PreviousStatus)

	// Deleting again is an error (the row is gone) and writes nothing.
	resp := query(t, user, `mutation($id: ID!) { DeleteAnime(id: $id) }`, map[string]any{"id": anime})
	assert.NotEmpty(t, resp.Errors)
	assert.Len(t, outboxFor(t, user), 5)
}

func TestWorkListChangesWriteActivityEvents(t *testing.T) {
	user := newUserID(t)
	work := uuid.NewString()
	const addWork = `mutation($input: UserWorkInput!) { AddWork(input: $input) { id status chapters } }`
	const updateWork = `mutation($input: UserWorkInput!) { UpdateWork(input: $input) { id status chapters } }`

	mustData(t, user, addWork, map[string]any{"input": map[string]any{"workID": work, "status": "PLANTOREAD"}}, nil)
	mustData(t, user, updateWork, map[string]any{"input": map[string]any{"workID": work, "status": "READING", "chapters": 5}}, nil)
	mustData(t, user, updateWork, map[string]any{"input": map[string]any{"workID": work, "status": "READING", "chapters": 9}}, nil)
	var del struct{ DeleteWork bool }
	mustData(t, user, `mutation($id: ID!) { DeleteWork(id: $id) }`, map[string]any{"id": work}, &del)
	assert.True(t, del.DeleteWork)

	got := outboxFor(t, user)
	types := make([]string, 0, len(got))
	for _, a := range got {
		types = append(types, a.Type)
		require.NotNil(t, a.WorkID)
		assert.Equal(t, work, *a.WorkID)
		assert.Nil(t, a.AnimeID)
	}
	assert.Equal(t, []string{"work.added", "work.status_changed", "work.removed"}, types)
	assert.Equal(t, "PLANTOREAD", *got[1].PreviousStatus)
	assert.Equal(t, "READING", *got[1].Status)
}

func TestEpisodeProgressIsSilent(t *testing.T) {
	user := newUserID(t)
	anime := uuid.NewString()

	mustData(t, user, addAnime, map[string]any{"input": map[string]any{"animeID": anime, "status": "WATCHING"}}, nil)
	resp := query(t, user, `mutation($input: MarkEpisodeInput!) { MarkEpisodeWatched(input: $input) { episodeNumber } }`,
		map[string]any{"input": map[string]any{"animeID": anime, "episodeNumber": 1}})
	require.Empty(t, resp.Errors, "%s", resp.Errors)

	got := outboxFor(t, user)
	require.Len(t, got, 1, "only the add is announced")
	assert.Equal(t, "anime.added", got[0].Type)
}

func TestAnonymousListChangeIsRejectedAndSilent(t *testing.T) {
	resp := query(t, "", addAnime, map[string]any{"input": map[string]any{"animeID": uuid.NewString(), "status": "WATCHING"}})
	assert.NotEmpty(t, resp.Errors)
}
