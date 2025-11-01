package registration

import (
	"errors"
	"log/slog"
	"net/http"
	sl "new_service/internal/lib/logger"
	"new_service/internal/lib/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type Request struct {
	Email    string `json:"email" env-required:"true"`
	Username string `json:"username"`
	Password string `json:"password" env-required:"true"`
}

type UserSaver interface {
	SaveUser(email string, password string, username string) error
}

func New(userSaver UserSaver, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Info("Invalid request", sl.Error(err))
			c.JSON(http.StatusBadRequest, response.Error(err))
			return
		}
		log.Info("Request body decoded successfully")

		err := userSaver.SaveUser(req.Email, req.Password, req.Username)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				log.Info("Failed to save user: email already exists", sl.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "Error",
					"message": "email already exists",
				})
				return
			}
			log.Info("Failed to save user", sl.Error(err))
			c.JSON(http.StatusBadRequest, response.Error(err))
			return
		}

		log.Info("user saved sucessfully")
	}
}
