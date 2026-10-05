package ds

type PancreatitisUser struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"type:varchar(25);not null;unique" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;not null;default:false" json:"is_moderator"`
}

func (PancreatitisUser) TableName() string {
	return "pancreatitis_users"
}
