package user_work_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/weeb-vip/list-service/internal/db/repositories/user_work"
	"github.com/weeb-vip/list-service/internal/events"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	svc "github.com/weeb-vip/list-service/internal/services/user_work"
)

func str(s string) *string                            { return &s }
func num(f float64) *float64                          { return &f }
func status(s svc.UserWorkStatus) *svc.UserWorkStatus { return &s }

type fakeRepo struct {
	user_work.UserWorkRepositoryImpl
	existing *user_work.UserWork
	deleted  *user_work.UserWork
}

func (f *fakeRepo) UpsertTx(_ context.Context, _ *gorm.DB, entity *user_work.UserWork) (*user_work.UpsertResult, error) {
	if f.existing == nil {
		entity.ID = "new-row"

		return &user_work.UpsertResult{Entity: entity, Created: true}, nil
	}
	entity.ID = f.existing.ID

	return &user_work.UpsertResult{Entity: entity, Previous: f.existing}, nil
}

func (f *fakeRepo) FindByUserIdAndWorkId(_ context.Context, _ string, _ string) (*user_work.UserWork, error) {
	return f.existing, nil
}

func (f *fakeRepo) DeleteTx(_ context.Context, _ *gorm.DB, entity *user_work.UserWork) error {
	f.deleted = entity

	return nil
}

type fakeWriter struct{ written []*events.Activity }

func (w *fakeWriter) Write(_ context.Context, _ *gorm.DB, a *events.Activity) error {
	w.written = append(w.written, a)

	return nil
}

var tx = user_anime.TxRunnerFunc(func(ctx context.Context, fn func(tx *gorm.DB) error) error { return fn(nil) })

func TestWorkUpsertAndDeleteEvents(t *testing.T) {
	t.Run("add then move to completed then chapters only", func(t *testing.T) {
		repo := &fakeRepo{}
		writer := &fakeWriter{}
		s := svc.NewUserWorkService(repo, tx, writer)

		_, err := s.Upsert(context.Background(), &svc.UserWork{UserID: "user_a", WorkID: "work-1", Status: status(svc.Reading)})
		require.NoError(t, err)

		repo.existing = &user_work.UserWork{ID: "row", UserID: str("user_a"), WorkID: str("work-1"), Status: str("READING")}
		_, err = s.Upsert(context.Background(), &svc.UserWork{UserID: "user_a", WorkID: "work-1", Status: status(svc.Completed)})
		require.NoError(t, err)

		repo.existing = &user_work.UserWork{ID: "row", UserID: str("user_a"), WorkID: str("work-1"), Status: str("COMPLETED")}
		chapters := 40
		_, err = s.Upsert(context.Background(), &svc.UserWork{UserID: "user_a", WorkID: "work-1", Status: status(svc.Completed), Chapters: &chapters})
		require.NoError(t, err)

		repo.existing = &user_work.UserWork{ID: "row", UserID: str("user_a"), WorkID: str("work-1"), Status: str("COMPLETED")}
		_, err = s.Upsert(context.Background(), &svc.UserWork{UserID: "user_a", WorkID: "work-1", Status: status(svc.Completed), Score: num(9)})
		require.NoError(t, err)

		require.Len(t, writer.written, 3)
		assert.Equal(t, events.WorkAdded, writer.written[0].Type)
		assert.Equal(t, events.WorkStatusChanged, writer.written[1].Type)
		assert.Equal(t, "READING", *writer.written[1].PreviousStatus)
		assert.Equal(t, "COMPLETED", *writer.written[1].Status)
		assert.Equal(t, events.WorkScored, writer.written[2].Type)
		for _, ev := range writer.written {
			assert.Equal(t, "work-1", *ev.WorkID)
			assert.Nil(t, ev.AnimeID)
		}
	})

	t.Run("delete emits work.removed with the previous status", func(t *testing.T) {
		repo := &fakeRepo{existing: &user_work.UserWork{ID: "row", UserID: str("user_a"), WorkID: str("work-1"), Status: str("READING")}}
		writer := &fakeWriter{}
		s := svc.NewUserWorkService(repo, tx, writer)

		require.NoError(t, s.Delete(context.Background(), "user_a", "work-1"))
		require.NotNil(t, repo.deleted)
		require.Len(t, writer.written, 1)
		assert.Equal(t, events.WorkRemoved, writer.written[0].Type)
		assert.Equal(t, "READING", *writer.written[0].PreviousStatus)
	})
}
