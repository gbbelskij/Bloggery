package logout

import (
	"log/slog"
	"net/http"
	sl "new_service/internal/lib/logger"
	jwt_auth "new_service/pkg/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func New(log *slog.Logger, rdb *redis.Client, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {

		token, err := c.Cookie("jwt_token")
		if err != nil {
			log.Info("failed to get jwt_token")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to get jwt token"})
			return
		}

		jti, ok := c.Get("jti")
		if !ok {
			log.Info("no jti")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "no jti"})
			return
		}

		string_jti, ok := jti.(string)
		if !ok {
			log.Info("invalid jti")
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid jti"})
			return
		}

		expUnix, err := jwt_auth.GetExpFromRawToken(token, jwtSecret)
		if err != nil {
			log.Info("failed to get exp", sl.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to get exp"})
			return
		}

		nowUnix := time.Now().Unix()
		ttlSeconds := nowUnix - expUnix
		if ttlSeconds < 0 {
			ttlSeconds = 0
		}

		rdb.Set(c.Request.Context(), string_jti, "revoked", time.Duration(ttlSeconds))
		log.Info("logged out successfully")
		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	}
}
