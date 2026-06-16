package broker

import "time"

type SpecialistDeleted struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}
