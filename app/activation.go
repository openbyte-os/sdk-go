package app

import "github.com/openbyte-os/sdk-go/translation"

type ActivationScope string

const (
	ActivationScopeWorkspace ActivationScope = "workspace"
	ActivationScopeUser      ActivationScope = "user"
)

type ActivationActionType string

const (
	ActivationActionTypeComplete ActivationActionType = "complete"
	ActivationActionTypeContinue ActivationActionType = "continue"
	ActivationActionTypeVerify   ActivationActionType = "verify"
	ActivationActionTypeSetup    ActivationActionType = "setup"
	ActivationActionTypeCreate   ActivationActionType = "create"
)

type ActivationStep struct {
	ID               string
	Name             translation.Text
	Description      translation.Text
	Icon             string
	DestinationPath  string
	InstructionsPath string // path to load instructions
	Scope            ActivationScope
	Priority         int
	ActionType       ActivationActionType
	Required         bool
}
