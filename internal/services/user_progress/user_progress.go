package user_progress

import (
	"context"
	"time"

	progress_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_progress"
	user_anime_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	user_work_repo "github.com/weeb-vip/list-service/internal/db/repositories/user_work"
)

type UserProgressServiceImpl interface {
	MarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int, watchedAt *time.Time) (*progress_repo.UserAnimeEpisode, error)
	UnmarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int) error
	WatchedEpisodes(ctx context.Context, userID string, animeID string) ([]*progress_repo.UserAnimeEpisode, error)

	MarkChapter(ctx context.Context, userID string, workID string, chapterNumber int, readAt *time.Time) (*progress_repo.UserWorkChapter, error)
	UnmarkChapter(ctx context.Context, userID string, workID string, chapterNumber int) error
	ReadChapters(ctx context.Context, userID string, workID string) ([]*progress_repo.UserWorkChapter, error)
}

type UserProgressService struct {
	Progress  progress_repo.UserProgressRepositoryImpl
	UserAnime user_anime_repo.UserAnimeRepositoryImpl
	UserWork  user_work_repo.UserWorkRepositoryImpl
}

func NewUserProgressService(
	progress progress_repo.UserProgressRepositoryImpl,
	userAnime user_anime_repo.UserAnimeRepositoryImpl,
	userWork user_work_repo.UserWorkRepositoryImpl,
) UserProgressServiceImpl {
	return &UserProgressService{Progress: progress, UserAnime: userAnime, UserWork: userWork}
}

// MarkEpisode records one episode and keeps user_anime.episodes in step.
//
// Two representations of the same fact now exist -- a count on user_anime and a
// row per episode -- and the rule is that rows win once any exist. Everything
// that reads the count today keeps working, without it drifting away from what
// the ticks say.
//
// The subtlety is the first tick. A user with "9 watched" and no rows would
// drop to 1, because the count starts being derived the moment a row appears.
// So their existing count is materialised as rows 1..9 first, and the count is
// unchanged by their own action. It only happens once: after that there are
// rows, and nothing to backfill.
func (s *UserProgressService) MarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int, watchedAt *time.Time) (*progress_repo.UserAnimeEpisode, error) {
	if err := s.backfillEpisodesFromCount(ctx, userID, animeID); err != nil {
		return nil, err
	}

	entry, err := s.Progress.MarkEpisode(ctx, userID, animeID, episodeNumber, watchedAt)
	if err != nil {
		return nil, err
	}

	if err := s.syncEpisodeCount(ctx, userID, animeID); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *UserProgressService) UnmarkEpisode(ctx context.Context, userID string, animeID string, episodeNumber int) error {
	// Backfill here too. Unticking episode 3 of a legacy "9 watched" should
	// leave eight, not zero.
	if err := s.backfillEpisodesFromCount(ctx, userID, animeID); err != nil {
		return err
	}

	if err := s.Progress.UnmarkEpisode(ctx, userID, animeID, episodeNumber); err != nil {
		return err
	}

	return s.syncEpisodeCount(ctx, userID, animeID)
}

func (s *UserProgressService) WatchedEpisodes(ctx context.Context, userID string, animeID string) ([]*progress_repo.UserAnimeEpisode, error) {
	return s.Progress.WatchedEpisodes(ctx, userID, animeID)
}

// backfillEpisodesFromCount turns a legacy count into the rows it implied, once.
//
// Skipped when rows already exist, which is the ordinary case after the first
// interaction, and when the anime is not on the user's list at all -- there is
// no count to preserve.
func (s *UserProgressService) backfillEpisodesFromCount(ctx context.Context, userID string, animeID string) error {
	existing, err := s.Progress.CountWatchedEpisodes(ctx, userID, animeID)
	if err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	userAnime, err := s.UserAnime.FindByUserIdAndAnimeId(ctx, userID, animeID)
	if err != nil || userAnime == nil || userAnime.Episodes == nil || *userAnime.Episodes <= 0 {
		// A missing row is not an error here: the user may be ticking an episode
		// of something they have not added, and the mark itself still stands.
		return nil
	}

	return s.Progress.MarkEpisodesUpTo(ctx, userID, animeID, *userAnime.Episodes)
}

// syncEpisodeCount writes the derived count back to user_anime.
//
// Only when the anime is already on the user's list. Ticking an episode is not
// the same as adding a show, and inventing a list entry here would put things on
// people's lists they never added.
func (s *UserProgressService) syncEpisodeCount(ctx context.Context, userID string, animeID string) error {
	userAnime, err := s.UserAnime.FindByUserIdAndAnimeId(ctx, userID, animeID)
	if err != nil || userAnime == nil {
		return nil
	}

	total, err := s.Progress.CountWatchedEpisodes(ctx, userID, animeID)
	if err != nil {
		return err
	}

	count := int(total)
	if userAnime.Episodes != nil && *userAnime.Episodes == count {
		return nil
	}
	userAnime.Episodes = &count

	_, err = s.UserAnime.Upsert(ctx, userAnime)

	return err
}

func (s *UserProgressService) MarkChapter(ctx context.Context, userID string, workID string, chapterNumber int, readAt *time.Time) (*progress_repo.UserWorkChapter, error) {
	if err := s.backfillChaptersFromCount(ctx, userID, workID); err != nil {
		return nil, err
	}

	entry, err := s.Progress.MarkChapter(ctx, userID, workID, chapterNumber, readAt)
	if err != nil {
		return nil, err
	}

	if err := s.syncChapterCount(ctx, userID, workID); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *UserProgressService) UnmarkChapter(ctx context.Context, userID string, workID string, chapterNumber int) error {
	if err := s.backfillChaptersFromCount(ctx, userID, workID); err != nil {
		return err
	}

	if err := s.Progress.UnmarkChapter(ctx, userID, workID, chapterNumber); err != nil {
		return err
	}

	return s.syncChapterCount(ctx, userID, workID)
}

func (s *UserProgressService) ReadChapters(ctx context.Context, userID string, workID string) ([]*progress_repo.UserWorkChapter, error) {
	return s.Progress.ReadChapters(ctx, userID, workID)
}

// backfillChaptersFromCount is the episode backfill for works. It exists on day
// one because user_work.chapters shipped as a count before this did, so the same
// legacy shape is already possible.
func (s *UserProgressService) backfillChaptersFromCount(ctx context.Context, userID string, workID string) error {
	existing, err := s.Progress.CountReadChapters(ctx, userID, workID)
	if err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	userWork, err := s.UserWork.FindByUserIdAndWorkId(ctx, userID, workID)
	if err != nil || userWork == nil || userWork.Chapters == nil || *userWork.Chapters <= 0 {
		return nil
	}

	for number := 1; number <= *userWork.Chapters; number++ {
		if _, err := s.Progress.MarkChapter(ctx, userID, workID, number, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserProgressService) syncChapterCount(ctx context.Context, userID string, workID string) error {
	userWork, err := s.UserWork.FindByUserIdAndWorkId(ctx, userID, workID)
	if err != nil || userWork == nil {
		return nil
	}

	total, err := s.Progress.CountReadChapters(ctx, userID, workID)
	if err != nil {
		return err
	}

	count := int(total)
	if userWork.Chapters != nil && *userWork.Chapters == count {
		return nil
	}
	userWork.Chapters = &count

	_, err = s.UserWork.Upsert(ctx, userWork)

	return err
}
