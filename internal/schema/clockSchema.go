package schema

import "time"

type PunchRequest struct {
	Pin string `json:"pin" binding:"required,len=4,numeric"`
}

type PunchResponse struct {
	Action string    `json:"action"`
	Name   string    `json:"name"`
	At     time.Time `json:"at"`
}

type UpdateClockEntryRequest struct {
	ClockInAt  time.Time  `json:"clock_in_at" binding:"required"`
	ClockOutAt *time.Time `json:"clock_out_at"`
}
