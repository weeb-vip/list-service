package graph

import (
	"context"
	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_list"
	"github.com/weeb-vip/list-service/internal/services/user_work"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Config           config.Config
	UserListService  user_list.UserListServiceImpl
	UserAnimeService user_anime.UserAnimeServiceImpl
	UserWorkService  user_work.UserWorkServiceImpl
	Context          context.Context
}
