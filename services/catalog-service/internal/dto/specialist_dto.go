package dto

type SpecialistCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1"`
	Description string `json:"description" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

type SpecialistUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}
