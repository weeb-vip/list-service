package user_anime

import (
	"context"
	"errors"
	"time"
	"github.com/google/uuid"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/metrics"
	metrics_lib "github.com/weeb-vip/go-metrics-lib"
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
	startTime := time.Now()

	// check if animeid and userid already exist
	var existing *UserAnime
	err := a.db.DB.WithContext(ctx).Where("user_id = ? AND anime_id = ?", userAnime.UserID, userAnime.AnimeID).First(&existing).Error
	if err != nil {
		if err.Error() != "record not found" && !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
				Service: "list-service",
				Table:   "user_anime",
				Method:  metrics_lib.DatabaseMetricMethodSelect,
				Result:  metrics_lib.Error,
			})
			return nil, err
		}
	}
	// if err is gorm.ErrRecordNotFound, create new userAnime
	if errors.Is(err, gorm.ErrRecordNotFound) {
		userAnime.ID = uuid.New().String()
		err := a.db.DB.WithContext(ctx).Create(userAnime).Error
		if err != nil {
			_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
				Service: "list-service",
				Table:   "user_anime",
				Method:  metrics_lib.DatabaseMetricMethodInsert,
				Result:  metrics_lib.Error,
				Env:     metrics.GetCurrentEnv(),
			})
			return nil, err
		}

		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodInsert,
			Result:  metrics_lib.Success,
			Env:     metrics.GetCurrentEnv(),
		})
		return userAnime, nil
	}

	// if found, update
	userAnime.ID = existing.ID
	userAnime.CreatedAt = existing.CreatedAt
	userAnime.UpdatedAt = existing.UpdatedAt
	userAnime.ListID = existing.ListID
	userAnime.CreatedAt = existing.CreatedAt
	userAnime.UpdatedAt = existing.UpdatedAt
	err = a.db.DB.WithContext(ctx).Save(userAnime).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodUpdate,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodUpdate,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userAnime, nil
}

func (a *UserAnimeRepository) Delete(ctx context.Context, userAnime *UserAnime) error {
	startTime := time.Now()

	err := a.db.DB.WithContext(ctx).Delete(userAnime).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodDelete,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodDelete,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return nil
}

func (a *UserAnimeRepository) FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*UserAnime, int64, error) {
	startTime := time.Now()

	var userAnimes []*UserAnime
	var total int64
	var err error
	if status != nil {
		// sort by created_at desc
		err = a.db.DB.WithContext(ctx).Where("user_id = ? AND status = ?", userId, *status).Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&userAnimes).Error
	} else {
		err = a.db.DB.WithContext(ctx).Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&userAnimes).Error
	}

	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, 0, err
	}

	// count based on status
	if status != nil {
		err = a.db.DB.WithContext(ctx).Model(&UserAnime{}).Where("user_id = ? AND status = ?", userId, *status).Count(&total).Error
	} else {
		err = a.db.DB.WithContext(ctx).Model(&UserAnime{}).Where("user_id = ?", userId).Count(&total).Error
	}

	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, 0, err
	}

	// check if total is 0
	if total == 0 {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Success,
		})
		return nil, 0, nil
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userAnimes, total, nil
}

func (a *UserAnimeRepository) FindByAnimeId(ctx context.Context, animeId string) ([]*UserAnime, error) {
	startTime := time.Now()

	var userAnimes []*UserAnime
	err := a.db.DB.WithContext(ctx).Where("anime_id = ?", animeId).Find(&userAnimes).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByUserIdAndAnimeId(ctx context.Context, userId string, animeId string) (*UserAnime, error) {
	startTime := time.Now()

	var userAnime UserAnime
	err := a.db.DB.WithContext(ctx).Where("user_id = ? AND anime_id = ?", userId, animeId).First(&userAnime).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return &userAnime, nil
}

func (a *UserAnimeRepository) FindByUserIdAndAnimeIds(ctx context.Context, userId string, animeIds []string) ([]*UserAnime, error) {
	startTime := time.Now()

	var userAnimes []*UserAnime
	err := a.db.DB.WithContext(ctx).Where("user_id = ? AND anime_id IN ?", userId, animeIds).Find(&userAnimes).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userAnimes, nil
}

func (a *UserAnimeRepository) FindByListId(ctx context.Context, listId string) ([]*UserAnime, error) {
	startTime := time.Now()

	var userAnimes []*UserAnime
	err := a.db.DB.WithContext(ctx).Where("list_id = ?", listId).Find(&userAnimes).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "list-service",
			Table:   "user_anime",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "list-service",
		Table:   "user_anime",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userAnimes, nil
}
