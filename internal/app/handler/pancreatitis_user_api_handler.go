package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"ranson-backend/internal/app/auth"
	"ranson-backend/internal/app/ds"
)

func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	if err := parseForm(ctx); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	login := strings.TrimSpace(ctx.Request.FormValue("login"))
	password := ctx.Request.FormValue("password")

	if login == "" || utf8.RuneCountInString(login) > 25 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("login: обязательное поле, не длиннее 25 символов"))
		return
	}
	if len(password) < 6 || len(password) > 72 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("password: от 6 до 72 символов"))
		return
	}

	existing, err := h.Repository.GetUserByLogin(login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if existing != nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("пользователь с таким логином уже существует"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	user := ds.PancreatitisUser{
		Login:    login,
		Password: string(hash),
	}

	if err = h.Repository.AddUser(&user); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    user, 
		"message": "пользователь успешно зарегистрирован",
	})
}

func (h *Handler) LoginAPI(ctx *gin.Context) {
	if err := parseForm(ctx); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	login := strings.TrimSpace(ctx.Request.FormValue("login"))
	password := ctx.Request.FormValue("password")

	if login == "" || password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("login и password: обязательные поля"))
		return
	}

	user, err := h.Repository.GetUserByLogin(login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("неверный логин или пароль"))
		return
	}

	token, err := auth.GenerateToken(h.JWT.Secret, h.JWT.ExpiresIn, user.ID, user.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": LoginData{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(h.JWT.ExpiresIn.Seconds()),
		},
		"message": "аутентификация успешна",
	})
}

func (h *Handler) LogoutAPI(ctx *gin.Context) {
	token := ctx.GetString(ctxTokenKey)

	value, _ := ctx.Get(ctxClaimsKey)
	claims, ok := value.(*auth.Claims)
	if !ok || token == "" || claims.ExpiresAt == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("требуется авторизация"))
		return
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl > 0 {
		if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), token, ttl); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "выход выполнен, токен отозван",
	})
}
