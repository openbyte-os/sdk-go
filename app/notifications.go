package app

import "github.com/openbyte-os/sdk-go/translation"

// Notifications declares the notifications an app publishes. Once an app
// declares any groups, publishing a type it hasn't declared is rejected.
type Notifications struct {
	Groups []NotificationGroup `json:"groups,omitempty"`
}

// NotificationGroup is a set of related notification types that users can
// mute or configure together.
type NotificationGroup struct {
	Key         string             `json:"key"`
	Name        translation.Text   `json:"name"`
	Description translation.Text   `json:"description,omitzero"`
	Types       []NotificationType `json:"types"`
}

// NotificationType is one kind of notification, published as its category.
// Keys are unique across the whole app: a publish names only the type, and
// its group comes from the definition.
type NotificationType struct {
	Key         string           `json:"key"`
	Name        translation.Text `json:"name"`
	Description translation.Text `json:"description,omitzero"`
}

// Type finds a declared type and the group it belongs to; both are nil when it isn't declared.
func (n *Notifications) Type(key string) (*NotificationGroup, *NotificationType) {
	if n == nil {
		return nil, nil
	}
	for g := range n.Groups {
		for t := range n.Groups[g].Types {
			if n.Groups[g].Types[t].Key == key {
				return &n.Groups[g], &n.Groups[g].Types[t]
			}
		}
	}
	return nil, nil
}
