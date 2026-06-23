package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort					string
	DatabaseURL					string
	KafkaBroker					string
	KafkaGroupID				string

	TopicUsersEvents			string
	TopicCatalogEvents			string
	TopicBookingEvents			string
	TopicNotificationsEvents	string
	TopicGatewayEvents			string

	TopicAuditEvents			string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading environment variables directly")
	}

	config := &Config{
		HTTPPort:					getEnv("HTTP_PORT", "8084"),
		DatabaseURL:				getEnv("DATABASE_URL", "host=localhost user=postgres password=4545 dbname=notification_db port=5432 sslmode=disable"),
		KafkaBroker:				getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "notification-audit-service"),

		TopicUsersEvents:			getEnv("TOPIC_USERS_EVENTS", "users.events"),
		TopicCatalogEvents:			getEnv("TOPIC_CATALOG_EVENTS", "catalog.events"),
		TopicBookingEvents:			getEnv("TOPIC_BOOKING_EVENTS", "booking.events"),
		TopicNotificationsEvents: 	getEnv("TOPIC_NOTIFICATIONS_EVENTS", "notifications.events"),
		TopicGatewayEvents:			getEnv("TOPIC_GATEWAY_EVENTS", "gateway.events"),

		TopicAuditEvents:			getEnv("TOPIC_AUDIT_EVENTS", "audit.events"),
	}

	config.DatabaseURL = requireEnv("DATABASE_URL")
 
	return config
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("обязательная переменная окружения %s не задана", key)
	}
	return v
}
