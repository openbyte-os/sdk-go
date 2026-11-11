package app

import "testing"

func TestNotificationsType(t *testing.T) {
	n := &Notifications{Groups: []NotificationGroup{
		{Key: "billing", Types: []NotificationType{{Key: "invoice-paid"}, {Key: "payment-failed"}}},
		{Key: "backups", Types: []NotificationType{{Key: "backup-done"}}},
	}}
	tests := []struct {
		name      string
		n         *Notifications
		key       string
		wantGroup string
	}{
		{name: "first group", n: n, key: "payment-failed", wantGroup: "billing"},
		{name: "second group", n: n, key: "backup-done", wantGroup: "backups"},
		{name: "undeclared", n: n, key: "other"},
		{name: "group key is not a type", n: n, key: "billing"},
		{name: "nil", n: nil, key: "invoice-paid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, typ := tt.n.Type(tt.key)
			if tt.wantGroup == "" {
				if g != nil || typ != nil {
					t.Fatalf("Type(%q) = %v, %v; want nil", tt.key, g, typ)
				}
				return
			}
			if g == nil || g.Key != tt.wantGroup || typ == nil || typ.Key != tt.key {
				t.Fatalf("Type(%q) = %v, %v; want group %q", tt.key, g, typ, tt.wantGroup)
			}
		})
	}
}
