package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Veoler/notifications-audit-service/internal/config"
	"github.com/Veoler/notifications-audit-service/internal/kafka"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"github.com/Veoler/notifications-audit-service/internal/repository"
	"github.com/Veoler/notifications-audit-service/internal/service"
	"github.com/Veoler/notifications-audit-service/internal/transport"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	log.Println("[STARTUP] configuration loaded successfully")

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("[STARTUP] failed to connect to database: %v", err)
	}
	log.Println("[STARTUP] database connected successfully")

	if err := db.AutoMigrate(&model.Notification{}, &model.AuditLog{}); err != nil {
		log.Fatalf("[STARTUP] database migration failed: %v", err)
	}
	log.Println("[STARTUP] database migrations applied successfully")

	notifRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	kafka.InitWriter(cfg)
	defer kafka.CloseWriter()

	producer := kafka.NewProducer(cfg)

	notifSvc := service.NewNotificationService(notifRepo, producer)
	auditSvc := service.NewAuditService(auditRepo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafka.StartConsumers(ctx, cfg, notifSvc, auditSvc)

	notifHandler := transport.NewNotificationHandler(notifSvc)
	auditHandler := transport.NewAuditHandler(auditSvc)
	router := transport.SetupRouter(notifHandler, auditHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("[STARTUP] HTTP server is running on port :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[STARTUP] HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SHUTDOWN] shutdown signal received, initiating graceful shutdown...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[SHUTDOWN] HTTP server shutdown error: %v", err)
	}

	log.Println("[SHUTDOWN] service stopped completely")
}