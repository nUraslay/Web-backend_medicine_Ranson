package repository

import (
	"fmt"
	"sort"
	"strings"
)

// MinioBaseURL — адрес, по которому раздаются картинки и видео из Minio.
const MinioBaseURL = "http://localhost:9000/ranson-media/"

type Status string

const (
	StatusDraft     Status = "черновик"
	StatusPublished Status = "опубликован"
	StatusDeleted   Status = "удален"
)

type PancreatitisSign struct {
	ID             int
	Title          string
	Description    string
	Category       string
	Threshold      float64
	ImageKey       string
	VideoKey       string
	Status         Status
	LikedByUserIDs []int
}

func (s PancreatitisSign) ImageURL() string {
	return MinioBaseURL + s.ImageKey
}

func (s PancreatitisSign) VideoURL() string {
	return MinioBaseURL + s.VideoKey
}

type Repository struct {
	signs []PancreatitisSign	
}

func NewRepository() (*Repository, error) {
	signs := []PancreatitisSign{
		{
			ID:             1,
			Title:          "Возраст старше 55 лет",
			Description:    "Возраст пациента на момент поступления свыше 55 лет является independent-предиктором неблагоприятного течения острого панкреатита за счет снижения компенсаторных резервов организма.",
			Category:       "При поступлении",
			Threshold:      55,
			ImageKey:       "age-admission.jpg",
			VideoKey:       "age-admission.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(842),
		},
		{
			ID:             2,
			Title:          "Лейкоциты > 16 х 10^9/л",
			Description:    "Лейкоцитоз выше 16 х 10^9/л при поступлении отражает выраженность системной воспалительной реакции и коррелирует с тяжестью деструктивных изменений поджелудочной железы.",
			Category:       "При поступлении",
			Threshold:      16,
			ImageKey:       "wbc-admission.jpg",
			VideoKey:       "wbc-admission.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(650),
		},
		{
			ID:             3,
			Title:          "Глюкоза крови > 11,1 ммоль/л",
			Description:    "Уровень глюкозы при поступлении свыше 11,1 ммоль/л является независимым предиктором неблагоприятного исхода и связан с некрозом островкового аппарата поджелудочной железы.",
			Category:       "При поступлении",
			Threshold:      11.1,
			ImageKey:       "glucose-admission.jpg",
			VideoKey:       "glucose-admission.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(204),
		},
		{
			ID:             4,
			Title:          "ЛДГ сыворотки > 350 МЕ/л",
			Description:    "Повышение лактатдегидрогеназы свыше 350 МЕ/л свидетельствует о цитолизе тканей поджелудочной железы и является одним из ранних биохимических маркеров тяжести процесса.",
			Category:       "При поступлении",
			Threshold:      350,
			ImageKey:       "ldh-admission.jpg",
			VideoKey:       "ldh-admission.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(388),
		},
		{
			ID:             5,
			Title:          "АСТ (АсАТ) > 250 МЕ/л",
			Description:    "Повышение аспартатаминотрансферазы свыше 250 МЕ/л при поступлении указывает на выраженное повреждение паренхиматозных клеток и коррелирует с тяжестью острого процесса.",
			Category:       "При поступлении",
			Threshold:      250,
			ImageKey:       "ast-admission.jpg",
			VideoKey:       "ast-admission.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(295),
		},
		{
			ID:             6,
			Title:          "Снижение гематокрита более чем на 10%",
			Description:    "Падение гематокрита более чем на 10% от исходного за первые 48 часов отражает секвестрацию жидкости и кровопотерю в зону воспаления, что ухудшает прогноз.",
			Category:       "Через 48 часов",
			Threshold:      10,
			ImageKey:       "hematocrit-48h.jpg",
			VideoKey:       "hematocrit-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(220),
		},
		{
			ID:             7,
			Title:          "Повышение мочевины крови > 1,8 ммоль/л",
			Description:    "Прирост мочевины крови более чем на 1,8 ммоль/л за 48 часов, несмотря на инфузионную терапию, отражает нарастающую почечную дисфункцию на фоне тяжелого панкреатита.",
			Category:       "Через 48 часов",
			Threshold:      1.8,
			ImageKey:       "bun-48h.jpg",
			VideoKey:       "bun-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(176),
		},
		{
			ID:             8,
			Title:          "Кальций сыворотки < 2 ммоль/л",
			Description:    "Снижение сывороточного кальция ниже 2 ммоль/л через 48 часов связано с жировым некрозом и омылением жиров в очаге воспаления, что говорит о тяжелом течении процесса.",
			Category:       "Через 48 часов",
			Threshold:      2,
			ImageKey:       "calcium-48h.jpg",
			VideoKey:       "calcium-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(410),
		},
		{
			ID:             9,
			Title:          "paO2 < 60 мм рт.ст.",
			Description:    "Снижение парциального давления кислорода в артериальной крови ниже 60 мм рт.ст. через 48 часов указывает на развитие дыхательной недостаточности как осложнения панкреатита.",
			Category:       "Через 48 часов",
			Threshold:      60,
			ImageKey:       "pao2-48h.jpg",
			VideoKey:       "pao2-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(198),
		},
		{
			ID:             10,
			Title:          "Дефицит оснований > 4 мЭкв/л",
			Description:    "Дефицит оснований более 4 мЭкв/л через 48 часов свидетельствует о метаболическом ацидозе и тканевой гипоперфузии на фоне системной воспалительной реакции.",
			Category:       "Через 48 часов",
			Threshold:      4,
			ImageKey:       "base-deficit-48h.jpg",
			VideoKey:       "base-deficit-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(143),
		},
		{
			ID:             11,
			Title:          "Секвестрация жидкости > 6000 мл",
			Description:    "Секвестрация более 6000 мл жидкости за первые 48 часов отражает выраженность третьего пространства и коррелирует с тяжестью пареза кишечника и забрюшинной клетчатки.",
			Category:       "Через 48 часов",
			Threshold:      6000,
			ImageKey:       "fluid-sequestration-48h.jpg",
			VideoKey:       "fluid-sequestration-48h.mp4",
			Status:         StatusPublished,
			LikedByUserIDs: makeIDs(167),
		},
		{
			ID:             12,
			Title:          "С-реактивный белок > 150 мг/л",
			Description:    "Уровень СРБ выше 150 мг/л через 48 часов как дополнительный маркер тяжести системного воспаления. Требует проверки перед публикацией.",
			Category:       "Через 48 часов",
			Threshold:      150,
			ImageKey:       "crp-draft.jpg",
			VideoKey:       "crp-draft.mp4",
			Status:         StatusDraft,
			LikedByUserIDs: nil,
		},

		{
			ID:             13,
			Title:          "Амилаза крови",
			Description:    "Исключен из актуальной шкалы Рэнсона как неспецифичный показатель, не коррелирующий напрямую с тяжестью течения заболевания.",
			Category:       "При поступлении",
			Threshold:      0,
			ImageKey:       "amylase-deleted.jpg",
			VideoKey:       "amylase-deleted.mp4",
			Status:         StatusDeleted,
			LikedByUserIDs: nil,
		},
	}

	return &Repository{signs: signs}, nil
}

func makeIDs(n int) []int {
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}
	return ids
}

func (r *Repository) GetPublishedSigns() ([]PancreatitisSign, error) {
	var result []PancreatitisSign
	for _, s := range r.signs {
		if s.Status == StatusPublished {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("нет опубликованных признаков панкреатита")
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *Repository) GetPublishedSignByID(id int) (PancreatitisSign, error) {
	for _, s := range r.signs {
		if s.ID == id && s.Status == StatusPublished {
			return s, nil
		}
	}
	return PancreatitisSign{}, fmt.Errorf("опубликованный признак с id %d не найден", id)
}

func (r *Repository) GetNextPublishedSign(currentID int) (PancreatitisSign, error) {
	signs, err := r.GetPublishedSigns()
	if err != nil {
		return PancreatitisSign{}, err
	}

	for i, s := range signs {
		if s.ID == currentID {
			nextIndex := (i + 1) % len(signs)
			return signs[nextIndex], nil
		}
	}
	return PancreatitisSign{}, fmt.Errorf("признак с id %d не найден среди опубликованных", currentID)
}

func (r *Repository) GetDraftSign() (PancreatitisSign, error) {
	for _, s := range r.signs {
		if s.Status == StatusDraft {
			return s, nil
		}
	}
	return PancreatitisSign{}, fmt.Errorf("черновик не найден")
}

func (r *Repository) GetPublishedSignsByTitle(query string) ([]PancreatitisSign, error) {
	signs, err := r.GetPublishedSigns()
	if err != nil {
		return nil, err
	}

	var result []PancreatitisSign
	lowerQuery := strings.ToLower(query)
	for _, s := range signs {
		if strings.Contains(strings.ToLower(s.Title), lowerQuery) {
			result = append(result, s)
		}
	}
	return result, nil
}
