package dto

type CreateServiceRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"decription" binding:"required"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,gt=0"`
	Price           int    `json:"price" binding:"required,gt=0"`
	IsActive        bool   `json:"is_active"`
}

type UpdateServiceRequest struct {
	Title           *string `json:"title"`
	Description     *string `json:"decription"`
	DurationMinutes *int    `json:"duration_minutes"`
	Price           *int    `json:"price"`
	IsActive        *bool   `json:"is_active"`
}
