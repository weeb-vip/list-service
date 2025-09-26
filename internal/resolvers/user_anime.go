package resolvers

import (
	"context"
	"errors"
	"time"

	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	user_anime2 "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/dataloader"
	"github.com/weeb-vip/list-service/internal/logger"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/metrics"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"strconv"
	"strings"
)

func ConvertUserAnimeToGraphql(userAnimeEntity *user_anime2.UserAnime) (*model.UserAnime, error) {
	var status *model.Status
	if userAnimeEntity.Status != nil {
		statuss := model.Status(*userAnimeEntity.Status)
		status = &statuss
	} else {
		status = nil
	}

	var tags []string
	if userAnimeEntity.Tags != nil {
		// split tags by comma
		tags = strings.Split(*userAnimeEntity.Tags, ",")
	} else {
		tags = nil
	}

	return &model.UserAnime{
		ID:                 userAnimeEntity.ID,
		UserID:             *userAnimeEntity.UserID,
		AnimeID:            *userAnimeEntity.AnimeID,
		Status:             status,
		Episodes:           userAnimeEntity.Episodes,
		Score:              userAnimeEntity.Score,
		Tags:               tags,
		ListID:             userAnimeEntity.ListID,
		Rewatching:         userAnimeEntity.Rewatching,
		RewatchingEpisodes: userAnimeEntity.RewatchingEpisodes,
	}, nil
}

func UpsertUserAnime(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, userAnime model.UserAnimeInput) (*model.UserAnime, error) {
	// Start tracing span
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "UpsertUserAnime")
	span.SetAttributes(
		attribute.String("resolver.name", "UpsertUserAnime"),
		attribute.String("anime.id", userAnime.AnimeID),
	)
	defer span.End()

	startTime := time.Now()

	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		span.RecordError(errors.New("User ID is missing, unauthenticated"))
		span.SetStatus(codes.Error, "User ID is missing, unauthenticated")

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"UpsertUserAnime",
			metrics.Error,
		)

		return nil, errors.New("User ID is missing, unauthenticated")
	}

	span.SetAttributes(attribute.String("user.id", *userID))
	var status *user_anime.UserAnimeStatus
	if userAnime.Status != nil {
		statuss := user_anime.UserAnimeStatus(*userAnime.Status)
		status = &statuss
	} else {
		status = nil
	}
	// Convert model.UserListInput to user_list.UserList
	userAnimeEntity := &user_anime.UserAnime{
		ID:                 userAnime.ID,
		UserID:             *userID,
		AnimeID:            userAnime.AnimeID,
		Status:             status,
		Score:              userAnime.Score,
		Episodes:           userAnime.Episodes,
		Rewatching:         userAnime.Rewatching,
		RewatchingEpisodes: userAnime.RewatchingEpisodes,
		Tags:               userAnime.Tags,
		ListID:             userAnime.ListID,
	}

	createdUserAnime, err := userAnimeService.Upsert(ctx, userAnimeEntity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"UpsertUserAnime",
			metrics.Error,
		)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.String("user_anime.id", createdUserAnime.ID))

	metrics.GetAppMetrics().ResolverMetric(
		float64(time.Since(startTime).Milliseconds()),
		"UpsertUserAnime",
		metrics.Success,
	)

	return ConvertUserAnimeToGraphql(createdUserAnime)
}

func DeleteUserAnime(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, id string) error {
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return errors.New("User ID is missing, unauthenticated")
	}

	err := userAnimeService.Delete(ctx, *userID, id)
	if err != nil {
		return err
	}

	return nil
}

func GetUserAnimesByID(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, input model.UserAnimesInput) (*model.UserAnimePaginated, error) {
	// Start tracing span
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "GetUserAnimesByID")
	span.SetAttributes(
		attribute.String("resolver.name", "GetUserAnimesByID"),
		attribute.Int("page", input.Page),
		attribute.Int("limit", input.Limit),
	)
	defer span.End()

	startTime := time.Now()

	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		span.RecordError(errors.New("User ID is missing, unauthenticated"))
		span.SetStatus(codes.Error, "User ID is missing, unauthenticated")

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"GetUserAnimesByID",
			metrics.Error,
		)

		return nil, errors.New("User ID is missing, unauthenticated")
	}

	span.SetAttributes(attribute.String("user.id", *userID))

	var status *string
	if input.Status != nil {
		statuss := string(*input.Status)
		status = &statuss
	} else {
		status = nil
	}
	userAnimeEntity, total, err := userAnimeService.FindByUserId(ctx, *userID, status, input.Page, input.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"GetUserAnimesByID",
			metrics.Error,
		)

		return nil, err
	}

	if userAnimeEntity == nil {
		return nil, nil
	}

	var userAnimeModels []*model.UserAnime
	for _, userAnime := range userAnimeEntity {
		userAnimeModel, err := ConvertUserAnimeToGraphql(userAnime)
		if err != nil {
			return nil, err
		}
		userAnimeModels = append(userAnimeModels, userAnimeModel)
	}

	var totalAsString string
	if total != 0 {
		totalAsString = strconv.Itoa(int(total))
	} else {
		totalAsString = "0"
	}
	userAnimePaginated := &model.UserAnimePaginated{
		Page:   input.Page,
		Limit:  input.Limit,
		Total:  totalAsString,
		Animes: userAnimeModels,
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Int("user_anime.count", len(userAnimeModels)))

	metrics.GetAppMetrics().ResolverMetric(
		float64(time.Since(startTime).Milliseconds()),
		"GetUserAnimesByID",
		metrics.Success,
	)

	return userAnimePaginated, nil
}

func GetUserAnimeByAnimeID(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, animeID string) (*model.UserAnime, error) {
	// Start tracing span
	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "GetUserAnimeByAnimeID")
	span.SetAttributes(
		attribute.String("resolver.name", "GetUserAnimeByAnimeID"),
		attribute.String("anime.id", animeID),
	)
	defer span.End()

	startTime := time.Now()

	log := logger.FromCtx(ctx)
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		log.Error().Msg("User ID is missing, unauthenticated")
		span.SetStatus(codes.Error, "User ID is missing, unauthenticated")

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"GetUserAnimeByAnimeID",
			metrics.Error,
		)

		return nil, nil
	}

	span.SetAttributes(attribute.String("user.id", *userID))

	log.Info().
		Str("userID", *userID).
		Str("animeID", animeID).
		Msg("Fetching user anime for userID")
	userAnimeEntity, err := userAnimeService.FindByUserIdAndAnimeId(ctx, *userID, animeID)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		metrics.GetAppMetrics().ResolverMetric(
			float64(time.Since(startTime).Milliseconds()),
			"GetUserAnimeByAnimeID",
			metrics.Error,
		)

		return nil, err
	}

	if userAnimeEntity == nil {
		return nil, nil
	}

	userAnimeModel, err := ConvertUserAnimeToGraphql(userAnimeEntity)
	if err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	if userAnimeEntity != nil {
		span.SetAttributes(attribute.String("user_anime.id", userAnimeEntity.ID))
	}

	metrics.GetAppMetrics().ResolverMetric(
		float64(time.Since(startTime).Milliseconds()),
		"GetUserAnimeByAnimeID",
		metrics.Success,
	)

	return userAnimeModel, nil
}

func GetUserAnimeByAnimeIDWithLoader(ctx context.Context, animeID string) (*model.UserAnime, error) {
	// Try to get DataLoader from context
	if loader, ok := dataloader.GetUserAnimeLoader(ctx); ok {
		// Get userID from context
		req := requestinfo.FromContext(ctx)
		if req.UserID == nil {
			return nil, nil
		}
		
		// Use DataLoader to batch the request
		key := dataloader.UserAnimeKey{
			UserID:  *req.UserID,
			AnimeID: animeID,
		}
		
		return loader.Load(ctx, key)
	}
	
	// Fallback to individual query if DataLoader not available
	// This should not happen in normal flow
	return nil, errors.New("DataLoader not available in context")
}
