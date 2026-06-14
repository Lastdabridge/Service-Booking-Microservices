package transport

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(notifHandler *NotificationHandler, auditHandler *AuditHandler) *gin.Engine {
	r := gin.Default()

	auth := r.Group("/", RequireAuth())
	{
		notifications := auth.Group("/notifications")
		{
			notifications.GET("/my", notifHandler.GetMy)
			notifications.PATCH("/:id/read", notifHandler.MarkAsRead)
		}

		audit := auth.Group("/audit", RequireRole("admin"))
		{
			audit.GET("/events", auditHandler.GetAll)
			audit.GET("/events/:id", auditHandler.GetByID)
		}
	}

	return r
}
