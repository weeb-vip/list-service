package resolvers

import (
	"context"
	"errors"
	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	user_anime2 "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
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
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return nil, errors.New("User ID is missing, unauthenticated")
	}
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
		return nil, err
	}

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

func GetUserAnimeByID(ctx context.Context, userAnimeService user_anime.UserAnimeServiceImpl, input model.UserAnimesInput) (*model.UserAnimePaginated, error) {
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return nil, errors.New("User ID is missing, unauthenticated")
	}

	var status *string
	if input.Status != nil {
		statuss := string(*input.Status)
		status = &statuss
	} else {
		status = nil
	}
	userAnimeEntity, total, err := userAnimeService.FindByUserId(ctx, *userID, status, input.Page, input.Limit)
	if err != nil {
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

	return userAnimePaginated, nil
}
