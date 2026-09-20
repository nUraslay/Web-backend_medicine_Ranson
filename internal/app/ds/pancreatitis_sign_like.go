package ds


type PancreatitisSignLike struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;uniqueIndex:idx_pancreatitis_sign_like"`
	SignID uint `gorm:"not null;uniqueIndex:idx_pancreatitis_sign_like"`

	User PancreatitisUser `gorm:"foreignKey:UserID"`
	Sign PancreatitisSign `gorm:"foreignKey:SignID"`
}

func (PancreatitisSignLike) TableName() string {
	return "pancreatitis_sign_likes"
}
