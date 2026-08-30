package resolvers

import (
	"context"
	"errors"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	progress_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_progress"
	"github.com/weeb-vip/list-service/internal/services/user_progress"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// parseTimestamp reads an optional client-supplied time.
//
// An unparseable value is ignored rather than rejected: the field only says
// when something was watched, and losing that is not worth failing the mark the
// user actually asked for.
func parseTimestamp(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil
	}

	return &parsed
}

func convertWatchedEpisode(entity *progress_repo.UserAnimeEpisode) *model.WatchedEpisode {
	watchedAt := entity.WatchedAt.Format(time.RFC3339)

	return &model.WatchedEpisode{
		ID:            entity.ID,
		AnimeID:       entity.AnimeID,
		EpisodeNumber: entity.EpisodeNumber,
		WatchedAt:     &watchedAt,
	}
}

func convertReadChapter(entity *progress_repo.UserWorkChapter) *model.ReadChapter {
	readAt := entity.ReadAt.Format(time.RFC3339)

	return &model.ReadChapter{
		ID:            entity.ID,
		WorkID:        entity.WorkID,
		ChapterNumber: entity.ChapterNumber,
		ReadAt:        &readAt,
	}
}

func MarkEpisodeWatched(ctx context.Context, service user_progress.UserProgressServiceImpl, input model.MarkEpisodeInput) (*model.WatchedEpisode, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "MarkEpisodeWatched")
	span.SetAttributes(
		attribute.String("resolver.name", "MarkEpisodeWatched"),
		attribute.String("anime.id", input.AnimeID),
		attribute.Int("episode.number", input.EpisodeNumber),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkEpisodeWatched", metrics.Error)

		return nil, err
	}

	entry, err := service.MarkEpisode(ctx, *req.UserID, input.AnimeID, input.EpisodeNumber, parseTimestamp(input.WatchedAt))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkEpisodeWatched", metrics.Error)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkEpisodeWatched", metrics.Success)

	return convertWatchedEpisode(entry), nil
}

func UnmarkEpisodeWatched(ctx context.Context, service user_progress.UserProgressServiceImpl, input model.MarkEpisodeInput) error {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UnmarkEpisodeWatched")
	span.SetAttributes(
		attribute.String("resolver.name", "UnmarkEpisodeWatched"),
		attribute.String("anime.id", input.AnimeID),
		attribute.Int("episode.number", input.EpisodeNumber),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkEpisodeWatched", metrics.Error)

		return err
	}

	if err := service.UnmarkEpisode(ctx, *req.UserID, input.AnimeID, input.EpisodeNumber); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkEpisodeWatched", metrics.Error)

		return err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkEpisodeWatched", metrics.Success)

	return nil
}

func WatchedEpisodes(ctx context.Context, service user_progress.UserProgressServiceImpl, animeID string) ([]*model.WatchedEpisode, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "WatchedEpisodes")
	span.SetAttributes(
		attribute.String("resolver.name", "WatchedEpisodes"),
		attribute.String("anime.id", animeID),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "WatchedEpisodes", metrics.Error)

		return nil, err
	}

	found, err := service.WatchedEpisodes(ctx, *req.UserID, animeID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "WatchedEpisodes", metrics.Error)

		return nil, err
	}

	// Never nil: the field is non-null, and nothing watched is an empty list.
	episodes := make([]*model.WatchedEpisode, 0, len(found))
	for _, entity := range found {
		episodes = append(episodes, convertWatchedEpisode(entity))
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "WatchedEpisodes", metrics.Success)

	return episodes, nil
}

func MarkChapterRead(ctx context.Context, service user_progress.UserProgressServiceImpl, input model.MarkChapterInput) (*model.ReadChapter, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "MarkChapterRead")
	span.SetAttributes(
		attribute.String("resolver.name", "MarkChapterRead"),
		attribute.String("work.id", input.WorkID),
		attribute.Int("chapter.number", input.ChapterNumber),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkChapterRead", metrics.Error)

		return nil, err
	}

	entry, err := service.MarkChapter(ctx, *req.UserID, input.WorkID, input.ChapterNumber, parseTimestamp(input.ReadAt))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkChapterRead", metrics.Error)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "MarkChapterRead", metrics.Success)

	return convertReadChapter(entry), nil
}

func UnmarkChapterRead(ctx context.Context, service user_progress.UserProgressServiceImpl, input model.MarkChapterInput) error {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UnmarkChapterRead")
	span.SetAttributes(
		attribute.String("resolver.name", "UnmarkChapterRead"),
		attribute.String("work.id", input.WorkID),
		attribute.Int("chapter.number", input.ChapterNumber),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkChapterRead", metrics.Error)

		return err
	}

	if err := service.UnmarkChapter(ctx, *req.UserID, input.WorkID, input.ChapterNumber); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkChapterRead", metrics.Error)

		return err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UnmarkChapterRead", metrics.Success)

	return nil
}

func ReadChapters(ctx context.Context, service user_progress.UserProgressServiceImpl, workID string) ([]*model.ReadChapter, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "ReadChapters")
	span.SetAttributes(
		attribute.String("resolver.name", "ReadChapters"),
		attribute.String("work.id", workID),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "ReadChapters", metrics.Error)

		return nil, err
	}

	found, err := service.ReadChapters(ctx, *req.UserID, workID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "ReadChapters", metrics.Error)

		return nil, err
	}

	chapters := make([]*model.ReadChapter, 0, len(found))
	for _, entity := range found {
		chapters = append(chapters, convertReadChapter(entity))
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "ReadChapters", metrics.Success)

	return chapters, nil
}
