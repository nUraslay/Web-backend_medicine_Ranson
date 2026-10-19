package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ranson-backend/internal/app/auth"
)

const (
	bearerPrefix = "Bearer "
	ctxUserIDKey = "userID"
	ctxTokenKey  = "token"
	ctxClaimsKey = "claims"
)

func userIDFromContext(ctx *gin.Context) uint {
	return ctx.GetUint(ctxUserIDKey)
}

func (h *Handler) authenticate(ctx *gin.Context) (*auth.Claims, string, int, error) {
	header := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return nil, "", http.StatusUnauthorized, errors.New("требуется заголовок Authorization: Bearer <токен>")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))

	claims, err := auth.ParseToken(h.JWT.Secret, tokenStr)
	if err != nil {
		return nil, "", http.StatusUnauthorized, errors.New("токен недействителен или истёк")
	}

	blacklisted, err := h.Redis.IsJWTBlacklisted(ctx.Request.Context(), tokenStr)
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}
	if blacklisted {
		return nil, "", http.StatusUnauthorized, errors.New("токен отозван (выполнен logout)")
	}

	return claims, tokenStr, 0, nil
}

func (h *Handler) setAuth(ctx *gin.Context, claims *auth.Claims, token string) {
	ctx.Set(ctxUserIDKey, claims.UserID)
	ctx.Set(ctxTokenKey, token)
	ctx.Set(ctxClaimsKey, claims)
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, token, status, err := h.authenticate(ctx)
		if err != nil {
			h.errorHandler(ctx, status, err)
			ctx.Abort()
			return
		}

		h.setAuth(ctx, claims, token)
		ctx.Next()
	}
}

func (h *Handler) AuthOptional() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if claims, token, _, err := h.authenticate(ctx); err == nil {
			h.setAuth(ctx, claims, token)
		}

		ctx.Next()
	}
}
