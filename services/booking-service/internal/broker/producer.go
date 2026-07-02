package broker

import (
	"booking-service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type BookingEventsProducer struct {
	writer *kafka.Writer
}

type Booking struct {
	Event        string    `json:"event"`
	BookingID    uint      `json:"booking_id"`
	ClientID     uint      `json:"client_id"`
	Weekday      string    `json:"weekday"`
	SpecialistID uint      `json:"specialist_id"`
	ServiceID    uint      `json:"service_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p BookingEventsProducer) Close() error {
	if p.writer == nil {
		return nil
	}

	return p.writer.Close()
}

func (p BookingEventsProducer) PublishBookingEvent(booking models.Appointment, event string) error {
	events := Booking{
		Event:        event,
		BookingID:    booking.ID,
		ClientID:     booking.ClientID,
		Weekday:      booking.Weekday,
		SpecialistID: booking.SpecialistID,
		ServiceID:    booking.ServiceID,
		StartTime:    *booking.StartTime,
		EndTime:      *booking.EndTime,
		CreatedAt:    booking.CreatedAt,
	}

	eventJson, err := json.Marshal(events)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("booking-%d", booking.ID)),
		Value: eventJson,
	}

	err = p.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return err
	}

	log.Printf("Событие отправлено в Kafka: booking_id=%d", booking.ID)
	return nil
}
