package deletePost

import (
	"log/slog"
	"net/http"
	sl "new_service/internal/lib/logger"
	"new_service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Request struct {
	PostId uuid.UUID `json:"post_id" env-required:"true"`
}

type PostDeleter interface {
	DeletePost(post_id uuid.UUID) error
	GetPost(post_id uuid.UUID) (models.DbPost, error)
}

func New(log *slog.Logger, postDeleter PostDeleter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Info("invalid request", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
			return
		}

		postToDelete, err := postDeleter.GetPost(req.PostId)
		if err != nil {
			log.Info("failed to get post", sl.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get post"})
			return
		}

		providedUserId, ok := c.Get("user_id")
		if !ok {
			log.Info("no user id")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "no user id"})
			return
		}
		parsedProvidedUserId, err := uuid.Parse(providedUserId.(string))
		if err != nil {
			log.Info("invalid user id", sl.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid user id"})
			return
		}
		if postToDelete.UserId != parsedProvidedUserId {
			log.Info("not user's post, forbidden")
			c.JSON(http.StatusForbidden, gin.H{"message": "not your post"})
			return
		}
		log.Info("post author checked successfully")

		err = postDeleter.DeletePost(req.PostId)
		if err != nil {
			log.Info("failed to delete post", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to delete post"})
			return
		}

		log.Info("deleted post successfully")
		c.JSON(http.StatusOK, gin.H{"message": "deleted post successfully"})
	}
}
