package dummy

import (
	"time"
)

type Dummy struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Ranking   *int      `gorm:"column:ranking;null" json:"ranking"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// set table name
func (Dummy) TableName() string {
	return "dummy"
}
