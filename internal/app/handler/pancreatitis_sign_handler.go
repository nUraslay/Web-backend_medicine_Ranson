package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/ds"
	"ranson-backend/internal/app/repository"
)

const (
	defaultImageURL = "/static/img/default_pancreatitis_sign.jpg"
	defaultVideoURL = "/static/video/default_pancreatitis_sign.mp4"
)

type SignFeedView struct {
	ID          uint
	Title       string
	Description string
	VideoURL    string
	LikesCount  int
}

type SignCardView struct {
	ID         uint
	Title      string
	Stage      string
	ImageURL   string
	LikesCount int
}

func urlOrDefault(url string, defaultURL string) string {
	if strings.TrimSpace(url) == "" {
		return defaultURL
	}
	return url
}

func toFeedView(s ds.PancreatitisSign, likesCount int) SignFeedView {
	return SignFeedView{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		VideoURL:    urlOrDefault(s.VideoURL, defaultVideoURL),
		LikesCount:  likesCount,
	}
}

func toCardView(s ds.PancreatitisSign, likesCount int) SignCardView {
	return SignCardView{
		ID:         s.ID,
		Title:      s.Title,
		Stage:      s.Stage,
		ImageURL:   urlOrDefault(s.ImageURL, defaultImageURL),
		LikesCount: likesCount,
	}
}

func (h *Handler) SignFeedHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	nextParam := ctx.Query("next")

	var sign *ds.PancreatitisSign
	var err error

	if idParam == "" {
		sign, err = h.Repository.GetFirstPublishedSign()
	} else {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			logrus.Error(convErr)
			ctx.HTML(http.StatusBadRequest, "feed_pancreatitis.html", gin.H{"Error": "Некорректный ID признака"})
			return
		}

		if nextParam == "true" {
			sign, err = h.Repository.GetNextPublishedSign(id)
		} else {
			sign, err = h.Repository.GetPublishedSignByID(id)
		}
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if sign == nil {
		ctx.HTML(http.StatusNotFound, "feed_pancreatitis.html", gin.H{"Error": "Признак не найден"})
		return
	}

	likesCount, err := h.Repository.GetLikesCount(sign.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "feed_pancreatitis.html", gin.H{
		"Sign": toFeedView(*sign, likesCount),
	})
}

func (h *Handler) SignAddHandler(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftSign(currentUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "add_pancreatitis.html", gin.H{
		"Draft": draft,
	})
}

func (h *Handler) SignGridHandler(ctx *gin.Context) {
	filterParam := ctx.Query("filter")

	var signs []ds.PancreatitisSign
	var err error

	if filterParam == "" {
		signs, err = h.Repository.GetPublishedSigns()
	} else {
		signs, err = h.Repository.SearchPublishedSignsByTitle(filterParam)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likesCounts, err := h.Repository.GetLikesCounts()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	cards := make([]SignCardView, 0, len(signs))
	for _, s := range signs {
		cards = append(cards, toCardView(s, likesCounts[s.ID]))
	}

	ctx.HTML(http.StatusOK, "grid_pancreatitis.html", gin.H{
		"Signs":  cards,
		"Filter": filterParam,
	})
}

func (h *Handler) SignCreateHandler(ctx *gin.Context) {
	title := strings.TrimSpace(ctx.PostForm("title"))
	if title == "" || utf8.RuneCountInString(title) > 100 {
		ctx.HTML(http.StatusBadRequest, "add_pancreatitis.html", gin.H{
			"Error": "Название обязательно и не должно быть длиннее 100 символов",
		})
		return
	}

	draft, err := h.Repository.GetDraftSign(currentUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if draft == nil {
		_, err = h.Repository.CreateDraftSign(title, currentUserID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.Redirect(http.StatusFound, "/pancreatitis-signs/add")
}

func (h *Handler) SignPublishHandler(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftSign(currentUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft == nil {
		ctx.Redirect(http.StatusFound, "/pancreatitis-signs/add")
		return
	}

	title := strings.TrimSpace(ctx.PostForm("title"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	stage := strings.TrimSpace(ctx.PostForm("stage"))
	thresholdValue, convErr := strconv.ParseFloat(strings.TrimSpace(ctx.PostForm("threshold_value")), 64)

	validationError := ""
	switch {
	case title == "" || utf8.RuneCountInString(title) > 100:
		validationError = "Название обязательно и не должно быть длиннее 100 символов"
	case description == "" || utf8.RuneCountInString(description) > 500:
		validationError = "Описание обязательно и не должно быть длиннее 500 символов"
	case stage == "" || utf8.RuneCountInString(stage) > 30:
		validationError = "Этап оценки обязателен и не должен быть длиннее 30 символов"
	case convErr != nil || !(thresholdValue >= 0 && thresholdValue < 1e8):
		validationError = "Пороговое значение должно быть числом от 0 до 99999999"
	}
	if validationError != "" {
		ctx.HTML(http.StatusBadRequest, "add_pancreatitis.html", gin.H{
			"Error": validationError,
			"Draft": draft,
		})
		return
	}

	err = h.Repository.PublishSign(draft.ID, title, description, stage, thresholdValue)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/pancreatitis-signs/grid")
}

func (h *Handler) SignDeleteHandler(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.PostForm("sign_id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный ID признака"))
		return
	}

	err = h.Repository.DeleteSign(uint(id))
	if errors.Is(err, repository.ErrSignNotFound) {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/pancreatitis-signs/grid")
}
