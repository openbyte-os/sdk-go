package gateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/openbyte-os/sdk-go/app"
)

var notificationsApp = app.NewID("kx", "notifications")

// Notification is published to the workspace inbox by Notify. Category is a
// notification type key the calling app declares in its definition.
type Notification struct {
	// ID identifies the notification within the calling app; publishing the same
	// ID with a higher Revision updates it. Empty publishes a new one-off notification.
	ID       string `json:"id"`
	Revision int64  `json:"revision"`
	Category string `json:"category"`
	Group    string `json:"group,omitempty"`
	Title    string `json:"title"`
	Body     string `json:"body,omitempty"`
	// State defaults to posted; tasks use queued, running, succeeded, failed or cancelled.
	State    string `json:"state"`
	Progress *int   `json:"progress,omitempty"`
	// Path is a destination inside the calling app, starting with "/".
	Path string `json:"path,omitempty"`
	// Priority is debug, info, warning, error or critical.
	Priority string               `json:"priority,omitempty"`
	Actions  []NotificationAction `json:"actions,omitempty"`
	// TTL is seconds until the notification expires.
	TTL       int64      `json:"ttl,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// Users are the recipients' user IDs; empty means everyone authorized for the calling app.
	Users []string `json:"users,omitempty"`
}

// NotificationAction opens Path in the calling app, or runs one of its declared
// Rubix actions with Inputs as form values.
type NotificationAction struct {
	Label  string            `json:"label,omitempty"`
	Icon   string            `json:"icon,omitempty"`
	Path   string            `json:"path,omitempty"`
	Action string            `json:"action,omitempty"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

// Notify publishes n as the calling app through the platform notifications app.
func (w *WorkspaceClient) Notify(ctx context.Context, n Notification) error {
	if n.ID == "" {
		id := make([]byte, 16)
		_, _ = rand.Read(id)
		n.ID = hex.EncodeToString(id)
	}
	if n.Revision == 0 {
		n.Revision = 1
	}
	if n.State == "" {
		n.State = "posted"
	}
	_, err := w.CallOperation(ctx, notificationsApp, "publish", n, nil)
	return err
}
