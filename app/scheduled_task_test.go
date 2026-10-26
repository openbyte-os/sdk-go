package app

import (
	"encoding/json"
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
	if len(parsed.ScheduledTasks) != 1 || parsed.ScheduledTasks[0] != def.ScheduledTasks[0] {
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
	data, err = json.Marshal(ScheduledTaskResponse{Status: "Queued", Success: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"status":"Queued","success":true}` {
		t.Fatalf("response JSON: %s", data)
	}
}
