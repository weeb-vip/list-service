package user_list

import (
	"context"
	"github.com/weeb-vip/list-service/internal/db/repositories/user_list"
	"strings"
)

type UserList struct {
	ID          *string
	UserID      string
	Name        string
	IsPublic    bool
	Tags        []string
	Description *string
}

type UserListServiceImpl interface {
	GetUserListsByID(ctx context.Context, userID string) ([]*user_list.UserList, error)
	Upsert(ctx context.Context, userList *UserList) (*user_list.UserList, error)
	DeleteUserList(ctx context.Context, userid string, id string) error
}

type UserListService struct {
	Repository user_list.UserListRepositoryImpl
}

func NewUserListService(repository user_list.UserListRepositoryImpl) UserListServiceImpl {
	return &UserListService{
		Repository: repository,
	}
}

func (u *UserListService) GetUserListsByID(ctx context.Context, userID string) ([]*user_list.UserList, error) {
	userLists, err := u.Repository.FindByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userLists, nil
}

func (u *UserListService) Upsert(ctx context.Context, userList *UserList) (*user_list.UserList, error) {
	// Convert model.UserList to user_list.UserList
	// convert tags to comma separated string
	tags := strings.Join(userList.Tags, ",")
	var id string
	if userList.ID != nil {
		id = *userList.ID
	} else {
		id = ""

	}
	userListEntity := &user_list.UserList{
		ID:          id,
		UserID:      &userList.UserID,
		Name:        &userList.Name,
		IsPublic:    &userList.IsPublic,
		Tags:        &tags,
		Description: userList.Description,
	}

	// Upsert the user list
	createdUserList, err := u.Repository.Upsert(ctx, userListEntity)
	if err != nil {
		return nil, err
	}

	// Convert user_list.UserList back to model.UserList
	return createdUserList, nil
}

func (u *UserListService) DeleteUserList(ctx context.Context, userid string, id string) error {
	userList, err := u.Repository.FindById(ctx, id)
	if err != nil {
		return err
	}

	if userList == nil {
		return nil
	}

	if *userList.UserID != userid {
		return nil
	}

	err = u.Repository.Delete(ctx, userList)
	if err != nil {
		return err
	}

	return nil
}
