package user_anime

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/weeb-vip/list-service/internal/db"
	"gorm.io/gorm"
)

type UserAnimeRepositoryImpl interface {
	Upsert(ctx context.Context, userAnime *UserAnime) (*UserAnime, error)
	Delete(ctx context.Context, userAnime *UserAnime) error
	FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*UserAnime, int64, error)
	FindByAnimeId(ctx context.Context, animeId string) ([]*UserAnime, error)
	FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*UserAnime, error)
	FindByUserIdAndAnimeIds(ctx context.Context, userId string, animeIds []string) ([]*UserAnime, error)
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
	var existing *UserAnime
	err := a.db.DB.Where("user_id = ? AND anime_id = ?", userAnime.UserID, userAnime.AnimeID).First(&existing).Error
	if err != nil {
		if err.Error() != "record not found" {
			return nil, err
		}
	}
	// if err is gorm.ErrRecordNotFound, create new userAnime
	if errors.Is(err, gorm.ErrRecordNotFound) {
		userAnime.ID = uuid.New().String()
		err := a.db.DB.Create(userAnime).Error
		if err != nil {
			return nil, err
		}
		return userAnime, nil
	}

	// if found, update
	userAnime.ID = existing.ID
	userAnime.CreatedAt = existing.CreatedAt
	userAnime.UpdatedAt = existing.UpdatedAt
	userAnime.ListID = existing.ListID
	userAnime.CreatedAt = existing.CreatedAt
	userAnime.UpdatedAt = existing.UpdatedAt
	err = a.db.DB.Save(userAnime).Error
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

func (a *UserAnimeRepository) FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*UserAnime, int64, error) {
	var userAnimes []*UserAnime
	var total int64
	var err error
	if status != nil {
		// sort by created_at desc
		err = a.db.DB.Where("user_id = ? AND status = ?", userId, *status).Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&userAnimes).Error
	} else {
		err = a.db.DB.Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&userAnimes).Error
	}

	if err != nil {
		return nil, 0, err
	}

	// count based on status
	if status != nil {
		err = a.db.DB.Model(&UserAnime{}).Where("user_id = ? AND status = ?", userId, *status).Count(&total).Error
	} else {
		err = a.db.DB.Model(&UserAnime{}).Where("user_id = ?", userId).Count(&total).Error
	}

	if err != nil {
		return nil, 0, err
	}

	// check if total is 0
	if total == 0 {
		return nil, 0, nil
	}

	return userAnimes, total, nil
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

func (a *UserAnimeRepository) FindByUserIdAndAnimeIds(ctx context.Context, userId string, animeIds []string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("user_id = ? AND anime_id IN ?", userId, animeIds).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByListId(ctx context.Context, listId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("list_id = ?", listId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}
