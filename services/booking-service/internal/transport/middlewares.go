package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	AdminRole      string = "admin"
	ClientRole     string = "client"
	SpecialistRole string = "specialist"
)

const (
	GetAllAppointmentsBySpecialistID string = "/appointments/specialist/:id"
	GetAll                           string = "/appointments/all"
	ChangeStatus                     string = "/appointments/:id/status"
	CreateAppointment                string = "/appointments"
	DeleteAppointment                string = "/appointments/:id"
	GetAllMyAppointments             string = "/appointments/my"
)

func AuthorizationMiddleware(ctx *gin.Context) {
	userRole := ctx.GetHeader("X-User-Role")
	if ctx.FullPath() == GetAll && ctx.Request.Method == http.MethodGet {
		if userRole != AdminRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == GetAllMyAppointments && ctx.Request.Method == http.MethodGet {
		if userRole != ClientRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не client"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == GetAllAppointmentsBySpecialistID && ctx.Request.Method == http.MethodGet {
		isAdminOrSpecialist := userRole == AdminRole || userRole == SpecialistRole
		if !isAdminOrSpecialist {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin или specialist"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == ChangeStatus && ctx.Request.Method == http.MethodPatch {
		isAdminOrSpecialist := userRole == AdminRole || userRole == SpecialistRole
		if !isAdminOrSpecialist {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не admin или specialist"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == CreateAppointment && ctx.Request.Method == http.MethodPost {
		if userRole != ClientRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не client"})
			ctx.Abort()
			return
		}
	}

	if ctx.FullPath() == DeleteAppointment && ctx.Request.Method == http.MethodDelete {
		if userRole != ClientRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "вы не client"})
			ctx.Abort()
			return
		}
	}
}
