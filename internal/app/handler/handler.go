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

type ServiceFeedView struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
	VideoURL    string
	LikesCount  int
}

type ServiceCardView struct {
	ID         int
	Title      string
	Category   string
	ImageURL   string
	LikesCount int
}

func toFeedView(s repository.Service) ServiceFeedView {
	return ServiceFeedView{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		ImageURL:    s.ImageURL(),
		VideoURL:    s.VideoURL(),
		LikesCount:  len(s.LikedByUserIDs),
	}
}

func toCardView(s repository.Service) ServiceCardView {
	return ServiceCardView{
		ID:         s.ID,
		Title:      s.Title,
		Category:   s.Category,
		ImageURL:   s.ImageURL(),
		LikesCount: len(s.LikedByUserIDs),
	}
}

// FeedHandler контроллер
func (h *Handler) FeedHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	nextParam := ctx.Query("next")

	var service repository.Service
	var err error

	if idParam == "" {
		var services []repository.Service
		services, err = h.Repository.GetPublishedServices()
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusNotFound, "feed.html", gin.H{"Error": "Услуги не найдены"})
			return
		}
		service = services[0]
	} else {
		var id int
		id, err = strconv.Atoi(idParam)
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusBadRequest, "feed.html", gin.H{"Error": "Некорректный ID услуги"})
			return
		}

		if nextParam == "true" {
			service, err = h.Repository.GetNextPublishedService(id)
		} else {
			service, err = h.Repository.GetPublishedServiceByID(id)
		}
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusNotFound, "feed.html", gin.H{"Error": "Услуга не найдена"})
			return
		}
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"Service": toFeedView(service),
	})
}

// AddHandler контроллер
func (h *Handler) AddHandler(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftService()
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "add.html", gin.H{"Error": "Черновик не найден"})
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"Draft": draft,
	})
}

// GridHandler котроллер
func (h *Handler) GridHandler(ctx *gin.Context) {
	filterParam := ctx.Query("filter")

	var services []repository.Service
	var err error

	if filterParam == "" {
		services, err = h.Repository.GetPublishedServices()
	} else {
		services, err = h.Repository.GetPublishedServicesByTitle(filterParam)
	}

	if err != nil {
		logrus.Error(err)
	}

	cards := make([]ServiceCardView, 0, len(services))
	for _, s := range services {
		cards = append(cards, toCardView(s))
	}

	ctx.HTML(http.StatusOK, "grid.html", gin.H{
		"Services": cards,
		"Filter":   filterParam,
	})
}
