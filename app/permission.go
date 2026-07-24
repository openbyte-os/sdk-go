package app

import "github.com/openbyte-os/sdk-go/translation"

type Permission struct {
	Key         string           `json:"key"`
	Name        translation.Text `json:"name"`
	Description translation.Text `json:"description,omitempty"`
	Meta        []PermissionMeta `json:"meta,omitempty"`
}

type PermissionMeta struct {
	Key             string              `json:"key"`
	Name            translation.Text    `json:"name"`
	Description     translation.Text    `json:"description,omitempty"`
	AvailableValues map[string]string   `json:"availableValues,omitempty"`
	Type            PropertyType        `json:"type"`
	DisplayType     PropertyDisplayType `json:"displayType,omitempty"`
	Min             *float64            `json:"min,omitempty"`
	Max             *float64            `json:"max,omitempty"`
}

type PermissionEffect string

const (
	PermissionEffectAllow PermissionEffect = "Allow"
	PermissionEffectDeny  PermissionEffect = "Deny"
)

type PermissionPolicy struct {
	Uuid        string
	Name        translation.Text
	Description translation.Text
	Statements  []PermissionStatement
}

const PermissionResourceAll = "*"

type PermissionStatement struct {
	Effect     PermissionEffect    `json:"e"`
	Permission ScopedKey           `json:"p"`
	Resource   string              `json:"r"` // path or resource indicator defined by the app
	Meta       map[string][]string `json:"m,omitempty"`
}
