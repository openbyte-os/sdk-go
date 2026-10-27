package app

import (
	"encoding/json"
	"time"
)

// ScheduledTask declares a recurring workspace task in an app definition.
type ScheduledTask struct {
	AllowManualTrigger bool       `json:"allowManualTrigger,omitempty"` // Workspace managers may run this task on demand, independently of its schedule.
	Required           bool       `json:"required,omitempty"`           // Runs whenever the app is installed; workspace managers cannot disable it.
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description,omitempty"`
	Configuration      []Property `json:"configuration,omitempty"` // Workspace task options, using the standard property labels, types, choices and defaults.
	Path               string     `json:"path"`                    // Task route under PathPrefixTasks, with an optional leading slash or existing prefix.
	Every              string     `json:"every"`                   // Whole-minute Go duration (e.g. 24h) or five-field cron expression (e.g. "0 1 * * *"); evaluated in UTC.
}

// ScheduledTaskTrigger identifies what initiated a task invocation.
type ScheduledTaskTrigger string

const (
	ScheduledTaskTriggerScheduled ScheduledTaskTrigger = "scheduled"
	ScheduledTaskTriggerManual    ScheduledTaskTrigger = "manual"
)

// ScheduledTaskRequest is sent as JSON when Rubix triggers an enabled workspace task.
// Workspace app settings are supplied separately through transport.RequestConfiguration.
type ScheduledTaskRequest struct {
	Trigger       ScheduledTaskTrigger `json:"trigger,omitempty"` // Scheduled or manually requested by a workspace manager.
	RunID         string               `json:"runID,omitempty"`   // Stable occurrence ID, also supplied as the Idempotency-Key header.
	TaskID        string               `json:"taskID"`
	Workspace     string               `json:"workspace"`
	ScheduledAt   time.Time            `json:"scheduledAt"`
	Configuration json.RawMessage      `json:"configuration"` // Workspace's task-specific JSON object.
}

// ScheduledTaskStatus describes the outcome reported by the receiving app.
type ScheduledTaskStatus string

const (
	ScheduledTaskStatusFailed  ScheduledTaskStatus = "failed"
	ScheduledTaskStatusSuccess ScheduledTaskStatus = "success"
	ScheduledTaskStatusQueued  ScheduledTaskStatus = "queued"
)

// ScheduledTaskResponse should be returned as JSON from a scheduled task endpoint.
// Queued acknowledges asynchronous work; StatusURL can open its current state in the app.
type ScheduledTaskResponse struct {
	Status     ScheduledTaskStatus `json:"status"`
	Message    string              `json:"message"`
	ReceivedID string              `json:"receivedID"`          // Vendor-generated receipt/job identifier.
	StatusURL  string              `json:"statusURL,omitempty"` // App-relative path, e.g. /jobs/job-123. May include a query string.
	// Deprecated: set Status to success, failed or queued. Retained for older task handlers.
	Success bool `json:"success,omitempty"`
}
