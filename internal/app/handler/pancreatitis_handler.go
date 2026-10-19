package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/config"
	"ranson-backend/internal/app/redis"
	"ranson-backend/internal/app/repository"
)

const FrontendPath = "../Web-frontend_medicine_Ranson"

const currentUserID uint = 1

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWT        config.JWTConfig
}

func NewHandler(r *repository.Repository, rd *redis.Client, jwtCfg config.JWTConfig) *Handler {
	resolveMedia = r.MediaURL

	return &Handler{
		Repository: r,
		Redis:      rd,
		JWT:        jwtCfg,
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

	h.RegisterAPI(router)

	// Swagger UI: http://localhost:8080/swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json")))
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
