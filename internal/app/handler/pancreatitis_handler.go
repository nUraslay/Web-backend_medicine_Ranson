package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type SignFeedView struct {
	ID          int
	Title       string
	Description string
	VideoURL    string
	LikesCount  int
}

type SignCardView struct {
	ID         int
	Title      string
	Category   string
	ImageURL   string
	LikesCount int
}

func toFeedView(s repository.PancreatitisSign) SignFeedView {
	return SignFeedView{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		VideoURL:    s.VideoURL(),
		LikesCount:  len(s.LikedByUserIDs),
	}
}

func toCardView(s repository.PancreatitisSign) SignCardView {
	return SignCardView{
		ID:         s.ID,
		Title:      s.Title,
		Category:   s.Category,
		ImageURL:   s.ImageURL(),
		LikesCount: len(s.LikedByUserIDs),
	}
}

// SignFeedHandler контроллер
func (h *Handler) SignFeedHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	nextParam := ctx.Query("next")

	var sign repository.PancreatitisSign
	var err error

	if idParam == "" {
		var signs []repository.PancreatitisSign
		signs, err = h.Repository.GetPublishedSigns()
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusNotFound, "feed_pancreatitis.html", gin.H{"Error": "Признаки панкреатита не найдены"})
			return
		}
		sign = signs[0]
	} else {
		var id int
		id, err = strconv.Atoi(idParam)
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusBadRequest, "feed_pancreatitis.html", gin.H{"Error": "Некорректный ID признака"})
			return
		}

		if nextParam == "true" {
			sign, err = h.Repository.GetNextPublishedSign(id)
		} else {
			sign, err = h.Repository.GetPublishedSignByID(id)
		}
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusNotFound, "feed_pancreatitis.html", gin.H{"Error": "Признак не найден"})
			return
		}
	}

	ctx.HTML(http.StatusOK, "feed_pancreatitis.html", gin.H{
		"Sign": toFeedView(sign),
	})
}

// SignAddHandler контроллер
func (h *Handler) SignAddHandler(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftSign()
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "add_pancreatitis.html", gin.H{"Error": "Черновик не найден"})
		return
	}

	ctx.HTML(http.StatusOK, "add_pancreatitis.html", gin.H{
		"Draft": draft,
	})
}

// SignGridHandler котроллер
func (h *Handler) SignGridHandler(ctx *gin.Context) {
	filterParam := ctx.Query("filter")

	var signs []repository.PancreatitisSign
	var err error

	if filterParam == "" {
		signs, err = h.Repository.GetPublishedSigns()
	} else {
		signs, err = h.Repository.GetPublishedSignsByTitle(filterParam)
	}

	if err != nil {
		logrus.Error(err)
	}

	cards := make([]SignCardView, 0, len(signs))
	for _, s := range signs {
		cards = append(cards, toCardView(s))
	}

	ctx.HTML(http.StatusOK, "grid_pancreatitis.html", gin.H{
		"Signs":  cards,
		"Filter": filterParam,
	})
}
