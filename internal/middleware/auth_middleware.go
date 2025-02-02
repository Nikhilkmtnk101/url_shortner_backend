package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/utils"
	"net/http"
	"strings"
)

func AuthMiddleware(sessionRepo *repository.SessionRepository, jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			utils.NewResponse().
				SetStatus(http.StatusUnauthorized).
				SetMessage("Authorization header is required").
				SetErrorCode("UNAUTHORIZED").
				SetData(nil).
				Build(ctx)
			ctx.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			utils.NewResponse().
				SetStatus(http.StatusUnauthorized).
				SetMessage("Invalid token").
				SetErrorCode("UNAUTHORIZED").
				SetData(nil).
				Build(ctx)
			ctx.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["user_id"].(float64))
		session, err := sessionRepo.GetUserSession(ctx, userID)

		if err != nil {
			utils.NewResponse().
				SetStatus(http.StatusUnauthorized).
				SetMessage("Logged out user").
				SetErrorCode("UNAUTHORIZED").
				SetData(nil).
				Build(ctx)
			ctx.Abort()
			return
		}
		if session == nil || session.AccessToken != tokenString {
			utils.NewResponse().
				SetStatus(http.StatusUnauthorized).
				SetMessage("Invalid token").
				SetErrorCode("UNAUTHORIZED").
				SetData(nil).
				Build(ctx)
			ctx.Abort()
			return
		}
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}
