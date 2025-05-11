package user_list

import (
	"context"
	"github.com/google/uuid"
	"github.com/weeb-vip/list-service/internal/db"
)

type UserListRepositoryImpl interface {
	FindAll(ctx context.Context) ([]*UserList, error)
	FindById(ctx context.Context, id string) (*UserList, error)
	FindByUserId(ctx context.Context, userId string) ([]*UserList, error)
	Upsert(ctx context.Context, userList *UserList) (*UserList, error)
	Delete(ctx context.Context, userList *UserList) error
	FindByName(ctx context.Context, name string) ([]*UserList, error)
	FindByNameAndUserId(ctx context.Context, name string, userId string) ([]*UserList, error)
}

type UserListRepository struct {
	db *db.DB
}

func NewUserListRepository(db *db.DB) UserListRepositoryImpl {
	return &UserListRepository{db: db}
}

func (a *UserListRepository) FindAll(ctx context.Context) ([]*UserList, error) {
	var userLists []*UserList
	err := a.db.DB.Find(&userLists).Error
	if err != nil {
		return nil, err
	}
	return userLists, nil
}

func (a *UserListRepository) FindById(ctx context.Context, id string) (*UserList, error) {
	var userList UserList
	err := a.db.DB.Where("id = ?", id).First(&userList).Error
	if err != nil {
		return nil, err
	}
	return &userList, nil
}

func (a *UserListRepository) FindByUserId(ctx context.Context, userId string) ([]*UserList, error) {
	var userLists []*UserList
	err := a.db.DB.Where("user_id = ?", userId).Find(&userLists).Error
	if err != nil {
		return nil, err
	}
	return userLists, nil
}

func (a *UserListRepository) Upsert(ctx context.Context, userList *UserList) (*UserList, error) {
	if userList.ID == "" {
		userList.ID = uuid.New().String()
		err := a.db.DB.Save(userList).Error
		if err != nil {
			return nil, err
		}
		return userList, nil
	}

	// update existing user list
	err := a.db.DB.Model(userList).Where("id = ?", userList.ID).Updates(userList).Error
	if err != nil {
		return nil, err
	}

	return userList, nil
}

func (a *UserListRepository) Delete(ctx context.Context, userList *UserList) error {
	err := a.db.DB.Delete(userList).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *UserListRepository) FindByName(ctx context.Context, name string) ([]*UserList, error) {
	var userLists []*UserList
	err := a.db.DB.Where("name = ?", name).Find(&userLists).Error
	if err != nil {
		return nil, err
	}
	return userLists, nil
}

func (a *UserListRepository) FindByNameAndUserId(ctx context.Context, name string, userId string) ([]*UserList, error) {
	var userLists []*UserList
	err := a.db.DB.Where("name = ? AND user_id = ?", name, userId).Find(&userLists).Error
	if err != nil {
		return nil, err
	}
	return userLists, nil
}
