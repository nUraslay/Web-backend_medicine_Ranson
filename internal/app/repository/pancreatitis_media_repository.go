package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

var mediaExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
	"video/mp4":  ".mp4",
	"video/webm": ".webm",
}

func (r *Repository) UploadMedia(header *multipart.FileHeader, prefix string, contentType string) (string, error) {
	extension, ok := mediaExtensions[contentType]
	if !ok {
		return "", fmt.Errorf("неподдерживаемый тип файла: %s", contentType)
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	name := fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), extension)

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucketName,
		name,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})
	if err != nil {
		return "", fmt.Errorf("не удалось добавить объект в хранилище minio: %w", err)
	}

	return name, nil
}

func (r *Repository) RemoveMedia(name string) {
	if name == "" {
		return
	}
	_ = r.minio.RemoveObject(context.Background(), r.minioBucketName, name, minio.RemoveObjectOptions{})
}

func (r *Repository) MediaURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "http") {
		return value
	}
	return r.minioPublicURL + "/" + r.minioBucketName + "/" + value
}
