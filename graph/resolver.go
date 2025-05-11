package graph

import (
	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_list"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Config           config.Config
	UserListService  user_list.UserListServiceImpl
	UserAnimeService user_anime.UserAnimeServiceImpl
}
