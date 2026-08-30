package user_work

import (
	"time"

	"gorm.io/gorm"
)

// UserWork is one row of a reader's shelf: a work, and where they are in it.
//
// Shaped like UserAnime because it is the same job for a different medium, and
// the frontend renders both through the same card. Where they differ they
// differ honestly -- Chapters and Volumes instead of Episodes, because that is
// what a manga release is counted in.
type UserWork struct {
	ID        string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    *string        `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	WorkID    *string        `gorm:"column:work_id;type:uuid;not null" json:"work_id"`
	Status    *string        `gorm:"column:status" json:"status"`
	Score     *float64       `gorm:"column:score" json:"score"`
	Chapters  *int           `gorm:"column:chapters" json:"chapters"`
	Volumes   *int           `gorm:"column:volumes" json:"volumes"`
	Tags      *string        `gorm:"column:tags" json:"tags"`
	ListID    *string        `gorm:"column:list_id" json:"list_id"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (UserWork) TableName() string {
	return "user_work"
}
