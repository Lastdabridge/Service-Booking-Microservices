package validation

import (
	"catalog-service/internal/dto"
	"fmt"
	"strings"
	"time"
)

type Validator interface {
	ValidateServiceCreate(req dto.CreateServiceRequest) error
	ValidateServiceUpdate(req dto.UpdateServiceRequest) error
	ValidateSpecialistCreate(req dto.SpecialistCreateRequest) error
	ValidateSpecialistUpdate(req dto.SpecialistUpdateRequest) error
	ValidateScheduleCreate(specialistID uint, req dto.ScheduleCreateRequest) error
	ValidateScheduleUpdate(req dto.ScheduleUpdateRequest, currentStartTime, currentEndTime time.Time) error
	ValidateCreateSpecServ(req dto.CreateSpecServ) error
}

type ServiceValidator struct{}

func NewValidator() Validator {
	return &ServiceValidator{}
}

func (v *ServiceValidator) ValidateServiceCreate(req dto.CreateServiceRequest) error {
	if err := validateRequiredText(req.Title, "title", 2, 120); err != nil {
		return err
	}
	if err := validateOptionalText(req.Description, "description", 0, 2000); err != nil {
		return err
	}
	if req.DurationMinutes <= 0 || req.DurationMinutes > 1440 {
		return fmt.Errorf("duration_minutes must be between 1 and 1440")
	}
	if req.Price < 0 {
		return fmt.Errorf("price must not be negative")
	}
	return nil
}

func (v *ServiceValidator) ValidateServiceUpdate(req dto.UpdateServiceRequest) error {
	if req.Title != nil {
		if err := validateRequiredText(*req.Title, "title", 2, 120); err != nil {
			return err
		}
	}
	if req.Description != nil {
		if err := validateOptionalText(*req.Description, "description", 0, 2000); err != nil {
			return err
		}
	}
	if req.DurationMinutes != nil {
		if *req.DurationMinutes <= 0 || *req.DurationMinutes > 1440 {
			return fmt.Errorf("duration_minutes must be between 1 and 1440")
		}
	}
	if req.Price != nil && *req.Price < 0 {
		return fmt.Errorf("price must not be negative")
	}
	return nil
}

func (v *ServiceValidator) ValidateSpecialistCreate(req dto.SpecialistCreateRequest) error {
	if err := validateRequiredText(req.Name, "name", 2, 100); err != nil {
		return err
	}
	if err := validateRequiredText(req.Description, "description", 1, 1000); err != nil {
		return err
	}
	return nil
}

func (v *ServiceValidator) ValidateSpecialistUpdate(req dto.SpecialistUpdateRequest) error {
	if req.Name != nil {
		if err := validateRequiredText(*req.Name, "name", 2, 100); err != nil {
			return err
		}
	}
	if req.Description != nil {
		if err := validateRequiredText(*req.Description, "description", 1, 1000); err != nil {
			return err
		}
	}
	return nil
}

func (v *ServiceValidator) ValidateScheduleCreate(specialistID uint, req dto.ScheduleCreateRequest) error {
	if specialistID == 0 {
		return fmt.Errorf("specialist_id must be greater than 0")
	}
	if err := validateWeekday(req.Weekday); err != nil {
		return err
	}
	return nil
}

func (v *ServiceValidator) ValidateScheduleUpdate(req dto.ScheduleUpdateRequest, currentStartTime, currentEndTime time.Time) error {
	if req.Weekday != nil {
		if err := validateWeekday(*req.Weekday); err != nil {
			return err
		}
	}

	return nil
}

func (v *ServiceValidator) ValidateCreateSpecServ(req dto.CreateSpecServ) error {
	if req.ServiceID == 0 {
		return fmt.Errorf("service_id must be greater than 0")
	}
	if req.SpecialistID == 0 {
		return fmt.Errorf("specialist_id must be greater than 0")
	}
	return nil
}

func validateRequiredText(value, field string, minLen, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(trimmed) < minLen {
		return fmt.Errorf("%s must be at least %d characters", field, minLen)
	}
	if len(trimmed) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func validateOptionalText(value, field string, minLen, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if len(trimmed) < minLen {
		return fmt.Errorf("%s must be at least %d characters", field, minLen)
	}
	if len(trimmed) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func validateWeekday(value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return nil
	default:
		return fmt.Errorf("weekday must be one of monday, tuesday, wednesday, thursday, friday, saturday, sunday")
	}
}
