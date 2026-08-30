package resolvers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	"github.com/weeb-vip/list-service/internal/dataloader"
	user_work_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_work"
	"github.com/weeb-vip/list-service/internal/services/user_work"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func ConvertUserWorkToGraphql(entity *user_work_repo.UserWork) (*model.UserWork, error) {
	var status *model.WorkStatus
	if entity.Status != nil {
		value := model.WorkStatus(*entity.Status)
		status = &value
	}

	// Absent stays absent. Splitting an empty string yields [""], which reaches
	// a client as a single blank tag.
	var tags []string
	if entity.Tags != nil && *entity.Tags != "" {
		tags = strings.Split(*entity.Tags, ",")
	}

	return &model.UserWork{
		ID:       entity.ID,
		UserID:   *entity.UserID,
		WorkID:   *entity.WorkID,
		Status:   status,
		Score:    entity.Score,
		Chapters: entity.Chapters,
		Volumes:  entity.Volumes,
		Tags:     tags,
		ListID:   entity.ListID,
	}, nil
}

// UpsertUserWork is both AddWork and UpdateWork.
//
// One resolver for two mutations, as the anime side does: the repository keys
// on (user_id, work_id), so "add" and "update" are the same write and a client
// that adds twice does not create a second row.
func UpsertUserWork(ctx context.Context, service user_work.UserWorkServiceImpl, input model.UserWorkInput) (*model.UserWork, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UpsertUserWork")
	span.SetAttributes(
		attribute.String("resolver.name", "UpsertUserWork"),
		attribute.String("work.id", input.WorkID),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UpsertUserWork", metrics.Error)

		return nil, err
	}
	span.SetAttributes(attribute.String("user.id", *userID))

	var status *user_work.UserWorkStatus
	if input.Status != nil {
		value := user_work.UserWorkStatus(*input.Status)
		status = &value
	}

	created, err := service.Upsert(ctx, &user_work.UserWork{
		ID:       input.ID,
		UserID:   *userID,
		WorkID:   input.WorkID,
		Status:   status,
		Score:    input.Score,
		Chapters: input.Chapters,
		Volumes:  input.Volumes,
		Tags:     input.Tags,
		ListID:   input.ListID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UpsertUserWork", metrics.Error)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.String("user_work.id", created.ID))
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UpsertUserWork", metrics.Success)

	return ConvertUserWorkToGraphql(created)
}

// DeleteUserWork takes the work id, which is what the page showing the manga
// holds. Removing something that is not on the shelf succeeds: the caller asked
// for it to be gone, and it is.
func DeleteUserWork(ctx context.Context, service user_work.UserWorkServiceImpl, workID string) error {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "DeleteUserWork")
	span.SetAttributes(
		attribute.String("resolver.name", "DeleteUserWork"),
		attribute.String("work.id", workID),
	)
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "DeleteUserWork", metrics.Error)

		return err
	}

	if err := service.Delete(ctx, *userID, workID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "DeleteUserWork", metrics.Error)

		return err
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "DeleteUserWork", metrics.Success)

	return nil
}

// UserWorks is a page of the viewer's shelf.
func UserWorks(ctx context.Context, service user_work.UserWorkServiceImpl, input model.UserWorksInput) (*model.UserWorkPaginated, error) {
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UserWorks")
	span.SetAttributes(attribute.String("resolver.name", "UserWorks"))
	defer span.End()

	startTime := time.Now()

	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		err := errors.New("User ID is missing, unauthenticated")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorks", metrics.Error)

		return nil, err
	}

	var status *string
	if input.Status != nil {
		value := string(*input.Status)
		status = &value
	}

	found, total, err := service.FindByUserId(ctx, *userID, status, input.Page, input.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorks", metrics.Error)

		return nil, err
	}

	// Never nil: an empty shelf is an empty list, and a null works field would
	// make every client null-check a collection that is always present.
	works := make([]*model.UserWork, 0, len(found))
	for _, entity := range found {
		converted, err := ConvertUserWorkToGraphql(entity)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorks", metrics.Error)

			return nil, err
		}
		works = append(works, converted)
	}

	span.SetStatus(codes.Ok, "")
	metrics.GetAppMetrics().ResolverMetric(float64(time.Since(startTime).Milliseconds()), "UserWorks", metrics.Success)

	// Int64 is a string over the wire in this schema, as UserAnimePaginated.Total
	// already is.
	return &model.UserWorkPaginated{
		Page:  input.Page,
		Limit: input.Limit,
		Total: strconv.FormatInt(total, 10),
		Works: works,
	}, nil
}

// GetUserWorkByWorkIDWithLoader resolves Work.userWork through the request's
// dataloader, so a page of work cards costs one query rather than one each.
//
// A signed-out viewer gets nil rather than an error: the field is on a public
// page, and "not on your shelf" is the honest answer when there is no shelf.
func GetUserWorkByWorkIDWithLoader(ctx context.Context, workID string) (*model.UserWork, error) {
	loader, ok := dataloader.GetUserWorkLoader(ctx)
	if !ok {
		// The middleware installs it on every request, so its absence is a
		// wiring fault rather than something to paper over.
		return nil, errors.New("DataLoader not available in context")
	}

	req := requestinfo.FromContext(ctx)
	if req.UserID == nil {
		return nil, nil
	}

	return loader.Load(ctx, dataloader.UserWorkKey{
		UserID: *req.UserID,
		WorkID: workID,
	})
}
