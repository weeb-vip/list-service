package dummy

import (
	"context"
	"github.com/weeb-vip/golang-template/internal/db/repositories/dummy"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DummyServiceImpl interface {
	ByID(ctx context.Context, id string) (*dummy.Dummy, error)
}

type DummyService struct {
	Repository dummy.DummyRepositoryImpl
}

func NewDummyService(dummyRepository dummy.DummyRepositoryImpl) DummyServiceImpl {
	return &DummyService{
		Repository: dummyRepository,
	}
}

func (a *DummyService) ByID(ctx context.Context, id string) (*dummy.Dummy, error) {
	span, spanCtx := tracer.StartSpanFromContext(ctx, "AnimeByID")
	span.SetTag("service", "dummy")
	span.SetTag("type", "service")
	defer span.Finish()

	return a.Repository.FindById(spanCtx, id)
}
