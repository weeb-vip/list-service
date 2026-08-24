package dataloader

import (
	"context"
	"github.com/weeb-vip/list-service/graph/model"
	user_anime_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"strings"
	"sync"
	"time"
)

type UserAnimeKey struct {
	UserID  string
	AnimeID string
}

// userAnimeBatch is one round of batched loads.
//
// It is a value rather than fields on the loader so a batch that is being
// executed cannot be mutated by calls arriving for the next one -- the bug the
// previous single-set-of-fields design had.
type userAnimeBatch struct {
	keys    []UserAnimeKey
	results []*model.UserAnime
	errs    []error
	done    chan struct{}
}

type UserAnimeLoader struct {
	userAnimeService user_anime.UserAnimeServiceImpl

	mutex   sync.Mutex
	current *userAnimeBatch

	batchTimeout time.Duration
}

func NewUserAnimeLoader(userAnimeService user_anime.UserAnimeServiceImpl) *UserAnimeLoader {
	return &UserAnimeLoader{
		userAnimeService: userAnimeService,
		batchTimeout:     time.Millisecond * 16, // Small timeout to batch requests
	}
}

// convertUserAnimeToGraphql converts UserAnime entity to GraphQL model
func convertUserAnimeToGraphql(userAnimeEntity *user_anime_repo.UserAnime) (*model.UserAnime, error) {
	var status *model.Status
	if userAnimeEntity.Status != nil {
		statuss := model.Status(*userAnimeEntity.Status)
		status = &statuss
	} else {
		status = nil
	}

	var tags []string
	if userAnimeEntity.Tags != nil {
		tags = strings.Split(*userAnimeEntity.Tags, ",")
	} else {
		tags = nil
	}

	return &model.UserAnime{
		ID:                 userAnimeEntity.ID,
		UserID:             *userAnimeEntity.UserID,
		AnimeID:            *userAnimeEntity.AnimeID,
		Status:             status,
		Score:              userAnimeEntity.Score,
		Episodes:           userAnimeEntity.Episodes,
		Rewatching:         userAnimeEntity.Rewatching,
		RewatchingEpisodes: userAnimeEntity.RewatchingEpisodes,
		Tags:               tags,
		ListID:             userAnimeEntity.ListID,
	}, nil
}

// Load returns the user's entry for one anime, batching concurrent calls into
// a single query.
//
// Calls that arrive close together share a batch. A batch closes batchTimeout
// after its first call, runs one query, and the next Load after that opens a
// fresh one -- a request may go through as many batches as it has waves of
// field resolution.
//
// That last part is the fix for a hang. The previous implementation guarded
// executeBatch with a sync.Once and the loader is created per request, so a
// request got exactly one batch for its lifetime. Anything resolved after that
// batch closed was appended to the pending slice, handed a channel, and then
// waited on a channel nothing would ever close:
//
//	<-resultChan   // no ctx, no timeout, no second batch
//
// The goroutine blocked until the whole request was torn down, logging nothing.
// A federated query selecting userAnime under two roots -- getHomePageData asks
// for it under both topRatedAnime and newestAnime -- is exactly the shape that
// resolves in more than one wave, and the 16ms window makes it a race rather
// than a reliable failure.
//
// The wait also selects on ctx.Done() now. A caller whose request is cancelled
// or timed out should fail with that error rather than sit on a channel.
func (l *UserAnimeLoader) Load(ctx context.Context, key UserAnimeKey) (*model.UserAnime, error) {
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

// enqueue adds key to the open batch, opening one if there is none, and returns
// the batch it joined along with its position in it.
func (l *UserAnimeLoader) enqueue(ctx context.Context, key UserAnimeKey) (*userAnimeBatch, int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.current == nil {
		l.current = &userAnimeBatch{done: make(chan struct{})}
	}

	b := l.current
	b.keys = append(b.keys, key)
	index := len(b.keys) - 1

	// The first caller into a batch is the one that schedules its close.
	if index == 0 {
		go l.closeAfter(ctx, b)
	}

	return b, index
}

// closeAfter waits for the batch window, detaches the batch so later calls open
// a new one, and then runs it.
//
// Detaching before the query matters: the query can take seconds against a busy
// database, and anything arriving in that time must land in the next batch
// rather than in one that has already been copied and is on its way out.
func (l *UserAnimeLoader) closeAfter(ctx context.Context, b *userAnimeBatch) {
	timer := time.NewTimer(l.batchTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		// Still run it: callers are waiting, and giving them ctx.Err() through
		// the select in Load is better than leaving the batch unclosed.
	}

	l.mutex.Lock()
	if l.current == b {
		l.current = nil
	}
	l.mutex.Unlock()

	l.execute(ctx, b)
}

// execute runs one query per user in the batch and fills in every caller's slot,
// then releases them all by closing done.
func (l *UserAnimeLoader) execute(ctx context.Context, b *userAnimeBatch) {
	defer close(b.done)

	b.results = make([]*model.UserAnime, len(b.keys))
	b.errs = make([]error, len(b.keys))

	// Grouped by user so each user costs one query with an IN list, rather than
	// one query per anime.
	byUser := make(map[string][]string, 1)
	for _, key := range b.keys {
		byUser[key.UserID] = append(byUser[key.UserID], key.AnimeID)
	}

	for userID, animeIDs := range byUser {
		userAnimes, err := l.userAnimeService.FindByUserIdAndAnimeIds(ctx, userID, animeIDs)
		if err != nil {
			for i, key := range b.keys {
				if key.UserID == userID {
					b.errs[i] = err
				}
			}

			continue
		}

		found := make(map[string]*user_anime_repo.UserAnime, len(userAnimes))
		for _, ua := range userAnimes {
			if ua.AnimeID != nil {
				found[*ua.AnimeID] = ua
			}
		}

		for i, key := range b.keys {
			if key.UserID != userID {
				continue
			}

			ua, exists := found[key.AnimeID]
			if !exists {
				// Not in the user's list. nil is the answer, not an error.
				continue
			}

			converted, convertErr := convertUserAnimeToGraphql(ua)
			if convertErr != nil {
				b.errs[i] = convertErr

				continue
			}

			b.results[i] = converted
		}
	}
}
