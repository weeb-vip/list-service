package resolvers

import (
	"context"
	"strconv"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_work"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// The public counterparts of UserAnimes/UserWorks and their status counts.
//
// They differ from the authenticated versions in one thing: the userID comes in
// as an argument, not from the caller's own request context. There is no auth
// here because a public page has no viewer to authenticate; whether these are
// shown at all is the viewed user's own lists_public choice, enforced by the
// page that composes them. The reads themselves are the same service calls.

func PublicUserAnimes(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, userID string, input model.UserAnimesInput) (*model.UserAnimePaginated, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "PublicUserAnimes")
	span.SetAttributes(
		attribute.String("resolver.name", "PublicUserAnimes"),
		attribute.String("user.id", userID),
	)
	defer span.End()
	startTime := time.Now()

	var status *string
	if input.Status != nil {
		s := string(*input.Status)
		status = &s
	}

	found, total, err := userAnimeService.FindByUserId(ctx, userID, status, input.Page, input.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserAnimes", metrics.Error)
		return nil, err
	}

	animes := make([]*model.UserAnime, 0, len(found))
	for _, entity := range found {
		converted, err := ConvertUserAnimeToGraphql(entity)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		animes = append(animes, converted)
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserAnimes", metrics.Success)
	return &model.UserAnimePaginated{
		Page:   input.Page,
		Limit:  input.Limit,
		Total:  strconv.FormatInt(total, 10),
		Animes: animes,
	}, nil
}

func PublicUserWorks(ctx context.Context, userWorkService user_work.UserWorkServiceImpl, userID string, input model.UserWorksInput) (*model.UserWorkPaginated, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "PublicUserWorks")
	span.SetAttributes(
		attribute.String("resolver.name", "PublicUserWorks"),
		attribute.String("user.id", userID),
	)
	defer span.End()
	startTime := time.Now()

	var status *string
	if input.Status != nil {
		s := string(*input.Status)
		status = &s
	}

	found, total, err := userWorkService.FindByUserId(ctx, userID, status, input.Page, input.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserWorks", metrics.Error)
		return nil, err
	}

	works := make([]*model.UserWork, 0, len(found))
	for _, entity := range found {
		converted, err := ConvertUserWorkToGraphql(entity)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		works = append(works, converted)
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserWorks", metrics.Success)
	return &model.UserWorkPaginated{
		Page:  input.Page,
		Limit: input.Limit,
		Total: strconv.FormatInt(total, 10),
		Works: works,
	}, nil
}

func PublicUserAnimeStatusCounts(ctx context.Context, service user_anime.UserAnimeServiceImpl, userID string) (*model.UserAnimeStatusCounts, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "PublicUserAnimeStatusCounts")
	span.SetAttributes(attribute.String("resolver.name", "PublicUserAnimeStatusCounts"), attribute.String("user.id", userID))
	defer span.End()
	startTime := time.Now()

	counts, err := service.CountByStatus(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserAnimeStatusCounts", metrics.Error)
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserAnimeStatusCounts", metrics.Success)
	return &model.UserAnimeStatusCounts{
		Watching:    count(counts, string(model.StatusWatching)),
		PlanToWatch: count(counts, string(model.StatusPlantowatch)),
		Completed:   count(counts, string(model.StatusCompleted)),
		OnHold:      count(counts, string(model.StatusOnhold)),
		Dropped:     count(counts, string(model.StatusDropped)),
	}, nil
}

func PublicUserWorkStatusCounts(ctx context.Context, service user_work.UserWorkServiceImpl, userID string) (*model.UserWorkStatusCounts, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "PublicUserWorkStatusCounts")
	span.SetAttributes(attribute.String("resolver.name", "PublicUserWorkStatusCounts"), attribute.String("user.id", userID))
	defer span.End()
	startTime := time.Now()

	counts, err := service.CountByStatus(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserWorkStatusCounts", metrics.Error)
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "PublicUserWorkStatusCounts", metrics.Success)
	return &model.UserWorkStatusCounts{
		Reading:    count(counts, string(model.WorkStatusReading)),
		PlanToRead: count(counts, string(model.WorkStatusPlantoread)),
		Completed:  count(counts, string(model.WorkStatusCompleted)),
		OnHold:     count(counts, string(model.WorkStatusOnhold)),
		Dropped:    count(counts, string(model.WorkStatusDropped)),
	}, nil
}
