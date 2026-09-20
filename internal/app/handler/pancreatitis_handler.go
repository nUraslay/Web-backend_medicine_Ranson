package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/repository"
)

const FrontendPath = "../Web-frontend_medicine_Ranson"

const currentUserID uint = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/pancreatitis-signs/feed", h.SignFeedHandler)
	router.GET("/pancreatitis-signs/feed/:id", h.SignFeedHandler)
	router.GET("/pancreatitis-signs/add", h.SignAddHandler)
	router.GET("/pancreatitis-signs/grid", h.SignGridHandler)

	router.POST("/pancreatitis-signs/create", h.SignCreateHandler)
	router.POST("/pancreatitis-signs/publish", h.SignPublishHandler)
	router.POST("/pancreatitis-signs/delete", h.SignDeleteHandler)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob(FrontendPath + "/*.html")
	router.Static("/static", FrontendPath)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
