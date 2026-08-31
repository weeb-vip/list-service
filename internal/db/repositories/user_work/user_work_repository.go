package user_work

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	metrics_lib "github.com/weeb-vip/go-metrics-lib"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/metrics"
	"gorm.io/gorm"
)

const table = "user_work"

type UserWorkRepositoryImpl interface {
	Upsert(ctx context.Context, userWork *UserWork) (*UserWork, error)
	Delete(ctx context.Context, userWork *UserWork) error
	FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*UserWork, int64, error)
	FindByUserIdAndWorkId(ctx context.Context, userId string, workId string) (*UserWork, error)
	FindByUserIdAndWorkIds(ctx context.Context, userId string, workIds []string) ([]*UserWork, error)
	CountByStatus(ctx context.Context, userId string) (map[string]int64, error)
}

type UserWorkRepository struct {
	db *db.DB
}

func NewUserWorkRepository(db *db.DB) UserWorkRepositoryImpl {
	return &UserWorkRepository{db: db}
}

// record reports one database call. Written once here rather than inlined at
// every return, which is how user_anime's repository ended up with one metric
// missing its Env label and another its whole call.
func record(startTime time.Time, method metrics_lib.DatabaseMetricMethod, result metrics_lib.Result) {
	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: metrics.GetServiceName(),
		Table:   table,
		Method:  method,
		Result:  result,
		Env:     metrics.GetCurrentEnv(),
	})
}

// Upsert writes the reader's row for a work, creating it the first time.
//
// Keyed on (user_id, work_id) rather than the primary key, because the caller
// knows which work it is looking at and not whether a row already exists.
func (r *UserWorkRepository) Upsert(ctx context.Context, userWork *UserWork) (*UserWork, error) {
	startTime := time.Now()

	var existing UserWork
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND work_id = ?", userWork.UserID, userWork.WorkID).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		userWork.ID = uuid.New().String()
		if err := r.db.DB.WithContext(ctx).Create(userWork).Error; err != nil {
			record(startTime, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Error)
			return nil, err
		}
		record(startTime, metrics_lib.DatabaseMetricMethodInsert, metrics_lib.Success)

		return userWork, nil
	}

	// Carry the identity and the timestamps of the row that already exists.
	// Without this Save writes a zero created_at and the shelf, which orders by
	// it, reshuffles every time a reader updates their progress.
	userWork.ID = existing.ID
	userWork.CreatedAt = existing.CreatedAt
	userWork.UpdatedAt = existing.UpdatedAt
	if userWork.ListID == nil {
		userWork.ListID = existing.ListID
	}

	if err := r.db.DB.WithContext(ctx).Save(userWork).Error; err != nil {
		record(startTime, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Error)
		return nil, err
	}
	record(startTime, metrics_lib.DatabaseMetricMethodUpdate, metrics_lib.Success)

	return userWork, nil
}

func (r *UserWorkRepository) Delete(ctx context.Context, userWork *UserWork) error {
	startTime := time.Now()

	if err := r.db.DB.WithContext(ctx).Delete(userWork).Error; err != nil {
		record(startTime, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Error)
		return err
	}
	record(startTime, metrics_lib.DatabaseMetricMethodDelete, metrics_lib.Success)

	return nil
}

// FindByUserId is one page of a reader's shelf, newest first, optionally for a
// single status.
func (r *UserWorkRepository) FindByUserId(ctx context.Context, userId string, status *string, page int, limit int) ([]*UserWork, int64, error) {
	startTime := time.Now()

	query := r.db.DB.WithContext(ctx).Where("user_id = ?", userId)
	countQuery := r.db.DB.WithContext(ctx).Model(&UserWork{}).Where("user_id = ?", userId)
	if status != nil {
		query = query.Where("status = ?", *status)
		countQuery = countQuery.Where("status = ?", *status)
	}

	var userWorks []*UserWork
	err := query.Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&userWorks).Error
	if err != nil {
		record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, 0, err
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, 0, err
	}

	record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return userWorks, total, nil
}

// FindByUserIdAndWorkId returns nil when there is no row, rather than the
// driver's not-found error. Not being on a shelf is the ordinary case.
func (r *UserWorkRepository) FindByUserIdAndWorkId(ctx context.Context, userId string, workId string) (*UserWork, error) {
	startTime := time.Now()

	var userWork UserWork
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND work_id = ?", userId, workId).
		First(&userWork).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)
			return nil, nil
		}
		record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)

		return nil, err
	}
	record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return &userWork, nil
}

// FindByUserIdAndWorkIds is what the dataloader calls, so a page of work cards
// costs one query rather than one per card.
func (r *UserWorkRepository) FindByUserIdAndWorkIds(ctx context.Context, userId string, workIds []string) ([]*UserWork, error) {
	startTime := time.Now()

	if len(workIds) == 0 {
		return []*UserWork{}, nil
	}

	// Batched for the same reason the anime side is: a very large IN list makes
	// Postgres pick a different plan, and the shelf page can ask for hundreds.
	const batchSize = 500
	var userWorks []*UserWork
	for i := 0; i < len(workIds); i += batchSize {
		end := i + batchSize
		if end > len(workIds) {
			end = len(workIds)
		}

		var batch []*UserWork
		err := r.db.DB.WithContext(ctx).
			Where("user_id = ? AND work_id IN ?", userId, workIds[i:end]).
			Find(&batch).Error
		if err != nil {
			record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
			return nil, err
		}
		userWorks = append(userWorks, batch...)
	}
	record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	return userWorks, nil
}

// CountByStatus returns how many works the reader has in each status, in one
// grouped query rather than a count per tab.
//
// Model(&UserWork{}) so gorm applies the soft-delete scope: a removed row must
// not be counted, exactly as it is not listed. The map is keyed by the stored
// status string; a status with no rows is simply absent, and the caller fills
// the zero.
func (r *UserWorkRepository) CountByStatus(ctx context.Context, userId string) (map[string]int64, error) {
	startTime := time.Now()

	type statusCount struct {
		Status string
		Count  int64
	}
	var rows []statusCount
	err := r.db.DB.WithContext(ctx).
		Model(&UserWork{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ? AND status IS NOT NULL", userId).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Error)
		return nil, err
	}

	record(startTime, metrics_lib.DatabaseMetricMethodSelect, metrics_lib.Success)

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.Count
	}

	return counts, nil
}
