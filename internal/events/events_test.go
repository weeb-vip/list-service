package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func str(s string) *string   { return &s }
func num(f float64) *float64 { return &f }

func TestDecide(t *testing.T) {
	cases := []struct {
		name                  string
		created               bool
		prevStatus, newStatus *string
		prevScore, newScore   *float64
		want                  string
	}{
		{"new row is added", true, nil, str("watching"), nil, nil, AnimeAdded},
		{"new row with score is still added", true, nil, str("watching"), nil, num(8), AnimeAdded},
		{"status change", false, str("watching"), str("completed"), nil, nil, AnimeStatusChanged},
		{"status set from nothing", false, nil, str("watching"), nil, nil, AnimeStatusChanged},
		{"status cleared", false, str("watching"), nil, nil, nil, AnimeStatusChanged},
		{"status and score change reports the status", false, str("watching"), str("completed"), num(7), num(9), AnimeStatusChanged},
		{"score change", false, str("watching"), str("watching"), num(7), num(8), AnimeScored},
		{"score set from nothing", false, str("watching"), str("watching"), nil, num(8), AnimeScored},
		{"nothing relevant changed", false, str("watching"), str("watching"), num(7), num(7), ""},
		{"progress-only save", false, str("watching"), str("watching"), nil, nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Decide(tc.created, tc.prevStatus, tc.newStatus, tc.prevScore, tc.newScore, AnimeAdded, AnimeStatusChanged, AnimeScored)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("works use their own names", func(t *testing.T) {
		assert.Equal(t, WorkAdded, Decide(true, nil, nil, nil, nil, WorkAdded, WorkStatusChanged, WorkScored))
		assert.Equal(t, WorkScored, Decide(false, str("READING"), str("READING"), nil, num(5), WorkAdded, WorkStatusChanged, WorkScored))
	})
}
