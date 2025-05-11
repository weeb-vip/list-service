package resolvers

import (
	"context"
	"errors"
	"github.com/weeb-vip/list-service/graph/model"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	user_list2 "github.com/weeb-vip/list-service/internal/db/repositories/user_list"
	"github.com/weeb-vip/list-service/internal/services/user_list"
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
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return nil, errors.New("User ID is missing, unauthenticated")
	}
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
	createdUserList, err := userListService.UpsertUserList(ctx, userListEntity)
	if err != nil {
		return nil, err
	}

	// Convert user_list.UserList back to model.UserList

	return ConvertUserListToGraphql(createdUserList)
}

func GetUserListsByID(ctx context.Context, userListService user_list.UserListServiceImpl) ([]*model.UserList, error) {
	// get userid from requestInfo
	req := requestinfo.FromContext(ctx)
	userID := req.UserID
	if userID == nil {
		return nil, errors.New("User ID is missing, unauthenticated")
	}
	userLists, err := userListService.GetUserListsByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	userListModels := make([]*model.UserList, len(userLists))
	for i, userListEntity := range userLists {
		userListModel, err := ConvertUserListToGraphql(userListEntity)
		if err != nil {
			return nil, err
		}
		userListModels[i] = userListModel
	}

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
