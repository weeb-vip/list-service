package user_anime

import (
	"context"
	"github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"strings"
)

type UserAnimeStatus string

const (
	Watching    UserAnimeStatus = "watching"
	Completed   UserAnimeStatus = "completed"
	OnHold      UserAnimeStatus = "onhold"
	Dropped     UserAnimeStatus = "dropped"
	PlanToWatch UserAnimeStatus = "plantowatch"
)

type UserAnime struct {
	ID                 *string          `json:"id"`
	UserID             string           `json:"user_id"`
	AnimeID            string           `json:"anime_id"`
	Status             *UserAnimeStatus `json:"status"`
	Score              *float64         `json:"score"`
	Episodes           *int             `json:"episodes"`
	Rewatching         *int             `json:"rewatching"`
	RewatchingEpisodes *int             `json:"rewatching_episodes"`
	Tags               []string         `json:"tags"`
	ListID             *string          `json:"list_id"`
	CreatedAt          string           `json:"created_at"`
	UpdatedAt          string           `json:"updated_at"`
	DeletedAt          string           `json:"deleted_at"`
}

type UserAnimePaginated struct {
	Page   int          `json:"page"`
	Limit  int          `json:"limit"`
	Total  int          `json:"total"`
	Animes []*UserAnime `json:"animes"`
}

type UserAnimeServiceImpl interface {
	Upsert(ctx context.Context, userAnime *UserAnime) (*user_anime.UserAnime, error)
	Delete(ctx context.Context, userid string, id string) error
	FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*user_anime.UserAnime, int64, error)
}

type UserAnimeService struct {
	Repository user_anime.UserAnimeRepositoryImpl
}

func NewUserAnimeService(userAnimeRepository user_anime.UserAnimeRepositoryImpl) UserAnimeServiceImpl {
	return &UserAnimeService{
		Repository: userAnimeRepository,
	}
}

func (a *UserAnimeService) Upsert(ctx context.Context, userAnime *UserAnime) (*user_anime.UserAnime, error) {

	tags := strings.Join(userAnime.Tags, ",")
	var id string
	if userAnime.ID != nil {
		id = *userAnime.ID
	} else {
		id = ""
	}
	var status *string
	if userAnime.Status != nil {
		statuss := string(*userAnime.Status)
		status = &statuss
	} else {
		status = nil
	}
	userAnimeEntity := &user_anime.UserAnime{
		ID:                 id,
		UserID:             &userAnime.UserID,
		AnimeID:            &userAnime.AnimeID,
		Status:             status,
		Score:              userAnime.Score,
		Episodes:           userAnime.Episodes,
		Rewatching:         userAnime.Rewatching,
		RewatchingEpisodes: userAnime.RewatchingEpisodes,
		Tags:               &tags,
		ListID:             userAnime.ListID,
	}

	return a.Repository.Upsert(ctx, userAnimeEntity)
}

func (a *UserAnimeService) Delete(ctx context.Context, userid string, id string) error {
	userAnime, err := a.Repository.FindByUserIdAndAnimeId(ctx, userid, id)
	if err != nil {
		return err
	}

	if userAnime == nil {
		return nil
	}

	if *userAnime.UserID != userid {
		return nil
	}

	err = a.Repository.Delete(ctx, userAnime)
	if err != nil {
		return err
	}

	return nil
}

func (a *UserAnimeService) FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*user_anime.UserAnime, int64, error) {
	userAnimes, total, err := a.Repository.FindByUserId(ctx, userId, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return userAnimes, total, nil
}
