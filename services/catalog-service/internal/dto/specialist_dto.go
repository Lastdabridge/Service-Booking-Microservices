package dto

type SpecialistCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required,max=1000"`
	IsActive    bool   `json:"is_active"`
}

type SpecialistUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}
