package dataloader

import (
	"context"
	"sync"
	"time"

	"strings"

	"github.com/weeb-vip/list-service/graph/model"
	user_work_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_work"
	"github.com/weeb-vip/list-service/internal/services/user_work"
)

// convertUserWorkToGraphql is a local copy rather than the resolvers package's.
// resolvers imports this package for the loaders, so importing it back would be
// a cycle -- which is why the anime loader keeps its own copy too.
func convertUserWorkToGraphql(entity *user_work_repo.UserWork) (*model.UserWork, error) {
	var status *model.WorkStatus
	if entity.Status != nil {
		value := model.WorkStatus(*entity.Status)
		status = &value
	}

	var tags []string
	if entity.Tags != nil && *entity.Tags != "" {
		tags = strings.Split(*entity.Tags, ",")
	}

	return &model.UserWork{
		ID:       entity.ID,
		UserID:   *entity.UserID,
		WorkID:   *entity.WorkID,
		Status:   status,
		Score:    entity.Score,
		Chapters: entity.Chapters,
		Volumes:  entity.Volumes,
		Tags:     tags,
		ListID:   entity.ListID,
	}, nil
}

type UserWorkKey struct {
	UserID string
	WorkID string
}

// userWorkBatch is one round of batched loads.
//
// A value rather than fields on the loader, for the reason the anime loader
// documents: a batch being executed must not be mutated by calls arriving for
// the next one.
type userWorkBatch struct {
	keys    []UserWorkKey
	results []*model.UserWork
	errs    []error
	done    chan struct{}
}

type UserWorkLoader struct {
	userWorkService user_work.UserWorkServiceImpl

	mutex   sync.Mutex
	current *userWorkBatch

	batchTimeout time.Duration
}

func NewUserWorkLoader(userWorkService user_work.UserWorkServiceImpl) *UserWorkLoader {
	return &UserWorkLoader{
		userWorkService: userWorkService,
		batchTimeout:    time.Millisecond * 16,
	}
}

// Load returns the viewer's row for one work, batching concurrent calls.
//
// Without this, a page of work cards costs one query per card. The wait selects
// on ctx.Done() so a cancelled request fails with its own error rather than
// sitting on a channel.
func (l *UserWorkLoader) Load(ctx context.Context, key UserWorkKey) (*model.UserWork, error) {
	b, index := l.enqueue(ctx, key)

	select {
	case <-b.done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if index < len(b.errs) && b.errs[index] != nil {
		return nil, b.errs[index]
	}
	if index < len(b.results) {
		return b.results[index], nil
	}

	return nil, nil
}

func (l *UserWorkLoader) enqueue(ctx context.Context, key UserWorkKey) (*userWorkBatch, int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.current == nil {
		l.current = &userWorkBatch{done: make(chan struct{})}
	}

	b := l.current
	b.keys = append(b.keys, key)
	index := len(b.keys) - 1

	// The first caller into a batch schedules its close.
	if index == 0 {
		go l.closeAfter(ctx, b)
	}

	return b, index
}

// closeAfter waits out the window, detaches the batch so later calls open a new
// one, then runs it. Detaching before the query matters: the query can take
// seconds, and anything arriving meanwhile must land in the next batch.
func (l *UserWorkLoader) closeAfter(ctx context.Context, b *userWorkBatch) {
	timer := time.NewTimer(l.batchTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		// Run it anyway: callers are waiting, and ctx.Err() through Load's
		// select is better than an unclosed batch.
	}

	l.mutex.Lock()
	if l.current == b {
		l.current = nil
	}
	l.mutex.Unlock()

	l.execute(ctx, b)
}

func (l *UserWorkLoader) execute(ctx context.Context, b *userWorkBatch) {
	defer close(b.done)

	b.results = make([]*model.UserWork, len(b.keys))
	b.errs = make([]error, len(b.keys))

	// Grouped by user, so each user costs one query with an IN list.
	byUser := make(map[string][]string, 1)
	for _, key := range b.keys {
		byUser[key.UserID] = append(byUser[key.UserID], key.WorkID)
	}

	for userID, workIDs := range byUser {
		userWorks, err := l.userWorkService.FindByUserIdAndWorkIds(ctx, userID, workIDs)
		if err != nil {
			for i, key := range b.keys {
				if key.UserID == userID {
					b.errs[i] = err
				}
			}

			continue
		}

		found := make(map[string]*user_work_repo.UserWork, len(userWorks))
		for _, uw := range userWorks {
			if uw.WorkID != nil {
				found[*uw.WorkID] = uw
			}
		}

		for i, key := range b.keys {
			if key.UserID != userID {
				continue
			}

			uw, exists := found[key.WorkID]
			if !exists {
				// Not on the user's shelf. nil is the answer, not an error.
				continue
			}

			converted, convertErr := convertUserWorkToGraphql(uw)
			if convertErr != nil {
				b.errs[i] = convertErr

				continue
			}

			b.results[i] = converted
		}
	}
}
