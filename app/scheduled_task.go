package app

import (
	"encoding/json"
	"time"
)

// ScheduledTask declares a recurring workspace task in an app definition.
type ScheduledTask struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`  // Absolute app-relative path receiving ScheduledTaskRequest via POST.
	Every string `json:"every"` // Go duration in whole minutes, at least one minute; aligned to UTC epoch boundaries.
}

// ScheduledTaskRequest is sent as JSON when Rubix triggers an enabled workspace task.
// Workspace app settings are supplied separately through transport.RequestConfiguration.
type ScheduledTaskRequest struct {
	TaskID        string          `json:"taskID"`
	Workspace     string          `json:"workspace"`
	ScheduledAt   time.Time       `json:"scheduledAt"`
	Configuration json.RawMessage `json:"configuration"` // Workspace's task-specific JSON object.
}

// ScheduledTaskResponse should be returned as JSON from a scheduled task endpoint.
// Success indicates the task completed or was accepted for asynchronous processing.
// Status is a human-readable description; HTTP 204 may be used for an empty success response.
type ScheduledTaskResponse struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
}
