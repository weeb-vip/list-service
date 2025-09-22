package user_list

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/metrics"
	metrics_lib "github.com/weeb-vip/go-metrics-lib"
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
	startTime := time.Now()

	var userLists []*UserList
	err := a.db.DB.WithContext(ctx).Find(&userLists).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userLists, nil
}

func (a *UserListRepository) FindById(ctx context.Context, id string) (*UserList, error) {
	startTime := time.Now()

	var userList UserList
	err := a.db.DB.WithContext(ctx).Where("id = ?", id).First(&userList).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return &userList, nil
}

func (a *UserListRepository) FindByUserId(ctx context.Context, userId string) ([]*UserList, error) {
	startTime := time.Now()

	var userLists []*UserList
	err := a.db.DB.WithContext(ctx).Where("user_id = ?", userId).Find(&userLists).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userLists, nil
}

func (a *UserListRepository) Upsert(ctx context.Context, userList *UserList) (*UserList, error) {
	startTime := time.Now()

	if userList.ID == "" {
		userList.ID = uuid.New().String()
		err := a.db.DB.WithContext(ctx).Save(userList).Error
		if err != nil {
			_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
				Service: metrics.GetServiceName(),
				Table:   "user_lists",
				Method:  metrics_lib.DatabaseMetricMethodInsert,
				Result:  metrics_lib.Error,
				Env:     metrics.GetCurrentEnv(),
			})
			return nil, err
		}

		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodInsert,
			Result:  metrics_lib.Success,
			Env:     metrics.GetCurrentEnv(),
		})
		return userList, nil
	}

	// update existing user list
	err := a.db.DB.WithContext(ctx).Model(userList).Where("id = ?", userList.ID).Updates(userList).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodUpdate,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodUpdate,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userList, nil
}

func (a *UserListRepository) Delete(ctx context.Context, userList *UserList) error {
	startTime := time.Now()

	err := a.db.DB.WithContext(ctx).Delete(userList).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodDelete,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodDelete,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return nil
}

func (a *UserListRepository) FindByName(ctx context.Context, name string) ([]*UserList, error) {
	startTime := time.Now()

	var userLists []*UserList
	err := a.db.DB.WithContext(ctx).Where("name = ?", name).Find(&userLists).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userLists, nil
}

func (a *UserListRepository) FindByNameAndUserId(ctx context.Context, name string, userId string) ([]*UserList, error) {
	startTime := time.Now()

	var userLists []*UserList
	err := a.db.DB.WithContext(ctx).Where("name = ? AND user_id = ?", name, userId).Find(&userLists).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: metrics.GetServiceName(),
			Table:   "user_lists",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
			Env:     metrics.GetCurrentEnv(),
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   "user_lists",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
		Env:     metrics.GetCurrentEnv(),
	})
	return userLists, nil
}
