package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"new_service/internal/config"
	sl "new_service/internal/lib/logger"
	custom_errors "new_service/internal/repository"
	jwt_auth "new_service/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" env-required:"true"`
}

type UserGetter interface {
	GetUserPasswordByEmail(email string) (string, uuid.UUID, error)
	GetUserPasswordByUsername(username string) (string, uuid.UUID, error)
}

func New(log *slog.Logger, cfg *config.Config, userGetter UserGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Info("failed to decode request body", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
			return
		}

		if req.Email == "" && req.Username == "" {
			log.Info("invalid request: no email or username")
			c.JSON(http.StatusBadRequest, gin.H{"message": "email or username must be providen"})
			return
		}

		var err error
		var user_id uuid.UUID
		if req.Email != "" {
			user_id, err = CheckUserPassword(userGetter.GetUserPasswordByEmail, req.Email, req.Password)
		} else {
			user_id, err = CheckUserPassword(userGetter.GetUserPasswordByUsername, req.Username, req.Password)
		}
		if err != nil {
			if errors.Is(err, custom_errors.ErrUserDoesNotExist) {
				log.Info("No user with such email and username")
				c.JSON(http.StatusNotFound, gin.H{"message": "No user with such email and username"})
				return
			}
			if errors.Is(err, custom_errors.ErrInvalidPassword) {
				log.Info("Invalid password")
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid password"})
				return
			}

			log.Info("failed to check password", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to check password"})
			return
		}

		jwt_token, err := jwt_auth.MakeJwtToken(cfg.JWTSecret, user_id)
		if err != nil {
			log.Info("failed to create jwt", sl.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to create jwt"})
			return
		}

		c.SetCookie(
			"jwt_token",
			jwt_token,
			60*60*24,
			"/",
			"",
			false,
			true,
		)
	}
}

func CheckUserPassword(GetUserPassword func(string) (string, uuid.UUID, error), email_or_username string, request_password string) (uuid.UUID, error) {
	user_password, user_id, err := GetUserPassword(email_or_username)
	if err != nil {
		return user_id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user_password), []byte(request_password))
	if err != nil {
		return user_id, custom_errors.ErrInvalidPassword
	}

	return user_id, nil
}
