package dummy

import (
	"context"
	metrics_lib "github.com/TempMee/go-metrics-lib"
	"github.com/weeb-vip/golang-template/internal/db"
	"github.com/weeb-vip/golang-template/metrics"
	"time"
)

type RECORD_TYPE string

type DummyRepositoryImpl interface {
	FindAll(ctx context.Context) ([]*Dummy, error)
	FindById(ctx context.Context, id string) (*Dummy, error)
}

type AnimeRepository struct {
	db *db.DB
}

func NewDummyRepository(db *db.DB) DummyRepositoryImpl {
	return &AnimeRepository{db: db}
}

func (a *AnimeRepository) FindAll(ctx context.Context) ([]*Dummy, error) {
	startTime := time.Now()

	var animes []*Dummy
	err := a.db.DB.Find(&animes).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "scraper-api",
			Table:   "dummy",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "scraper-api",
		Table:   "dummy",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
	})
	return animes, nil
}

func (a *AnimeRepository) FindById(ctx context.Context, id string) (*Dummy, error) {
	startTime := time.Now()

	var anime Dummy
	err := a.db.DB.Where("id = ?", id).First(&anime).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "scraper-api",
			Table:   "dummy",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "scraper-api",
		Table:   "dummy",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
	})
	return &anime, nil
}
