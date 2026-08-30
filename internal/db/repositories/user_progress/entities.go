// Package user_progress records which individual episodes and chapters a user
// has finished, as opposed to how many.
//
// One package for both because they are the same shape and the same queries --
// mark, unmark, list for one title, count for one title -- differing only in the
// column that names the unit. Splitting them would mean maintaining two copies
// of identical logic, which is how the anime and work schemas already drifted
// apart on status casing.
package user_progress

import (
	"time"

	"gorm.io/gorm"
)

// UserAnimeEpisode is one episode a user has watched.
//
// Identified by number rather than by the episode row's id, deliberately: the
// scraper reinserts an anime's episodes with fresh ids, and pointing at those
// would discard every user's history each time that happens.
type UserAnimeEpisode struct {
	ID            string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string         `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	AnimeID       string         `gorm:"column:anime_id;type:uuid;not null" json:"anime_id"`
	EpisodeNumber int            `gorm:"column:episode_number;not null" json:"episode_number"`
	WatchedAt     time.Time      `gorm:"column:watched_at" json:"watched_at"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (UserAnimeEpisode) TableName() string {
	return "user_anime_episode"
}

// UserWorkChapter is one chapter a user has read. Numbered for the same reason
// and one more: works have no chapter rows to point at.
type UserWorkChapter struct {
	ID            string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string         `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	WorkID        string         `gorm:"column:work_id;type:uuid;not null" json:"work_id"`
	ChapterNumber int            `gorm:"column:chapter_number;not null" json:"chapter_number"`
	ReadAt        time.Time      `gorm:"column:read_at" json:"read_at"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (UserWorkChapter) TableName() string {
	return "user_work_chapter"
}
