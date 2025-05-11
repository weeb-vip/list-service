package graph

import (
	"github.com/weeb-vip/golang-template/config"
	"github.com/weeb-vip/golang-template/internal/services/dummy"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Config       config.Config
	DummyService dummy.DummyServiceImpl
}
