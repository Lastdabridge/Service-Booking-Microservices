package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/gateway/internal/transport"
	"github.com/gin-gonic/gin"
)

func reverseProxy(target string) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()

	userProxy := reverseProxy(os.Getenv("USER_URL"))
	catalogProxy := reverseProxy(os.Getenv("CATALOG_URL"))
	bookingProxy := reverseProxy(os.Getenv("BOOKING_URL"))
	notificationsAndAuditProxy := reverseProxy(os.Getenv("NOTIFICATIONS_AUDIT_URL"))

	unprotected := r.Group("/api")

	protected := r.Group("/api")
	protected.Use(
		transport.AuthMiddleware,
		transport.InjectHeaders(),
	)

	unprotected.Any("/auth/register", userProxy)
	unprotected.Any("/auth/login", userProxy)
	protected.Any("/auth/me", userProxy)

	protected.Any("/services/*path", catalogProxy)

	protected.Any("/specialists/*path", catalogProxy)

	protected.Any("/appointments/*path", bookingProxy)

	protected.Any("/notifications/*path", notificationsAndAuditProxy)
	protected.Any("/audit/*path", notificationsAndAuditProxy)

	r.Run(":8080")
}
