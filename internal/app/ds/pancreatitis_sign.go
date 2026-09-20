package ds

import "time"

const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

type PancreatitisSign struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:varchar(500)"`
	Status      string `gorm:"type:varchar(15);not null;default:'черновик'"`
	ImageURL    string `gorm:"column:image_url;type:varchar(255)"`
	VideoURL    string `gorm:"column:video_url;type:varchar(255)"`


	Stage          string  `gorm:"type:varchar(30)"`   
	ThresholdValue float64 `gorm:"type:numeric(10,2)"` 

	CreatedAt time.Time  `gorm:"not null"` 
	FormedAt  *time.Time 
	CreatorID uint       `gorm:"not null"`

	Creator PancreatitisUser `gorm:"foreignKey:CreatorID"`
}

func (PancreatitisSign) TableName() string {
	return "pancreatitis_signs"
}
