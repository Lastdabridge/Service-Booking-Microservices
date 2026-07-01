package transport

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(notifHandler *NotificationHandler, auditHandler *AuditHandler) *gin.Engine {
	router := gin.Default()

	auth := router.Group("/", RequireAuth())
	{
		notifications := auth.Group("/notifications/")
		{
			notifications.GET("my", notifHandler.GetMyNotifs)
			notifications.PATCH(":id/read", notifHandler.MarkNotifAsRead)
		}

		audit := auth.Group("/audit/", RequireRole("admin"))
		{
			audit.GET("events", auditHandler.GetAllAudits)
			audit.GET("events/:id", auditHandler.GetAuditByID)
		}
	}

	return router
}
