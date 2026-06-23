package kafkadto

type NotificationCreatedEvent struct {
	Event          string `json:"event"`           
	NotificationID uint   `json:"notification_id"` 
	UserID         uint   `json:"user_id"`         
	Type           string `json:"type"`            
	SourceEvent    string `json:"source_event"`    
	CreatedAt      string `json:"created_at"`
}

type NotificationReadEvent struct {
	Event          string `json:"event"`           
	NotificationID uint   `json:"notification_id"` 
	UserID         uint   `json:"user_id"`         
	CreatedAt      string `json:"created_at"`
}

type NotificationFailedEvent struct {
	Event       string `json:"event"`        
	UserID      uint   `json:"user_id"`      
	SourceEvent string `json:"source_event"` 
	Reason      string `json:"reason"`       
	CreatedAt   string `json:"created_at"`
}

type AuditLoggedEvent struct {
	Event         string `json:"event"`          
	AuditID       uint   `json:"audit_id"`       
	SourceEvent   string `json:"source_event"`   
	SourceService string `json:"source_service"` 
	CreatedAt     string `json:"created_at"`
}

// публикуем в audit.events, когда обнаружена подозрительная активность (access.denied из gateway.events)
type SuspiciousActivityEvent struct {
	Event     string `json:"event"`      
	ActorID   uint  `json:"actor_id"`   
	Path      string `json:"path"`       
	Method    string `json:"method"`     
	Reason    string `json:"reason"`     
	CreatedAt string `json:"created_at"`
}

// универсальная структура для входящих событий
type KafkaEvent struct {
	Event string `json:"event"` 

	// Поля из users.events)
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	// Поля из booking.events
	BookingID    uint   `json:"booking_id"`
	ClientID     uint   `json:"client_id"`    
	SpecialistID uint   `json:"specialist_id"`
	ServiceID    uint   `json:"service_id"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`

	// Поля из catalog.events
	Title string `json:"title"`

	// Поля из gateway.events
	Path   string `json:"path"`
	Method string `json:"method"`

	CreatedAt string `json:"created_at"`
}