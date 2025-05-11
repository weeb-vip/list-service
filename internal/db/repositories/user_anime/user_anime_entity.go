package user_anime

import (
	"gorm.io/gorm"
	"time"
)

type UserAnime struct {
	ID                 string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             *string        `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	AnimeID            *string        `gorm:"column:anime_id;type:uuid;not null" json:"anime_id"`
	Status             *string        `gorm:"column:status" json:"status"`
	Score              *float64       `gorm:"column:score" json:"score"`
	Episodes           *int           `gorm:"column:episodes" json:"episodes"`
	Rewatching         *int           `gorm:"column:rewatching" json:"rewatching"`
	RewatchingEpisodes *int           `gorm:"column:rewatching_episodes" json:"rewatching_episodes"`
	Tags               *string        `gorm:"column:tags" json:"tags"`
	ListID             *string        `gorm:"column:list_id" json:"list_id"`
	CreatedAt          time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// set table name
func (UserAnime) TableName() string {
	return "user_anime"
}
