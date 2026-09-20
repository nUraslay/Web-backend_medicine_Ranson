package ds

type PancreatitisUser struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(25);not null;unique"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;not null;default:false"`
}

func (PancreatitisUser) TableName() string {
	return "pancreatitis_users"
}
