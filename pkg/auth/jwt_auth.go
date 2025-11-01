package jwt_auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func MakeJwtToken(secretKey string, user_id uuid.UUID) (string, error) {
	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"user_id": user_id,
		"jti":     jti,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetClaim(token *jwt.Token, key string) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to get claims")
	}
	value, ok := claims[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("failed to get value")
	}
	return value, nil
}

func GetExpFromRawToken(token string, jwt_secret string) (int64, error) {
	parsed_token, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwt_secret), nil
	})

	claims, ok := parsed_token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("failed to get claims")
	}
	exp, ok := claims["exp"].(float64)
	if !ok || exp == 0 {
		return 0, fmt.Errorf("failed to get exp")
	}

	return int64(exp), nil
}

func JWTAuthMiddleware(jwt_secret string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("jwt_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unathorized user"})
			return
		}

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(jwt_secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			return
		}

		user_id, err := GetClaim(token, "user_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "user id is missing"})
			return
		}

		jti, err := GetClaim(token, "jti")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "jti is missing"})
			return
		}

		exists, err := rdb.Exists(c.Request.Context(), jti).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
		if exists == 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "you have logged out"})
			return
		}

		c.Set("user_id", user_id)
		c.Set("jti", jti)
		c.Next()
	}
}
