package user_anime

import (
	"context"
	"github.com/google/uuid"
	"github.com/weeb-vip/list-service/internal/db"
)

type UserAnimeRepositoryImpl interface {
	Upsert(ctx context.Context, userAnime *UserAnime) (*UserAnime, error)
	Delete(ctx context.Context, userAnime *UserAnime) error
	FindByUserId(ctx context.Context, userId string) ([]*UserAnime, error)
	FindByAnimeId(ctx context.Context, animeId string) ([]*UserAnime, error)
	FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*UserAnime, error)
	FindByListId(ctx context.Context, listId string) ([]*UserAnime, error)
}

type UserAnimeRepository struct {
	db *db.DB
}

func NewUserAnimeRepository(db *db.DB) UserAnimeRepositoryImpl {
	return &UserAnimeRepository{db: db}
}

func (a *UserAnimeRepository) Upsert(ctx context.Context, userAnime *UserAnime) (*UserAnime, error) {

	// check if animeid and userid already exist
	var existing UserAnime
	err := a.db.DB.Where("user_id = ? AND anime_id = ?", userAnime.UserID, userAnime.AnimeID).First(&existing).Error
	if err != nil {
		if err.Error() != "record not found" {
			return nil, err
		}
	}
	// if not found, create new with uuid
	if existing.ID == "" {
		userAnime.ID = uuid.New().String()
		err := a.db.DB.Create(userAnime).Error
		if err != nil {
			return nil, err
		}
		return userAnime, nil

	}

	err = a.db.DB.Where("user_id = ? AND anime_id = ?", userAnime.UserID, userAnime.AnimeID).FirstOrCreate(userAnime).Error
	if err != nil {
		return nil, err
	}
	return userAnime, nil
}

func (a *UserAnimeRepository) Delete(ctx context.Context, userAnime *UserAnime) error {
	err := a.db.DB.Delete(userAnime).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *UserAnimeRepository) FindByUserId(ctx context.Context, userId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("user_id = ?", userId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByAnimeId(ctx context.Context, animeId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("anime_id = ?", animeId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*UserAnime, error) {
	var userAnime UserAnime
	err := a.db.DB.Where("user_id = ? AND anime_id = ?", userId, animeId).First(&userAnime).Error
	if err != nil {
		return nil, err
	}
	return &userAnime, nil
}

func (a *UserAnimeRepository) FindByListId(ctx context.Context, listId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("list_id = ?", listId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}
