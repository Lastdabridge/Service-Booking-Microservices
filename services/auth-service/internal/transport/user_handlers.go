package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/apperrors"
	model "github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/model"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	user := r.Group("/auth")
	{
		user.POST("/register", h.Register)
		user.POST("/login", h.Login)
		user.GET("/me", h.MyAccount)
	}

}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.UserRegister

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.RegisterUser(req)
	if err != nil {
		errEx := apperrors.Get(err)
		c.JSON(errEx.StatusCode, gin.H{"error": errEx.Msg})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "you have been successfully registered"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.UserLogin

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.LoginUser(req)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) MyAccount(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing user context",
		})
		return
	}

	// надо ли проверять, если передает не клиент?
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	//нужно ли переводить в uint, если принимаем то, что пришло из gateway?
	user, err := h.service.GetMyAccount(uint(userID))
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrAccountNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
