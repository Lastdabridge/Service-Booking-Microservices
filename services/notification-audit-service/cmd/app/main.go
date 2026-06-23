package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Veoler/notification-audit-service/internal/config"
	"github.com/Veoler/notification-audit-service/internal/kafka"
	"github.com/Veoler/notification-audit-service/internal/model"
	"github.com/Veoler/notification-audit-service/internal/repository"
	"github.com/Veoler/notification-audit-service/internal/service"
	"github.com/Veoler/notification-audit-service/internal/transport"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	log.Println("[STARTUP] конфиг загружен")

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("[STARTUP] ошибка подключения к БД: %v", err)
	}
	log.Println("[STARTUP] база данных подключена")

	if err := db.AutoMigrate(&model.Notification{}, &model.AuditLog{}); err != nil {
		log.Fatalf("[STARTUP] ошибка миграции: %v", err)
	}
	log.Println("[STARTUP] миграции выполнены")

	notifRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	kafka.InitWriter(cfg.KafkaBroker)
	defer kafka.CloseWriter()

	publisher := kafka.NewProducer()

	notifSvc := service.NewNotificationService(notifRepo, publisher)
	auditSvc := service.NewAuditService(auditRepo, publisher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafka.StartConsumers(ctx, cfg.KafkaBroker, cfg.KafkaGroupID, notifSvc, auditSvc)

	notifHandler := transport.NewNotificationHandler(notifSvc)
	auditHandler := transport.NewAuditHandler(auditSvc)
	router := transport.SetupRouter(notifHandler, auditHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("[STARTUP] HTTP сервер запущен на :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[STARTUP] ошибка HTTP сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SHUTDOWN] получен сигнал завершения...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[SHUTDOWN] ошибка при остановке HTTP сервера: %v", err)
	}

	log.Println("[SHUTDOWN] сервис остановлен")
}