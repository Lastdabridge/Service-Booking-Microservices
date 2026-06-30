package broker

type UserEvent struct {
	Event     string    `json:"event"`
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt string `json:"created_at"`
}