package dto

import "time"

type ScheduleCreateRequest struct {
	Weekday      string    `json:"weekday" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    time.Time `json:"start_time" binding:"required,datetime=15:04"`
	EndTime      time.Time `json:"end_time" binding:"required,datetime=15:04"`
}

type ScheduleUpdateRequest struct {
	Weekday   *string    `json:"weekday" binding:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime *time.Time `json:"start_time" binding:"omitempty,datetime=15:04"`
	EndTime   *time.Time `json:"end_time" binding:"omitempty,datetime=15:04"`
}
