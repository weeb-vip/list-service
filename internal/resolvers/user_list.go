package resolvers

import (
	"context"
	"errors"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	user_list2 "github.com/weeb-vip/list-service/internal/db/repositories/user_list"
	"github.com/weeb-vip/list-service/internal/services/user_list"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	metrics_lib "github.com/weeb-vip/go-metrics-lib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"strings"
)

func ConvertUserListToGraphql(userListEntity *user_list2.UserList) (*model.UserList, error) {
	if userListEntity == nil {
		return nil, nil
	}

	tags := strings.Split(*userListEntity.Tags, ",")
	return &model.UserList{
		ID:          userListEntity.ID,
		UserID:      *userListEntity.UserID,
		Name:        *userListEntity.Name,
		IsPublic:    userListEntity.IsPublic,
		Tags:        tags,
		Description: userListEntity.Description,
	}, nil
}

func UpsertUserList(ctx context.Context, userListService user_list.UserListServiceImpl, userList model.UserListInput) (*model.UserList, error) {
	// Start tracing span
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UpsertUserList")
	span.SetAttributes(
		attribute.String("resolver.name", "UpsertUserList"),
		attribute.String("user_list.name", userList.Name),
	)
	defer span.End()

	startTime := time.Now()

	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		span.RecordError(errors.New("User ID is missing, unauthenticated"))
		span.SetStatus(codes.Error, "User ID is missing, unauthenticated")

		_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
			Resolver: "UpsertUserList",
			Service:  "list-service",
			Protocol: "graphql",
			Result:   metrics_lib.Error,
			Env:      metrics.GetCurrentEnv(),
		})

		return nil, errors.New("User ID is missing, unauthenticated")
	}

	span.SetAttributes(attribute.String("user.id", *userID))

	// Convert model.UserListInput to user_list.UserList
	var isPublic bool
	if userList.IsPublic != nil {
		isPublic = *userList.IsPublic
	} else {
		isPublic = false
	}
	userListEntity := &user_list.UserList{
		ID:          userList.ID,
		UserID:      *userID,
		Name:        userList.Name,
		IsPublic:    isPublic,
		Tags:        userList.Tags,
		Description: userList.Description,
	}
	createdUserList, err := userListService.Upsert(ctx, userListEntity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
			Resolver: "UpsertUserList",
			Service:  "list-service",
			Protocol: "graphql",
			Result:   metrics_lib.Error,
			Env:      metrics.GetCurrentEnv(),
		})

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.String("user_list.id", createdUserList.ID))

	_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
		Resolver: "UpsertUserList",
		Service:  "list-service",
		Protocol: "graphql",
		Result:   metrics_lib.Success,
		Env:      metrics.GetCurrentEnv(),
	})

	// Convert user_list.UserList back to model.UserList
	return ConvertUserListToGraphql(createdUserList)
}

func GetUserListsByID(ctx context.Context, userListService user_list.UserListServiceImpl) ([]*model.UserList, error) {
	// Start tracing span
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "GetUserListsByID")
	span.SetAttributes(
		attribute.String("resolver.name", "GetUserListsByID"),
	)
	defer span.End()

	startTime := time.Now()

	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		span.RecordError(errors.New("User ID is missing, unauthenticated"))
		span.SetStatus(codes.Error, "User ID is missing, unauthenticated")

		_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
			Resolver: "GetUserListsByID",
			Service:  "list-service",
			Protocol: "graphql",
			Result:   metrics_lib.Error,
			Env:      metrics.GetCurrentEnv(),
		})

		return nil, errors.New("User ID is missing, unauthenticated")
	}

	span.SetAttributes(attribute.String("user.id", *userID))

	userLists, err := userListService.GetUserListsByID(ctx, *userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
			Resolver: "GetUserListsByID",
			Service:  "list-service",
			Protocol: "graphql",
			Result:   metrics_lib.Error,
			Env:      metrics.GetCurrentEnv(),
		})

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Int("user_lists.count", len(userLists)))

	userListModels := make([]*model.UserList, len(userLists))
	for i, userListEntity := range userLists {
		userListModel, err := ConvertUserListToGraphql(userListEntity)
		if err != nil {
			return nil, err
		}
		userListModels[i] = userListModel
	}

	_ = metrics.NewMetricsInstance().ResolverMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.ResolverMetricLabels{
		Resolver: "GetUserListsByID",
		Service:  "list-service",
		Protocol: "graphql",
		Result:   metrics_lib.Success,
		Env:      metrics.GetCurrentEnv(),
	})

	return userListModels, nil
}

func DeleteUserList(ctx context.Context, userListService user_list.UserListServiceImpl, id string) error {
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return errors.New("User ID is missing, unauthenticated")
	}

	err := userListService.DeleteUserList(ctx, *userID, id)
	if err != nil {
		return err
	}

	return nil
}
