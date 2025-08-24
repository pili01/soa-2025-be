package models

import "time"

type ExecutionStatus string

const (
	ExecutionStatusPending    ExecutionStatus = "pending"
	ExecutionStatusInProgress ExecutionStatus = "in_progress"
	ExecutionStatusCompleted  ExecutionStatus = "completed"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusAborted    ExecutionStatus = "aborted"
)

type TourExecution struct {
	ID                int                `json:"id" bson:"_id"`
	TourID            int                `json:"tour_id" bson:"tour_id"`
	UserID            int                `json:"user_id" bson:"user_id"`
	StartedAt         *time.Time         `json:"started_at,omitempty" bson:"started_at,omitempty"`
	EndedAt           *time.Time         `json:"ended_at,omitempty" bson:"ended_at,omitempty"`
	LastActivity      *time.Time         `json:"last_activity,omitempty" bson:"last_activity,omitempty"`
	Status            ExecutionStatus    `json:"status" bson:"status"`
	FinishedKeypoints []FinishedKeyPoint `json:"finished_keypoints,omitempty" bson:"finished_keypoints,omitempty"`
}

type FinishedKeyPoint struct {
	KeypointID  int        `json:"keypoint_id" bson:"keypoint_id"`
	CompletedAt *time.Time `json:"completed_at" bson:"completed_at"`
}
