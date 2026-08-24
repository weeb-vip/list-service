package dataloader

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	user_anime_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
)

// fakeService records every batch query it is asked to run, so a test can assert
// on how many queries a set of Loads produced rather than only on their results.
type fakeService struct {
	user_anime.UserAnimeServiceImpl

	mu      sync.Mutex
	calls   [][]string
	delay   time.Duration
	failErr error
}

func (f *fakeService) FindByUserIdAndAnimeIds(ctx context.Context, userID string, animeIDs []string) ([]*user_anime_repo.UserAnime, error) {
	f.mu.Lock()
	ids := append([]string(nil), animeIDs...)
	f.calls = append(f.calls, ids)
	delay, failErr := f.delay, f.failErr
	f.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}
	if failErr != nil {
		return nil, failErr
	}

	out := make([]*user_anime_repo.UserAnime, 0, len(ids))
	for _, id := range ids {
		animeID, uid := id, userID
		out = append(out, &user_anime_repo.UserAnime{ID: "ua-" + id, UserID: &uid, AnimeID: &animeID})
	}

	return out, nil
}

func (f *fakeService) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.calls)
}

func newTestLoader(svc user_anime.UserAnimeServiceImpl) *UserAnimeLoader {
	l := NewUserAnimeLoader(svc)
	// Short window so the tests do not wait on the production 16ms more than
	// they must; the behaviour under test is the batch boundary, not its length.
	l.batchTimeout = 10 * time.Millisecond

	return l
}

// TestConcurrentLoadsShareOneQuery is the property the loader exists for: N
// fields resolving together must cost one query, not N.
func TestConcurrentLoadsShareOneQuery(t *testing.T) {
	svc := &fakeService{}
	l := newTestLoader(svc)

	const n = 20
	var wg sync.WaitGroup
	results := make([]*struct{ id string }, n)
	_ = results

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got, err := l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: string(rune('a' + i))})
			if err != nil {
				t.Errorf("load %d: %v", i, err)

				return
			}
			if got == nil {
				t.Errorf("load %d: nil result", i)
			}
		}(i)
	}
	wg.Wait()

	if got := svc.callCount(); got != 1 {
		t.Fatalf("expected 20 concurrent loads to batch into 1 query, got %d", got)
	}
}

// TestSecondWaveAfterBatchCloses is the regression test.
//
// The previous implementation guarded its single batch with sync.Once and the
// loader is per request, so anything loaded after the first batch closed waited
// on a channel that would never be closed. This test hangs on that code and
// passes on this one.
//
// It is the shape a federated query produces: getHomePageData selects userAnime
// under both topRatedAnime and newestAnime, which resolves in more than one
// wave.
func TestSecondWaveAfterBatchCloses(t *testing.T) {
	svc := &fakeService{}
	l := newTestLoader(svc)

	first, err := l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: "a1"})
	if err != nil {
		t.Fatalf("first wave: %v", err)
	}
	if first == nil {
		t.Fatal("first wave returned nil")
	}

	// Well past the batch window, so the first batch has definitely closed.
	time.Sleep(30 * time.Millisecond)

	done := make(chan struct{})
	var second *struct{ err error }
	go func() {
		defer close(done)
		got, err := l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: "a2"})
		if err != nil {
			second = &struct{ err error }{err}

			return
		}
		if got == nil {
			second = &struct{ err error }{errors.New("nil result")}
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("second wave never completed: the batch after the first one was never executed")
	}

	if second != nil {
		t.Fatalf("second wave: %v", second.err)
	}

	if got := svc.callCount(); got != 2 {
		t.Fatalf("expected two waves to produce 2 queries, got %d", got)
	}
}

// TestLoadDuringInFlightBatchStillResolves covers the widest version of the same
// gap: a load arriving while the previous batch's query is still running.
//
// Against a slow database that window is seconds wide, which is when the old
// code was most likely to strand a caller.
func TestLoadDuringInFlightBatchStillResolves(t *testing.T) {
	svc := &fakeService{delay: 150 * time.Millisecond}
	l := newTestLoader(svc)

	go func() { _, _ = l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: "a1"}) }()

	// Land inside the first batch's query rather than inside its window.
	time.Sleep(50 * time.Millisecond)

	done := make(chan error, 1)
	go func() {
		_, err := l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: "a2"})
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("load during in-flight batch: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a load arriving during an in-flight batch never resolved")
	}
}

// TestLoadRespectsContextCancellation: a caller whose request is gone should get
// the context error, not wait on a batch it no longer cares about.
func TestLoadRespectsContextCancellation(t *testing.T) {
	svc := &fakeService{delay: time.Second}
	l := newTestLoader(svc)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := l.Load(ctx, UserAnimeKey{UserID: "u1", AnimeID: "a1"})
		done <- err
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Load ignored context cancellation")
	}
}

// TestBatchErrorReachesEveryCaller: when the query fails, every caller in that
// batch must see the error rather than a silent nil.
func TestBatchErrorReachesEveryCaller(t *testing.T) {
	wantErr := errors.New("boom")
	svc := &fakeService{failErr: wantErr}
	l := newTestLoader(svc)

	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := l.Load(context.Background(), UserAnimeKey{UserID: "u1", AnimeID: string(rune('a' + i))})
			errs[i] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, wantErr) {
			t.Errorf("caller %d got %v, want %v", i, err, wantErr)
		}
	}
}
