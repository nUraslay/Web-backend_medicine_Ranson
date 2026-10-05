package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"ranson-backend/internal/app/ds"
	"ranson-backend/internal/app/repository"
)

var (
	allowedImageTypes = map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true}
	allowedVideoTypes = map[string]bool{"video/mp4": true, "video/webm": true}
)


type SignSerializer struct {
	ID             uint       `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Stage          string     `json:"stage"`
	ThresholdValue float64    `json:"threshold_value"`
	ImageURL       string     `json:"image_url"`
	VideoURL       string     `json:"video_url"`
	CreatedAt      time.Time  `json:"created_at"`
	FormedAt       *time.Time `json:"formed_at"`
	LikesCount     int        `json:"likes_count"`
	IsLiked        int        `json:"is_liked"` 
	IsMine         int        `json:"is_mine"`  
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func toSignSerializer(s ds.PancreatitisSign, likesCount int, liked bool) SignSerializer {
	return SignSerializer{
		ID:             s.ID,
		Title:          s.Title,
		Description:    s.Description,
		Status:         s.Status,
		Stage:          s.Stage,
		ThresholdValue: s.ThresholdValue,
		ImageURL:       urlOrDefault(s.ImageURL, defaultImageURL),
		VideoURL:       urlOrDefault(s.VideoURL, defaultVideoURL),
		CreatedAt:      s.CreatedAt,
		FormedAt:       s.FormedAt,
		LikesCount:     likesCount,
		IsLiked:        boolToInt(liked),
		IsMine:         boolToInt(s.CreatorID == CurrentUserID()),
	}
}

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")
	api.GET("/pancreatitis-signs", h.GetSignsAPI)
	api.GET("/pancreatitis-signs/feed", h.GetFeedAPI)
	api.GET("/pancreatitis-signs/draft", h.GetDraftAPI)
	api.POST("/pancreatitis-signs", h.AddSignAPI)
	api.PUT("/pancreatitis-signs/:id/publish", h.PublishSignAPI)
	api.DELETE("/pancreatitis-signs/:id", h.DeleteSignAPI)
	api.POST("/pancreatitis-signs/:id/like", h.LikeSignAPI)
	api.POST("/users", h.RegisterUserAPI)
	api.POST("/auth/login", h.LoginAPI)
	api.POST("/auth/logout", h.LogoutAPI)
}

func parseForm(ctx *gin.Context) error {
	err := ctx.Request.ParseMultipartForm(2 << 20)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}

	return nil
}

func (h *Handler) parseIDParam(ctx *gin.Context) (int, bool) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный id: нужно целое число больше 0"))
		return 0, false
	}

	return id, true
}

func optionalFile(ctx *gin.Context, field string) (*multipart.FileHeader, error) {
	header, err := ctx.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return header, nil
}

func validateFileUpload(header *multipart.FileHeader, allowed map[string]bool, maxSize int64) (string, int, error) {
	if header.Size > maxSize {
		return "", http.StatusBadRequest, fmt.Errorf("файл %s слишком большой (максимум %d МБ)", header.Filename, maxSize>>20)
	}

	file, err := header.Open()
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("не удалось получить файл")
	}
	defer file.Close()

	// Тип файла определяется по его первым 512 байтам
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", http.StatusInternalServerError, fmt.Errorf("не удалось прочитать файл")
	}
	if n == 0 {
		return "", http.StatusBadRequest, fmt.Errorf("файл пустой")
	}

	contentType := http.DetectContentType(buffer[:n])
	if !allowed[contentType] {
		return "", http.StatusBadRequest, fmt.Errorf("недопустимый тип файла: %s", contentType)
	}

	return contentType, 0, nil
}

type signInput struct {
	Title          *string
	Description    *string
	Stage          *string
	ThresholdValue *float64
}

func readSignInput(ctx *gin.Context) (signInput, error) {
	var input signInput
	form := ctx.Request.PostForm

	if form.Has("title") {
		title := strings.TrimSpace(ctx.Request.FormValue("title"))
		if title == "" || utf8.RuneCountInString(title) > 100 {
			return input, errors.New("title: обязательное поле, не длиннее 100 символов")
		}
		input.Title = &title
	}

	if form.Has("description") {
		description := strings.TrimSpace(ctx.Request.FormValue("description"))
		if utf8.RuneCountInString(description) > 500 {
			return input, errors.New("description: не длиннее 500 символов")
		}
		input.Description = &description
	}

	if form.Has("stage") {
		stage := strings.TrimSpace(ctx.Request.FormValue("stage"))
		if utf8.RuneCountInString(stage) > 30 {
			return input, errors.New("stage: не длиннее 30 символов")
		}
		input.Stage = &stage
	}

	if form.Has("threshold_value") {
		value, err := strconv.ParseFloat(strings.TrimSpace(ctx.Request.FormValue("threshold_value")), 64)
		if err != nil || !(value >= 0 && value < 1e8) {
			return input, errors.New("threshold_value: число от 0 до 99999999")
		}
		input.ThresholdValue = &value
	}

	return input, nil
}

func applySignInput(sign *ds.PancreatitisSign, input signInput) {
	if input.Title != nil {
		sign.Title = *input.Title
	}
	if input.Description != nil {
		sign.Description = *input.Description
	}
	if input.Stage != nil {
		sign.Stage = *input.Stage
	}
	if input.ThresholdValue != nil {
		sign.ThresholdValue = *input.ThresholdValue
	}
}

func (h *Handler) findOwnSign(ctx *gin.Context, id int, userID uint) (*ds.PancreatitisSign, bool) {
	sign, err := h.Repository.GetSignByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return nil, false
	}
	if sign == nil || sign.Status == ds.StatusDeleted {
		h.errorHandler(ctx, http.StatusNotFound, repository.ErrSignNotFound)
		return nil, false
	}
	if sign.CreatorID != userID {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("признак создан другим пользователем"))
		return nil, false
	}

	return sign, true
}

func (h *Handler) publishedSignsResponse(ctx *gin.Context, title string, maxThreshold *float64) {
	signs, err := h.Repository.GetPublishedSigns(title, maxThreshold)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likesCounts, err := h.Repository.GetLikesCounts()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	liked, err := h.Repository.GetLikedSignIDs(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	items := make([]SignSerializer, 0, len(signs))
	for _, s := range signs {
		items = append(items, toSignSerializer(s, likesCounts[s.ID], liked[s.ID]))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   items,
	})
}

func (h *Handler) GetSignsAPI(ctx *gin.Context) {
	title := strings.TrimSpace(ctx.Query("filter"))

	var maxThreshold *float64
	if raw := strings.TrimSpace(ctx.Query("max_value")); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil || value < 0 {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("max_value: неотрицательное число"))
			return
		}
		maxThreshold = &value
	}

	h.publishedSignsResponse(ctx, title, maxThreshold)
}

func (h *Handler) GetFeedAPI(ctx *gin.Context) {
	h.publishedSignsResponse(ctx, "", nil)
}

func (h *Handler) GetDraftAPI(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftSign(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   toSignSerializer(*draft, 0, false),
	})
}

func (h *Handler) AddSignAPI(ctx *gin.Context) {
	userID := CurrentUserID()

	if err := parseForm(ctx); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	input, err := readSignInput(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if input.Title == nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("title: обязательное поле"))
		return
	}

	draft, err := h.Repository.GetDraftSign(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft != nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("у пользователя уже есть черновик: опубликуйте или удалите его"))
		return
	}

	imageHeader, err := optionalFile(ctx, "image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoHeader, err := optionalFile(ctx, "video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var imageType, videoType string
	if imageHeader != nil {
		var code int
		imageType, code, err = validateFileUpload(imageHeader, allowedImageTypes, maxImageSize)
		if err != nil {
			h.errorHandler(ctx, code, err)
			return
		}
	}
	if videoHeader != nil {
		var code int
		videoType, code, err = validateFileUpload(videoHeader, allowedVideoTypes, maxVideoSize)
		if err != nil {
			h.errorHandler(ctx, code, err)
			return
		}
	}

	sign := ds.PancreatitisSign{
		Status:    ds.StatusDraft,
		CreatorID: userID,
	}
	applySignInput(&sign, input)

	if imageHeader != nil {
		sign.ImageURL, err = h.Repository.UploadMedia(imageHeader, "image", imageType)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	if videoHeader != nil {
		sign.VideoURL, err = h.Repository.UploadMedia(videoHeader, "video", videoType)
		if err != nil {
			h.Repository.RemoveMedia(sign.ImageURL)
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	if err = h.Repository.AddSign(&sign); err != nil {
		h.Repository.RemoveMedia(sign.ImageURL)
		h.Repository.RemoveMedia(sign.VideoURL)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    toSignSerializer(sign, 0, false),
		"message": "черновик успешно создан",
	})
}

func (h *Handler) PublishSignAPI(ctx *gin.Context) {
	userID := CurrentUserID()

	id, ok := h.parseIDParam(ctx)
	if !ok {
		return
	}

	if err := parseForm(ctx); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	sign, ok := h.findOwnSign(ctx, id, userID)
	if !ok {
		return
	}
	if sign.Status != ds.StatusDraft {
		h.errorHandler(ctx, http.StatusConflict, errors.New("опубликовать можно только черновик"))
		return
	}

	input, err := readSignInput(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	applySignInput(sign, input)

	if sign.Title == "" || sign.Description == "" || sign.Stage == "" {
		h.errorHandler(ctx, http.StatusBadRequest,
			errors.New("для публикации нужны title, description и stage (их можно передать в этом запросе)"))
		return
	}

	err = h.Repository.PublishDraftSign(sign.ID, userID, sign)
	if errors.Is(err, repository.ErrSignNotFound) {
		h.errorHandler(ctx, http.StatusConflict, errors.New("опубликовать можно только черновик"))
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	published, err := h.Repository.GetSignByID(id)
	if err != nil || published == nil {
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("не удалось прочитать опубликованный признак"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    toSignSerializer(*published, 0, false),
		"message": "признак опубликован",
	})
}

func (h *Handler) DeleteSignAPI(ctx *gin.Context) {
	userID := CurrentUserID()

	id, ok := h.parseIDParam(ctx)
	if !ok {
		return
	}

	sign, ok := h.findOwnSign(ctx, id, userID)
	if !ok {
		return
	}

	err := h.Repository.SoftDeleteSign(sign.ID, userID)
	if errors.Is(err, repository.ErrSignNotFound) {
		h.errorHandler(ctx, http.StatusConflict, errors.New("признак уже удален"))
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "признак успешно удален",
	})
}

func (h *Handler) LikeSignAPI(ctx *gin.Context) {
	userID := CurrentUserID()

	id, ok := h.parseIDParam(ctx)
	if !ok {
		return
	}

	if err := parseForm(ctx); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	like := strings.TrimSpace(ctx.Request.FormValue("like"))
	if like != "0" && like != "1" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("like: допустимые значения 0 или 1"))
		return
	}

	sign, err := h.Repository.GetPublishedSignByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if sign == nil {
		h.errorHandler(ctx, http.StatusNotFound, repository.ErrSignNotFound)
		return
	}

	if err = h.Repository.SetLike(userID, sign.ID, like == "1"); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likesCount, err := h.Repository.GetLikesCount(sign.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	message := "лайк снят"
	if like == "1" {
		message = "лайк поставлен"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
		"data": gin.H{
			"sign_id":     sign.ID,
			"is_liked":    boolToInt(like == "1"),
			"likes_count": likesCount,
		},
	})
}
