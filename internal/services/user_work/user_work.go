package user_work

import (
	"context"
	"strings"

	"github.com/weeb-vip/list-service/internal/db/repositories/user_work"
)

// UserWorkStatus is the reading equivalent of UserAnimeStatus.
//
// Its own type, and its own words. "Watching" a manga is wrong, and a shared
// enum would make the frontend translate WATCHING into "Reading" on one surface
// and not the other.
//
// Upper case, matching the GraphQL enum exactly. The anime side is ambiguous
// about this -- its constants are lower case while the resolver writes the enum
// value straight through, so the rows say WATCHING and migration 000003 is left
// rewriting a lower-case value nothing produces any more. One spelling here, and
// it is the one that reaches the column.
type UserWorkStatus string

const (
	Reading    UserWorkStatus = "READING"
	Completed  UserWorkStatus = "COMPLETED"
	OnHold     UserWorkStatus = "ONHOLD"
	Dropped    UserWorkStatus = "DROPPED"
	PlanToRead UserWorkStatus = "PLANTOREAD"
)

type UserWork struct {
	ID       *string         `json:"id"`
	UserID   string          `json:"user_id"`
	WorkID   string          `json:"work_id"`
	Status   *UserWorkStatus `json:"status"`
	Score    *float64        `json:"score"`
	Chapters *int            `json:"chapters"`
	Volumes  *int            `json:"volumes"`
	Tags     []string        `json:"tags"`
	ListID   *string         `json:"list_id"`
}

type UserWorkPaginated struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int         `json:"total"`
	Works []*UserWork `json:"works"`
}

type UserWorkServiceImpl interface {
	Upsert(ctx context.Context, userWork *UserWork) (*user_work.UserWork, error)
	Delete(ctx context.Context, userID string, workID string) error
	FindByUserId(ctx context.Context, userID string, status *string, page int, limit int) ([]*user_work.UserWork, int64, error)
	FindByUserIdAndWorkId(ctx context.Context, userID string, workID string) (*user_work.UserWork, error)
	FindByUserIdAndWorkIds(ctx context.Context, userID string, workIDs []string) ([]*user_work.UserWork, error)
}

type UserWorkService struct {
	Repository user_work.UserWorkRepositoryImpl
}

func NewUserWorkService(repository user_work.UserWorkRepositoryImpl) UserWorkServiceImpl {
	return &UserWorkService{Repository: repository}
}

func (s *UserWorkService) Upsert(ctx context.Context, userWork *UserWork) (*user_work.UserWork, error) {
	var id string
	if userWork.ID != nil {
		id = *userWork.ID
	}

	var status *string
	if userWork.Status != nil {
		value := string(*userWork.Status)
		status = &value
	}

	// Tags are stored as one comma-joined column, as they are for anime. Absent
	// rather than empty when there are none, so a row with no tags reads as null
	// instead of an empty string that later splits into [""].
	var tags *string
	if len(userWork.Tags) > 0 {
		joined := strings.Join(userWork.Tags, ",")
		tags = &joined
	}

	entity := &user_work.UserWork{
		ID:       id,
		UserID:   &userWork.UserID,
		WorkID:   &userWork.WorkID,
		Status:   status,
		Score:    userWork.Score,
		Chapters: userWork.Chapters,
		Volumes:  userWork.Volumes,
		Tags:     tags,
		ListID:   userWork.ListID,
	}

	return s.Repository.Upsert(ctx, entity)
}

// Delete takes the work id, not the row id, because that is what the caller
// holds: the page knows which manga it is showing.
//
// A row belonging to someone else is a silent no-op rather than an error, and
// the ownership check stays even though the lookup is already scoped by user --
// it costs nothing and it is the check that makes the intent obvious.
func (s *UserWorkService) Delete(ctx context.Context, userID string, workID string) error {
	userWork, err := s.Repository.FindByUserIdAndWorkId(ctx, userID, workID)
	if err != nil {
		return err
	}
	if userWork == nil || userWork.UserID == nil || *userWork.UserID != userID {
		return nil
	}

	return s.Repository.Delete(ctx, userWork)
}

func (s *UserWorkService) FindByUserId(ctx context.Context, userID string, status *string, page int, limit int) ([]*user_work.UserWork, int64, error) {
	return s.Repository.FindByUserId(ctx, userID, status, page, limit)
}

func (s *UserWorkService) FindByUserIdAndWorkId(ctx context.Context, userID string, workID string) (*user_work.UserWork, error) {
	return s.Repository.FindByUserIdAndWorkId(ctx, userID, workID)
}

func (s *UserWorkService) FindByUserIdAndWorkIds(ctx context.Context, userID string, workIDs []string) ([]*user_work.UserWork, error) {
	return s.Repository.FindByUserIdAndWorkIds(ctx, userID, workIDs)
}
