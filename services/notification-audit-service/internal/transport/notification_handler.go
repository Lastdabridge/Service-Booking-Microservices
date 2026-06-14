package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Veoler/notification-audit-service/internal/service"
)

type NotificationHandler struct {
	notification service.NotificationService
}

func NewNotificationHandler(notification service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notification: notification}
}

func (h *NotificationHandler) GetMy(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	notifications, err := h.notification.GetMyNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.notification.MarkNotificationAsRead(uint(id), userID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
