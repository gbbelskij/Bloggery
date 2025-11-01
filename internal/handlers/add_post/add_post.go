package add_post

import (
	"log/slog"
	"net/http"
	sl "new_service/internal/lib/logger"
	"new_service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Request struct {
	Title   string `json:"title" env-required:"true"`
	Content string `json:"content"`
}

type PostSaver interface {
	SavePost(*models.Post) error
}

func New(log *slog.Logger, postSaver PostSaver) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
			return
		}

		userId := c.GetString("user_id")
		parseUserId, err := uuid.Parse(userId)
		if err != nil {
			log.Info("invalid user id", slog.String("userId", userId))
			c.JSON(http.StatusBadRequest, "invalid user id")
			return
		}

		post_to_save := models.Post{
			UserId:  parseUserId,
			Title:   req.Title,
			Content: req.Content,
		}

		err = postSaver.SavePost(&post_to_save)
		if err != nil {
			log.Info("failed to save post", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to save post"})
			return
		}

		log.Info("post saved successfully")
		c.JSON(http.StatusOK, gin.H{"message": "post saved successfully"})
	}
}
