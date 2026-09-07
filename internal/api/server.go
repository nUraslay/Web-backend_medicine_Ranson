package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/handler"
	"ranson-backend/internal/app/repository"
)

const FrontendPath = "../Web-frontend_medicine_Ranson"

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob(FrontendPath + "/*.html")

	r.Static("/static", FrontendPath)

	r.GET("/feed", h.FeedHandler)
	r.GET("/feed/:id", h.FeedHandler) // GET ленты по ID

	r.GET("/add", h.AddHandler) // получение черновика

	r.GET("/grid", h.GridHandler) // список всех услуг

	r.Run()
	log.Println("Server down")
}
