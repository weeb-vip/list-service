package user_anime

import "github.com/weeb-vip/list-service/internal/db"

type UserAnimeRepositoryImpl interface {
	Upsert(userAnime *UserAnime) error
	Delete(userAnime *UserAnime) error
	FindByUserId(userId string) ([]*UserAnime, error)
	FindByAnimeId(animeId string) ([]*UserAnime, error)
	FindByUserIdAndAnimeId(userId string, animeId string) (*UserAnime, error)
	FindByListId(listId string) ([]*UserAnime, error)
}

type UserAnimeRepository struct {
	db *db.DB
}

func NewUserAnimeRepository(db *db.DB) UserAnimeRepositoryImpl {
	return &UserAnimeRepository{db: db}
}

func (a *UserAnimeRepository) Upsert(userAnime *UserAnime) error {
	err := a.db.DB.Save(userAnime).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *UserAnimeRepository) Delete(userAnime *UserAnime) error {
	err := a.db.DB.Delete(userAnime).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *UserAnimeRepository) FindByUserId(userId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("user_id = ?", userId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByAnimeId(animeId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("anime_id = ?", animeId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByUserIdAndAnimeId(userId string, animeId string) (*UserAnime, error) {
	var userAnime UserAnime
	err := a.db.DB.Where("user_id = ? AND anime_id = ?", userId, animeId).First(&userAnime).Error
	if err != nil {
		return nil, err
	}
	return &userAnime, nil
}

func (a *UserAnimeRepository) FindByListId(listId string) ([]*UserAnime, error) {
	var userAnimes []*UserAnime
	err := a.db.DB.Where("list_id = ?", listId).Find(&userAnimes).Error
	if err != nil {
		return nil, err
	}
	return userAnimes, nil
}
