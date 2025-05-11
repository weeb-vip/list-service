package user_list

import (
	"gorm.io/gorm"
	"time"
)

type UserList struct {
	ID          string         `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      *string        `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Name        *string        `gorm:"column:name;not null" json:"name"`
	Description *string        `gorm:"column:description" json:"description"`
	Tags        *string        `gorm:"column:tags" json:"tags"`
	IsPublic    *bool          `gorm:"column:is_public" json:"is_public"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// set table name
func (UserList) TableName() string {
	return "user_list"
}
