package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

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
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "заглушка: аутентификация будет реализована в лабораторной работе №4",
	})
}

func (h *Handler) LogoutAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "заглушка: деавторизация будет реализована в лабораторной работе №4",
	})
}
