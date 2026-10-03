package user_anime_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/events"
	svc "github.com/weeb-vip/list-service/internal/services/user_anime"
)

func str(s string) *string   { return &s }
func num(f float64) *float64 { return &f }
func n(i int) *int           { return &i }

// fakeRepo implements the two transactional methods the service uses and
// records what it was asked. Embedding the interface leaves the rest nil so
// an unexpected call fails loudly.
type fakeRepo struct {
	user_anime.UserAnimeRepositoryImpl
	existing *user_anime.UserAnime
	upserted *user_anime.UserAnime
	deleted  *user_anime.UserAnime
	err      error
}

func (f *fakeRepo) UpsertTx(_ context.Context, _ *gorm.DB, entity *user_anime.UserAnime) (*user_anime.UpsertResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.upserted = entity
	if f.existing == nil {
		entity.ID = "new-row"

		return &user_anime.UpsertResult{Entity: entity, Created: true}, nil
	}
	entity.ID = f.existing.ID

	return &user_anime.UpsertResult{Entity: entity, Previous: f.existing}, nil
}

func (f *fakeRepo) FindByUserIdAndAnimeId(_ context.Context, _ string, _ string) (*user_anime.UserAnime, error) {
	return f.existing, f.err
}

func (f *fakeRepo) DeleteTx(_ context.Context, _ *gorm.DB, entity *user_anime.UserAnime) error {
	f.deleted = entity

	return nil
}

type fakeWriter struct {
	written []*events.Activity
	err     error
}

func (w *fakeWriter) Write(_ context.Context, _ *gorm.DB, a *events.Activity) error {
	if w.err != nil {
		return w.err
	}
	w.written = append(w.written, a)

	return nil
}

// txRunner hands the function a nil transaction and reports whether it ran.
func txRunner(ran *bool) svc.TxRunner {
	return svc.TxRunnerFunc(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
		*ran = true

		return fn(nil)
	})
}

func existing(status string, score *float64, episodes int) *user_anime.UserAnime {
	return &user_anime.UserAnime{ID: "row-1", UserID: str("user_a"), AnimeID: str("anime-1"), Status: str(status), Score: score, Episodes: n(episodes)}
}

func TestUpsertEmitsTheRightEvent(t *testing.T) {
	cases := []struct {
		name     string
		existing *user_anime.UserAnime
		input    svc.UserAnime
		want     string
		wantPrev *string
	}{
		{
			name:  "first add",
			input: svc.UserAnime{UserID: "user_a", AnimeID: "anime-1", Status: ptrStatus(svc.Watching)},
			want:  events.AnimeAdded,
		},
		{
			name:     "status change",
			existing: existing("watching", nil, 3),
			input:    svc.UserAnime{UserID: "user_a", AnimeID: "anime-1", Status: ptrStatus(svc.Completed), Episodes: n(12)},
			want:     events.AnimeStatusChanged,
			wantPrev: str("watching"),
		},
		{
			name:     "score change",
			existing: existing("watching", num(7), 3),
			input:    svc.UserAnime{UserID: "user_a", AnimeID: "anime-1", Status: ptrStatus(svc.Watching), Score: num(9), Episodes: n(3)},
			want:     events.AnimeScored,
		},
		{
			name:     "progress only is silent",
			existing: existing("watching", num(7), 3),
			input:    svc.UserAnime{UserID: "user_a", AnimeID: "anime-1", Status: ptrStatus(svc.Watching), Score: num(7), Episodes: n(4)},
			want:     "",
		},
		{
			name:     "tags only is silent",
			existing: existing("watching", nil, 3),
			input:    svc.UserAnime{UserID: "user_a", AnimeID: "anime-1", Status: ptrStatus(svc.Watching), Episodes: n(3), Tags: []string{"favourite"}},
			want:     "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{existing: tc.existing}
			writer := &fakeWriter{}
			ran := false
			s := svc.NewUserAnimeService(repo, txRunner(&ran), writer)

			saved, err := s.Upsert(context.Background(), &tc.input)
			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.True(t, ran, "the upsert and the event share a transaction")
			assert.Equal(t, "anime-1", *repo.upserted.AnimeID)

			if tc.want == "" {
				assert.Empty(t, writer.written)

				return
			}
			require.Len(t, writer.written, 1)
			ev := writer.written[0]
			assert.Equal(t, tc.want, ev.Type)
			assert.Equal(t, "user_a", ev.UserID)
			assert.Equal(t, "anime-1", *ev.AnimeID)
			assert.Nil(t, ev.WorkID)
			assert.Equal(t, tc.wantPrev, ev.PreviousStatus)
			if tc.input.Status != nil {
				assert.Equal(t, string(*tc.input.Status), *ev.Status)
			}
		})
	}
}

func TestUpsertFailuresRollUpToTheCaller(t *testing.T) {
	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("db down")
		ran := false
		s := svc.NewUserAnimeService(&fakeRepo{err: boom}, txRunner(&ran), &fakeWriter{})
		_, err := s.Upsert(context.Background(), &svc.UserAnime{UserID: "user_a", AnimeID: "anime-1"})
		require.ErrorIs(t, err, boom)
	})

	t.Run("event write error fails the upsert so the transaction rolls back", func(t *testing.T) {
		boom := errors.New("outbox full")
		ran := false
		s := svc.NewUserAnimeService(&fakeRepo{}, txRunner(&ran), &fakeWriter{err: boom})
		_, err := s.Upsert(context.Background(), &svc.UserAnime{UserID: "user_a", AnimeID: "anime-1"})
		require.ErrorIs(t, err, boom)
	})
}

func TestDeleteEmitsRemoved(t *testing.T) {
	t.Run("own row", func(t *testing.T) {
		repo := &fakeRepo{existing: existing("completed", nil, 12)}
		writer := &fakeWriter{}
		ran := false
		s := svc.NewUserAnimeService(repo, txRunner(&ran), writer)

		require.NoError(t, s.Delete(context.Background(), "user_a", "anime-1"))
		require.NotNil(t, repo.deleted)
		require.Len(t, writer.written, 1)
		assert.Equal(t, events.AnimeRemoved, writer.written[0].Type)
		assert.Equal(t, "completed", *writer.written[0].PreviousStatus)
	})

	t.Run("someone else's row is left alone", func(t *testing.T) {
		repo := &fakeRepo{existing: existing("completed", nil, 12)}
		writer := &fakeWriter{}
		ran := false
		s := svc.NewUserAnimeService(repo, txRunner(&ran), writer)

		require.NoError(t, s.Delete(context.Background(), "user_b", "anime-1"))
		assert.Nil(t, repo.deleted)
		assert.Empty(t, writer.written)
		assert.False(t, ran)
	})

	t.Run("missing row is a no-op", func(t *testing.T) {
		repo := &fakeRepo{}
		writer := &fakeWriter{}
		ran := false
		s := svc.NewUserAnimeService(repo, txRunner(&ran), writer)

		require.NoError(t, s.Delete(context.Background(), "user_a", "anime-1"))
		assert.Empty(t, writer.written)
	})
}

func ptrStatus(s svc.UserAnimeStatus) *svc.UserAnimeStatus { return &s }
