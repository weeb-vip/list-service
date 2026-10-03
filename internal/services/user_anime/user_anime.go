package user_anime

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/events"
)

// TxRunner opens a database transaction. NewTxRunner wraps the real
// connection; tests pass one that hands the function a nil transaction.
type TxRunner interface {
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// TxRunnerFunc adapts a function to TxRunner.
type TxRunnerFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error

// Transaction implements TxRunner.
func (f TxRunnerFunc) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return f(ctx, fn)
}

// NewTxRunner wraps a gorm connection.
func NewTxRunner(database *gorm.DB) TxRunner {
	return TxRunnerFunc(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
		return database.WithContext(ctx).Transaction(fn)
	})
}

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
	FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*user_anime.UserAnime, error)
	FindByUserIdAndAnimeIds(ctx context.Context, userId string, animeIds []string) ([]*user_anime.UserAnime, error)
	CountByStatus(ctx context.Context, userID string) (map[string]int64, error)
}

type UserAnimeService struct {
	Repository user_anime.UserAnimeRepositoryImpl
	Tx         TxRunner
	Events     events.Writer
}

// NewUserAnimeService wires the service. Every list change the feed cares
// about is written to the outbox by Events inside the same transaction as
// the row, through Tx.
func NewUserAnimeService(userAnimeRepository user_anime.UserAnimeRepositoryImpl, tx TxRunner, eventWriter events.Writer) UserAnimeServiceImpl {
	return &UserAnimeService{
		Repository: userAnimeRepository,
		Tx:         tx,
		Events:     eventWriter,
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

	var saved *user_anime.UserAnime
	err := a.Tx.Transaction(ctx, func(tx *gorm.DB) error {
		result, err := a.Repository.UpsertTx(ctx, tx, userAnimeEntity)
		if err != nil {
			return err
		}
		saved = result.Entity

		var previousStatus *string
		var previousScore *float64
		if result.Previous != nil {
			previousStatus = result.Previous.Status
			previousScore = result.Previous.Score
		}
		eventType := events.Decide(result.Created, previousStatus, saved.Status, previousScore, saved.Score,
			events.AnimeAdded, events.AnimeStatusChanged, events.AnimeScored)
		if eventType == "" {
			return nil
		}

		activity := &events.Activity{
			Type:    eventType,
			UserID:  userAnime.UserID,
			AnimeID: &userAnime.AnimeID,
			Status:  saved.Status,
			Score:   saved.Score,
		}
		if eventType == events.AnimeStatusChanged {
			activity.PreviousStatus = previousStatus
		}

		return a.Events.Write(ctx, tx, activity)
	})
	if err != nil {
		return nil, err
	}

	return saved, nil
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

	return a.Tx.Transaction(ctx, func(tx *gorm.DB) error {
		if err := a.Repository.DeleteTx(ctx, tx, userAnime); err != nil {
			return err
		}

		return a.Events.Write(ctx, tx, &events.Activity{
			Type:           events.AnimeRemoved,
			UserID:         userid,
			AnimeID:        userAnime.AnimeID,
			PreviousStatus: userAnime.Status,
		})
	})
}

func (a *UserAnimeService) FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*user_anime.UserAnime, int64, error) {
	userAnimes, total, err := a.Repository.FindByUserId(ctx, userId, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return userAnimes, total, nil
}

func (a *UserAnimeService) FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*user_anime.UserAnime, error) {
	userAnime, err := a.Repository.FindByUserIdAndAnimeId(ctx, userId, animeId)
	if err != nil {
		// if gorm error not found just return nil
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if userAnime == nil {
		return nil, nil
	}

	return userAnime, nil
}

func (a *UserAnimeService) FindByUserIdAndAnimeIds(ctx context.Context, userId string, animeIds []string) ([]*user_anime.UserAnime, error) {
	userAnimes, err := a.Repository.FindByUserIdAndAnimeIds(ctx, userId, animeIds)
	if err != nil {
		return nil, err
	}

	return userAnimes, nil
}

func (a *UserAnimeService) CountByStatus(ctx context.Context, userID string) (map[string]int64, error) {
	return a.Repository.CountByStatus(ctx, userID)
}
