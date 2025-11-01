package getNextPosts

import (
	"log/slog"
	"net/http"
	"new_service/internal/handlers/structs"
	sl "new_service/internal/lib/logger"
	"new_service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostsGetter interface {
	GetNextPosts(userId uuid.UUID, paginarionParams structs.PaginationParams) ([]models.DbPost, error)
}

func New(log *slog.Logger, postsGetter PostsGetter) gin.HandlerFunc {
	return func(c *gin.Context) {

		userId := c.GetString("user_id")

		parsedUserId, err := uuid.Parse(c.GetString("user_id"))
		if err != nil {
			log.Info("invalid user id", slog.String("userId", userId))
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
			return
		}

		paginationParamas := structs.PaginationParams{}
		if err := c.ShouldBindQuery(&paginationParamas); err != nil {
			log.Info("invalid query params")
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid query params"})
			return
		}

		posts, err := postsGetter.GetNextPosts(parsedUserId, paginationParamas)
		if err != nil {
			log.Info("failed to get posts", sl.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to get posts"})
			return
		}

		log.Info("Next posts got successdully")
		c.JSON(http.StatusOK, posts)
	}
}
