package dataloader

import (
	"context"
	"sync"
	"time"
	"strings"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	user_anime_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/graph/model"
)

type UserAnimeKey struct {
	UserID  string
	AnimeID string
}

type UserAnimeLoader struct {
	userAnimeService user_anime.UserAnimeServiceImpl
	batch           []UserAnimeKey
	batchResult     []*model.UserAnime
	batchError      []error
	batchOnce       sync.Once
	batchChannels   []chan struct{}
	mutex           sync.Mutex
	batchTimeout    time.Duration
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

// Load loads a single user anime, batching the request with others
func (l *UserAnimeLoader) Load(ctx context.Context, key UserAnimeKey) (*model.UserAnime, error) {
	l.mutex.Lock()
	
	// Add this key to the batch
	index := len(l.batch)
	l.batch = append(l.batch, key)
	
	// Create a channel to wait for the batch result
	resultChan := make(chan struct{})
	l.batchChannels = append(l.batchChannels, resultChan)
	
	// If this is the first request in the batch, start the timer
	if index == 0 {
		go l.startBatchTimer(ctx)
	}
	
	l.mutex.Unlock()
	
	// Wait for the batch to complete
	<-resultChan
	
	// Return the result for this specific key
	if index < len(l.batchError) && l.batchError[index] != nil {
		return nil, l.batchError[index]
	}
	
	if index < len(l.batchResult) {
		return l.batchResult[index], nil
	}
	
	return nil, nil
}

func (l *UserAnimeLoader) startBatchTimer(ctx context.Context) {
	time.Sleep(l.batchTimeout)
	l.executeBatch(ctx)
}

func (l *UserAnimeLoader) executeBatch(ctx context.Context) {
	l.batchOnce.Do(func() {
		l.mutex.Lock()
		keys := make([]UserAnimeKey, len(l.batch))
		copy(keys, l.batch)
		channels := make([]chan struct{}, len(l.batchChannels))
		copy(channels, l.batchChannels)
		l.mutex.Unlock()
		
		// Group keys by userID for efficient querying
		userAnimeMap := make(map[string][]string) // userID -> animeIDs
		keyToIndex := make(map[UserAnimeKey]int)  // key -> index in batch
		
		for i, key := range keys {
			userAnimeMap[key.UserID] = append(userAnimeMap[key.UserID], key.AnimeID)
			keyToIndex[key] = i
		}
		
		results := make([]*model.UserAnime, len(keys))
		errors := make([]error, len(keys))
		
		// Process each user's anime IDs
		for userID, animeIDs := range userAnimeMap {
			userAnimes, err := l.userAnimeService.FindByUserIdAndAnimeIds(ctx, userID, animeIDs)
			if err != nil {
				// Set error for all keys belonging to this user
				for i, key := range keys {
					if key.UserID == userID {
						errors[i] = err
					}
				}
				continue
			}
			
			// Create a map for quick lookup
			userAnimeByAnimeID := make(map[string]*user_anime_repo.UserAnime)
			for _, ua := range userAnimes {
				if ua.AnimeID != nil {
					userAnimeByAnimeID[*ua.AnimeID] = ua
				}
			}
			
			// Fill results in the order of original keys
			for i, key := range keys {
				if key.UserID == userID {
					if userAnime, exists := userAnimeByAnimeID[key.AnimeID]; exists {
						converted, convertErr := convertUserAnimeToGraphql(userAnime)
						if convertErr != nil {
							errors[i] = convertErr
						} else {
							results[i] = converted
						}
					}
					// If userAnime doesn't exist, results[i] stays nil (which is correct)
				}
			}
		}
		
		l.batchResult = results
		l.batchError = errors
		
		// Notify all waiting goroutines
		for _, ch := range channels {
			close(ch)
		}
	})
}