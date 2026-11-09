package app

import (
	"strings"

	"github.com/openbyte-os/sdk-go/translation"
)

// ActionLaunch is how the console runs an action when its button is pressed.
type ActionLaunch string

const (
	ActionLaunchBackground ActionLaunch = "background" // (default) Submit in place; the button shows CompletedText on success
	ActionLaunchModal      ActionLaunch = "modal"      // Submit into a modal, for the app to render a confirmation form
	ActionLaunchSlide      ActionLaunch = "slide"      // Submit into a slide-over panel
)

// ActionStatus is an action's state for one payload, reported by its StatusPath.
type ActionStatus string

const (
	ActionStatusAvailable   ActionStatus = "available"   // The action can run; the button is idle
	ActionStatusCompleted   ActionStatus = "completed"   // Already done; the button shows CompletedText, or hides with HideCompleted
	ActionStatusProcessing  ActionStatus = "processing"  // Running elsewhere; the button shows its processing state
	ActionStatusUnavailable ActionStatus = "unavailable" // Cannot run for this payload; the button is disabled
)

// ParseActionStatus returns the status a header value names, ignoring case and surrounding space.
func ParseActionStatus(value string) (ActionStatus, bool) {
	status := ActionStatus(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case ActionStatusAvailable, ActionStatusCompleted, ActionStatusProcessing, ActionStatusUnavailable:
		return status, true
	}
	return "", false
}

// Action is a pre-defined submission other apps can place on their pages with
// <app-action-button gaid="vendor/app" action="key">, letting a user send data
// to this app without leaving the page, e.g. "Add to my to-do".
//
// The console POSTs the button's data, merged with any inputs inside the
// button, to Path with the user's context. A background handler answers with
// a 2xx to mark the button complete, optionally toasting through
// transport.ResponseAlertSuccess; transport.ResponseAlertDanger or a 4xx/5xx
// marks it failed. With ActionLaunchModal or ActionLaunchSlide the response
// body renders in that layer, e.g. a confirmation form.
//
// Rubix rejects a button submission before it reaches Path unless it carries
// a valid CSRF token and its fields satisfy Inputs (required, type, available
// values, regex). Requests that passed carry transport.RequestAction; the
// handler can still reject the action.
//
// The button renders only for users who can access this app and hold every
// RequiredPermissions entry.
type Action struct {
	Key         string           `json:"key"`
	Text        translation.Text `json:"text"`                  // Button label, e.g. "Add to my to-do"
	Icon        string           `json:"icon,omitempty"`        // Material icon shown with Text, e.g. "add"
	Description translation.Text `json:"description,omitempty"` // What the action does, for discovery and tooltips

	ProcessingText translation.Text `json:"processingText,omitempty"` // Label while a background submission is in flight, e.g. "Adding…"
	ProcessingIcon string           `json:"processingIcon,omitempty"` // Icon while a background submission is in flight; a spinner when empty

	CompletedText translation.Text `json:"completedText,omitempty"` // Label after a background submission succeeds, e.g. "On your list"
	CompletedIcon string           `json:"completedIcon,omitempty"` // Icon after a background submission succeeds, e.g. "check"

	Path   string       `json:"path"`             // App path that receives the POST
	Launch ActionLaunch `json:"launch,omitempty"` // How the submission is presented

	// Once marks an action that completes once per payload, e.g. "Add to my
	// to-do": a successful background submission is remembered in the browser
	// and the button renders completed on later page loads.
	Once bool `json:"once,omitempty"`
	// StatusPath is an app path the console sends a HEAD request to when the
	// button renders, with the button's fields as query parameters. It answers
	// with an ActionStatus in the transport.ResponseActionStatus header, which
	// replaces any completion the browser remembered.
	StatusPath string `json:"statusPath,omitempty"`
	// HideCompleted hides the button once completed instead of showing CompletedText.
	HideCompleted bool `json:"hideCompleted,omitempty"`

	Inputs              []Property  `json:"inputs,omitempty"`              // Data fields the action accepts
	RequiredPermissions []ScopedKey `json:"requiredPermissions,omitempty"` // Permissions the user needs to see the button
}

// LaunchMode returns the action's launch, defaulting to ActionLaunchBackground.
func (a Action) LaunchMode() ActionLaunch {
	switch a.Launch {
	case ActionLaunchModal, ActionLaunchSlide:
		return a.Launch
	default:
		return ActionLaunchBackground
	}
}

// Action returns the definition's action with the given key, or nil.
func (d *Definition) Action(key string) *Action {
	if d == nil || key == "" {
		return nil
	}
	for i := range d.Actions {
		if d.Actions[i].Key == key {
			return &d.Actions[i]
		}
	}
	return nil
}
