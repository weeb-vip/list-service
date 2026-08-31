package resolvers

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_work"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// count reads one status out of the grouped result, defaulting to zero.
//
// A status the reader has nothing in is absent from the map, not a stored zero,
// so every field is filled here rather than trusting the query to return a row
// for it. Int64 is a string over the wire in this schema, as Total already is.
func count(counts map[string]int64, status string) string {
	return strconv.FormatInt(counts[status], 10)
}

// UserAnimeStatusCounts returns the size of each watchlist status in one query,
// so the profile's tabs can each show a number without a request per tab.
func UserAnimeStatusCounts(ctx context.Context, service user_anime.UserAnimeServiceImpl) (*model.UserAnimeStatusCounts, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UserAnimeStatusCounts")
	span.SetAttributes(attribute.String("resolver.name", "UserAnimeStatusCounts"))
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserAnimeStatusCounts", metrics.Error)

		return nil, err
	}

	counts, err := service.CountByStatus(ctx, *userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserAnimeStatusCounts", metrics.Error)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserAnimeStatusCounts", metrics.Success)

	return &model.UserAnimeStatusCounts{
		Watching:    count(counts, string(model.StatusWatching)),
		PlanToWatch: count(counts, string(model.StatusPlantowatch)),
		Completed:   count(counts, string(model.StatusCompleted)),
		OnHold:      count(counts, string(model.StatusOnhold)),
		Dropped:     count(counts, string(model.StatusDropped)),
	}, nil
}

// UserWorkStatusCounts is the reading counterpart of UserAnimeStatusCounts.
func UserWorkStatusCounts(ctx context.Context, service user_work.UserWorkServiceImpl) (*model.UserWorkStatusCounts, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UserWorkStatusCounts")
	span.SetAttributes(attribute.String("resolver.name", "UserWorkStatusCounts"))
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorkStatusCounts", metrics.Error)

		return nil, err
	}

	counts, err := service.CountByStatus(ctx, *userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorkStatusCounts", metrics.Error)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorkStatusCounts", metrics.Success)

	return &model.UserWorkStatusCounts{
		Reading:    count(counts, string(model.WorkStatusReading)),
		PlanToRead: count(counts, string(model.WorkStatusPlantoread)),
		Completed:  count(counts, string(model.WorkStatusCompleted)),
		OnHold:     count(counts, string(model.WorkStatusOnhold)),
		Dropped:    count(counts, string(model.WorkStatusDropped)),
	}, nil
}
