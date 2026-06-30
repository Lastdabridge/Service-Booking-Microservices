package transport

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/gateway/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing authorization header",
		})
		ctx.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		ctx.Abort()
		return
	}

	ctx.Set("userID", claims.UserID)
	ctx.Set("userRole", claims.Role)

	ctx.Next()
}

func InjectHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID, _ := ctx.Get("userID")
		role, _ := ctx.Get("userRole")

		ctx.Request.Header.Set(
			"X-User-ID",
			fmt.Sprintf("%v", userID),
		)

		ctx.Request.Header.Set(
			"X-User-Role",
			fmt.Sprintf("%v", role),
		)

		ctx.Next()
	}
}
