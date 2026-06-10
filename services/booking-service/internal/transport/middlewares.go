package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthorizationMiddleware(ctx *gin.Context) {
	userRole := ctx.GetHeader("X-User-Role")
	if ctx.Request.URL.String() == "/appointments/all" && ctx.Request.Method == "GET" {
		if userRole != "admin" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == "/appointments/specialist/:id" && ctx.Request.Method == "GET" {
		isAdminOrSpecialist := userRole == "admin" || userRole == "specialist"
		if !isAdminOrSpecialist {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin или specialist"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == "/appointments/:id/status" && ctx.Request.Method == "PATCH" {
		isAdminOrSpecialist := userRole == "admin" || userRole == "specialist"
		if !isAdminOrSpecialist {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin или specialist"})
			ctx.Abort()
			return
		}
	}

	if ctx.Request.URL.String() == "/appointments" && ctx.Request.Method == "POST" {
		if userRole != "client" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не client"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == "/appointments/:id" && ctx.Request.Method == "DELETE" {
		if userRole != "client" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не client"})
			ctx.Abort()
			return
		}
	}
}
