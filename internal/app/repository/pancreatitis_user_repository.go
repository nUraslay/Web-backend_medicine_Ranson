package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ranson-backend/internal/app/ds"
)

func (r *Repository) GetUserByLogin(login string) (*ds.PancreatitisUser, error) {
	var user ds.PancreatitisUser

	err := r.db.Where("login = ?", login).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) AddUser(user *ds.PancreatitisUser) error {
	err := r.db.Create(user).Error
	if err != nil {
		return fmt.Errorf("ошибка при добавлении пользователя: %w", err)
	}

	return nil
}
