package app

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestScheduledTasksAffectDefinitionHash(t *testing.T) {
	def := Definition{}
	initial := def.GetHash(false)
	def.ScheduledTasks = []ScheduledTask{{ID: "sync", Name: "Sync", Path: "/tasks/sync", Every: "15m"}}
	withTask := def.GetHash(false)
	if initial == withTask {
		t.Fatal("task declaration did not change definition hash")
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"scheduledTasks":[`) {
		t.Fatal("missing scheduledTasks JSON field")
	}
	parsed, err := FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.ScheduledTasks) != 1 || !reflect.DeepEqual(parsed.ScheduledTasks[0], def.ScheduledTasks[0]) {
		t.Fatal("task declaration did not round trip")
	}
	def.ScheduledTasks[0].Every = "30m"
	if withTask == def.GetHash(false) {
		t.Fatal("interval did not change definition hash")
	}
}

func TestScheduledTaskJSONContracts(t *testing.T) {
	req := ScheduledTaskRequest{TaskID: "sync", Workspace: "workspace-1", ScheduledAt: time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC), Configuration: json.RawMessage(`{"full":true}`)}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"taskID":"sync","workspace":"workspace-1","scheduledAt":"2026-10-26T12:00:00Z","configuration":{"full":true}}` {
		t.Fatalf("request JSON: %s", data)
	}
	var decoded ScheduledTaskRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.TaskID != req.TaskID || decoded.Workspace != req.Workspace || !decoded.ScheduledAt.Equal(req.ScheduledAt) || string(decoded.Configuration) != string(req.Configuration) {
		t.Fatal("request did not round trip")
	}
	data, err = json.Marshal(ScheduledTaskResponse{Status: ScheduledTaskStatusQueued, Message: "Optimisation queued", ReceivedID: "job-123", StatusURL: "/jobs/job-123"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"status":"queued","message":"Optimisation queued","receivedID":"job-123","statusURL":"/jobs/job-123"}` {
		t.Fatalf("response JSON: %s", data)
	}
}

func TestScheduledTaskRequiredDefaultsAndHash(t *testing.T) {
	var optional ScheduledTask
	if err := json.Unmarshal([]byte(`{"id":"sync","path":"/sync","every":"1h"}`), &optional); err != nil {
		t.Fatal(err)
	}
	if optional.Required {
		t.Fatal("task without required must be optional")
	}
	def := Definition{ScheduledTasks: []ScheduledTask{optional}}
	hash := def.GetHash(false)
	def.ScheduledTasks[0].Required = true
	if def.GetHash(false) == hash {
		t.Fatal("required flag did not change definition hash")
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.ScheduledTasks) != 1 || !decoded.ScheduledTasks[0].Required {
		t.Fatal("required flag did not round trip")
	}
}

func TestScheduledTaskConfigurationRoundTripAndHash(t *testing.T) {
	def := Definition{ScheduledTasks: []ScheduledTask{{ID: "sync", Path: "/sync", Every: "1h"}}}
	initial := def.GetHash(false)
	def.ScheduledTasks[0].Configuration = []Property{{Key: "includeArchived", Type: PropertyTypeBoolean, Definition: &PropertyDefinition{DefaultValue: PropertyValue{Bool: true}}}}
	if def.GetHash(false) == initial {
		t.Fatal("option schema did not affect definition hash")
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.ScheduledTasks[0].Configuration) != 1 || !decoded.ScheduledTasks[0].Configuration[0].Definition.DefaultValue.Bool {
		t.Fatal("option schema did not round trip")
	}
}

func TestScheduledTaskManualTriggerIsAdditionalAndOptIn(t *testing.T) {
	task := ScheduledTask{ID: "sync", Path: "/sync", Every: "0 1 * * *"}
	if task.AllowManualTrigger {
		t.Fatal("manual triggering must default to disabled")
	}
	def := Definition{ScheduledTasks: []ScheduledTask{task}}
	originalHash := def.GetHash(false)
	def.ScheduledTasks[0].AllowManualTrigger = true
	if def.GetHash(false) == originalHash {
		t.Fatal("manual capability did not affect definition hash")
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.ScheduledTasks[0].AllowManualTrigger || decoded.ScheduledTasks[0].Every != task.Every {
		t.Fatal("manual option changed or replaced the schedule")
	}
	payload := ScheduledTaskRequest{TaskID: "sync", Trigger: ScheduledTaskTriggerManual}
	data, err = json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var request ScheduledTaskRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	if request.Trigger != ScheduledTaskTriggerManual {
		t.Fatal("manual invocation source was lost")
	}
}
