package dto

type ScheduleCreateRequest struct {
	SpecialistID uint   `json:"specialist_id" binding:"required"`
	Weekday      string `json:"weekday" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    string `json:"start_time" binding:"required,datetime=15:04"`
	EndTime      string `json:"end_time" binding:"required,datetime=15:04"`
}

type ScheduleUpdateRequest struct {
	Weekday   *string `json:"weekday" binding:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime *string `json:"start_time" binding:"omitempty,datetime=15:04"`
	EndTime   *string `json:"end_time" binding:"omitempty,datetime=15:04"`
}
