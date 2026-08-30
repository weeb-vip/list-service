package user_progress

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	metrics_lib "github.com/weeb-vip/go-metrics-lib"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/metrics"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// clauseOnConflictDoNothing skips rows that are already there.
//
// No conflict target: the unique index is partial (WHERE deleted_at IS NULL),
// and an untargeted DO NOTHING covers it without having to restate the index's
// predicate here, where it could drift out of step with the migration.
func clauseOnConflictDoNothing() clause.Expression {
	return clause.OnConflict{DoNothing: true}
}

type UserProgressRepositoryImpl interface {
	MarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int, watchedAt *time.Time) (*UserAnimeEpisode, error)
	UnmarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int) error
	WatchedEpisodes(ctx context.Context, userID string, animeID string) ([]*UserAnimeEpisode, error)
	CountWatchedEpisodes(ctx context.Context, userID string, animeID string) (int64, error)
	MarkEpisodesUpTo(ctx context.Context, userID string, animeID string, upTo int) error

	MarkChapter(ctx context.Context, userID string, workID string, chapterNumber int, readAt *time.Time) (*UserWorkChapter, error)
	UnmarkChapter(ctx context.Context, userID string, workID string, chapterNumber int) error
	ReadChapters(ctx context.Context, userID string, workID string) ([]*UserWorkChapter, error)
	CountReadChapters(ctx context.Context, userID string, workID string) (int64, error)
}

type UserProgressRepository struct {
	db *db.DB
}

func NewUserProgressRepository(db *db.DB) UserProgressRepositoryImpl {
	return &UserProgressRepository{db: db}
}

func record(startTime time.Time, table string, method metrics_lib.DatabaseMetricMethod, result metrics_lib.Result) {
	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   table,
		Method:  method,
		Result:  result,
		Env:     metrics.GetCurrentEnv(),
	})
}

const (
	episodeTable = "user_anime_episode"
	chapterTable = "user_work_chapter"
)

// MarkEpisode records an episode as watched, idempotently.
//
// Marking something already marked returns the existing row rather than
// failing: the user asked for it to be watched, and it is. Only watched_at is
// refreshed, and only when the caller supplied one.
func (r *UserProgressRepository) MarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int, watchedAt *time.Time) (*UserAnimeEpisode, error) {
	startTime := time.Now()

	var existing UserAnimeEpisode
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND anime_id = ? AND episode_number = ?", userID, animeID, episodeNumber).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}

	if err == nil {
		if watchedAt != nil && !watchedAt.Equal(existing.WatchedAt) {
			existing.WatchedAt = *watchedAt
			if err := r.db.DB.WithContext(ctx).Save(&existing).Error; err != nil {
				record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Error)
				return nil, err
			}
			record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Success)
		}

		return &existing, nil
	}

	entry := &UserAnimeEpisode{
		ID:            uuid.New().String(),
		UserID:        userID,
		AnimeID:       animeID,
		EpisodeNumber: episodeNumber,
		WatchedAt:     time.Now(),
	}
	if watchedAt != nil {
		entry.WatchedAt = *watchedAt
	}

	if err := r.db.DB.WithContext(ctx).Create(entry).Error; err != nil {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Error)
		return nil, err
	}
	record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Success)

	return entry, nil
}

// UnmarkEpisode removes the mark. Unmarking something unmarked succeeds.
func (r *UserProgressRepository) UnmarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int) error {
	startTime := time.Now()

	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND anime_id = ? AND episode_number = ?", userID, animeID, episodeNumber).
		Delete(&UserAnimeEpisode{}).Error
	if err != nil {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Error)
		return err
	}
	record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Success)

	return nil
}

func (r *UserProgressRepository) WatchedEpisodes(ctx context.Context, userID string, animeID string) ([]*UserAnimeEpisode, error) {
	startTime := time.Now()

	var episodes []*UserAnimeEpisode
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND anime_id = ?", userID, animeID).
		Order("episode_number asc").
		Find(&episodes).Error
	if err != nil {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}
	record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return episodes, nil
}

func (r *UserProgressRepository) CountWatchedEpisodes(ctx context.Context, userID string, animeID string) (int64, error) {
	startTime := time.Now()

	var total int64
	err := r.db.DB.WithContext(ctx).Model(&UserAnimeEpisode{}).
		Where("user_id = ? AND anime_id = ?", userID, animeID).
		Count(&total).Error
	if err != nil {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return 0, err
	}
	record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return total, nil
}

// MarkEpisodesUpTo backfills 1..upTo for a user who has a legacy count but no
// per-episode rows.
//
// Without this, the first episode someone ticks would take their progress from
// "9 watched" to "1 watched", because the count becomes derived from rows the
// moment any exist. Their nine episodes were real; this turns the number they
// already had into the rows it always implied.
//
// Contiguous from one, which is what a bare count can mean. A user whose
// progress was not contiguous can correct it by unticking, which is possible
// now and was not before.
func (r *UserProgressRepository) MarkEpisodesUpTo(ctx context.Context, userID string, animeID string, upTo int) error {
	startTime := time.Now()

	if upTo <= 0 {
		return nil
	}

	entries := make([]*UserAnimeEpisode, 0, upTo)
	now := time.Now()
	for number := 1; number <= upTo; number++ {
		entries = append(entries, &UserAnimeEpisode{
			ID:            uuid.New().String(),
			UserID:        userID,
			AnimeID:       animeID,
			EpisodeNumber: number,
			WatchedAt:     now,
		})
	}

	// Ignore rows that already exist rather than failing the batch: this runs
	// alongside a mark, and the episode being marked may already be among them.
	err := r.db.DB.WithContext(ctx).
		Clauses(clauseOnConflictDoNothing()).
		Create(&entries).Error
	if err != nil {
		record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Error)
		return err
	}
	record(startTime, episodeTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Success)

	return nil
}

func (r *UserProgressRepository) MarkChapter(ctx context.Context, userID string, workID string, chapterNumber int, readAt *time.Time) (*UserWorkChapter, error) {
	startTime := time.Now()

	var existing UserWorkChapter
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND work_id = ? AND chapter_number = ?", userID, workID, chapterNumber).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}

	if err == nil {
		if readAt != nil && !readAt.Equal(existing.ReadAt) {
			existing.ReadAt = *readAt
			if err := r.db.DB.WithContext(ctx).Save(&existing).Error; err != nil {
				record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Error)
				return nil, err
			}
			record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Success)
		}

		return &existing, nil
	}

	entry := &UserWorkChapter{
		ID:            uuid.New().String(),
		UserID:        userID,
		WorkID:        workID,
		ChapterNumber: chapterNumber,
		ReadAt:        time.Now(),
	}
	if readAt != nil {
		entry.ReadAt = *readAt
	}

	if err := r.db.DB.WithContext(ctx).Create(entry).Error; err != nil {
		record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Error)
		return nil, err
	}
	record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Success)

	return entry, nil
}

func (r *UserProgressRepository) UnmarkChapter(ctx context.Context, userID string, workID string, chapterNumber int) error {
	startTime := time.Now()

	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND work_id = ? AND chapter_number = ?", userID, workID, chapterNumber).
		Delete(&UserWorkChapter{}).Error
	if err != nil {
		record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Error)
		return err
	}
	record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Success)

	return nil
}

func (r *UserProgressRepository) ReadChapters(ctx context.Context, userID string, workID string) ([]*UserWorkChapter, error) {
	startTime := time.Now()

	var chapters []*UserWorkChapter
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Order("chapter_number asc").
		Find(&chapters).Error
	if err != nil {
		record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}
	record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return chapters, nil
}

func (r *UserProgressRepository) CountReadChapters(ctx context.Context, userID string, workID string) (int64, error) {
	startTime := time.Now()

	var total int64
	err := r.db.DB.WithContext(ctx).Model(&UserWorkChapter{}).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Count(&total).Error
	if err != nil {
		record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return 0, err
	}
	record(startTime, chapterTable, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return total, nil
}
