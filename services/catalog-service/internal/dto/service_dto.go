package dto

type CreateServiceRequest struct {
	Title           string `json:"title" binding:"required,min=2,max=120"`
	Description     string `json:"description" binding:"required,max=2000"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,gt=0,lte=1440"`
	Price           int    `json:"price" binding:"required,gte=0"`
	IsActive        bool   `json:"is_active"`
}

type UpdateServiceRequest struct {
	Title           *string `json:"title" binding:"omitempty,min=2,max=120"`
	Description     *string `json:"description" binding:"omitempty,max=2000"`
	DurationMinutes *int    `json:"duration_minutes" binding:"omitempty,gt=0,lte=1440"`
	Price           *int    `json:"price" binding:"omitempty,gte=0"`
	IsActive        *bool   `json:"is_active"`
}

type CreateSpecServ struct {
	ServiceID    uint `json:"service_id" binding:"required,gt=0"`
	SpecialistID uint `json:"specialist_id" binding:"required,gt=0"`
}
