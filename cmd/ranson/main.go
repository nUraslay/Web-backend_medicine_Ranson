package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"ranson-backend/internal/app/config"
	"ranson-backend/internal/app/dsn"
	"ranson-backend/internal/app/handler"
	"ranson-backend/internal/app/repository"
	"ranson-backend/internal/pkg"
)

func main() {
	logrus.Info("Application start!")

	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("не найден .env или в нем нет DB_HOST: запускайте команду из папки Web-backend_medicine_Ranson")
	}

	rep, err := repository.New(&repository.RepositorySettings{
		PostgresDSN:     postgresString,
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		MinioPublicURL:  os.Getenv("MINIO_PUBLIC_URL"),
	})
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()

	logrus.Info("Application terminated!")
}
