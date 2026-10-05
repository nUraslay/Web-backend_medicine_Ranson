package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ranson-backend/internal/app/ds"
)

var ErrSignNotFound = errors.New("признак не найден или недоступен для этого действия")

func (r *Repository) firstOrNil(query *gorm.DB) (*ds.PancreatitisSign, error) {
	var sign ds.PancreatitisSign

	err := query.First(&sign).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &sign, nil
}

func (r *Repository) GetPublishedSigns(title string, maxThreshold *float64) ([]ds.PancreatitisSign, error) {
	var signs []ds.PancreatitisSign

	query := r.db.Where("status = ?", ds.StatusPublished)
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if maxThreshold != nil {
		query = query.Where("threshold_value <= ?", *maxThreshold)
	}

	err := query.Order("id").Find(&signs).Error
	if err != nil {
		return nil, err
	}

	return signs, nil
}

func (r *Repository) GetMaxThreshold() (float64, error) {
	var max sql.NullFloat64

	err := r.db.Model(&ds.PancreatitisSign{}).
		Where("status = ?", ds.StatusPublished).
		Select("MAX(threshold_value)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}

	return max.Float64, nil // если записей нет, будет 0
}

func (r *Repository) GetPublishedSignByID(id int) (*ds.PancreatitisSign, error) {
	return r.firstOrNil(r.db.Where("id = ? AND status = ?", id, ds.StatusPublished))
}

// GetFirstPublishedSign возвращает первый опубликованный признак (ORM)
func (r *Repository) GetFirstPublishedSign() (*ds.PancreatitisSign, error) {
	return r.firstOrNil(r.db.Where("status = ?", ds.StatusPublished))
}

// GetNextPublishedSign возвращает следующий за currentID опубликованный признак. (ORM)
func (r *Repository) GetNextPublishedSign(currentID int) (*ds.PancreatitisSign, error) {
	next, err := r.firstOrNil(r.db.Where("status = ? AND id > ?", ds.StatusPublished, currentID))
	if err != nil {
		return nil, err
	}
	if next != nil {
		return next, nil
	}

	return r.GetFirstPublishedSign()
}

// GetDraftSign возвращает черновик пользователя или nil, если черновика нет (ORM)
func (r *Repository) GetDraftSign(userID uint) (*ds.PancreatitisSign, error) {
	return r.firstOrNil(r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft))
}

// likesCountRow — строка результата подсчета лайков по признакам
type likesCountRow struct {
	SignID uint
	Count  int
}

// GetLikesCounts возвращает количество лайков для каждого признака (ORM)
func (r *Repository) GetLikesCounts() (map[uint]int, error) {
	var rows []likesCountRow

	err := r.db.Model(&ds.PancreatitisSignLike{}).
		Select("sign_id, COUNT(*) AS count").
		Group("sign_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int, len(rows))
	for _, row := range rows {
		counts[row.SignID] = row.Count
	}

	return counts, nil
}

// GetLikesCount возвращает количество лайков одного признака (ORM)
func (r *Repository) GetLikesCount(signID uint) (int, error) {
	var count int64

	err := r.db.Model(&ds.PancreatitisSignLike{}).Where("sign_id = ?", signID).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// CreateDraftSign создает черновик признака: указаны название и (необязательно) загруженные фото/видео,
// остальные поля заполняются при публикации. Пустые imageURL/videoURL означают "использовать файлы по умолчанию" (ORM)
func (r *Repository) CreateDraftSign(title string, imageURL string, videoURL string, creatorID uint) (*ds.PancreatitisSign, error) {
	sign := ds.PancreatitisSign{
		Title:     title,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		Status:    ds.StatusDraft,
		CreatorID: creatorID,
	}

	err := r.db.Create(&sign).Error
	if err != nil {
		return nil, err
	}

	return &sign, nil
}

// PublishSign заполняет черновик и публикует его: меняет статус на "опубликован" и ставит дату формирования (ORM).
func (r *Repository) PublishSign(id uint, title string, description string, stage string, thresholdValue float64) error {
	result := r.db.Model(&ds.PancreatitisSign{}).
		Where("id = ? AND status = ?", id, ds.StatusDraft).
		Updates(map[string]interface{}{
			"title":           title,
			"description":     description,
			"stage":           stage,
			"threshold_value": thresholdValue,
			"status":          ds.StatusPublished,
			"formed_at":       time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSignNotFound
	}

	return nil
}

// DeleteSign логически удаляет опубликованный признак: меняет статус на "удален".Выполняется обычным SQL-запросом UPDATE, без ORM
func (r *Repository) DeleteSign(id uint) error {
	result := r.db.Exec(
		"UPDATE pancreatitis_signs SET status = ? WHERE id = ? AND status = ?",
		ds.StatusDeleted, id, ds.StatusPublished,
	)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении признака с id %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSignNotFound
	}

	return nil
}
