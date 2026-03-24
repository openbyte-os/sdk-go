package app

import "github.com/openbyte-os/sdk-go/translation"

type SettingsPage struct {
	ID           string
	Name         translation.Text
	Panels       []SettingsPanel
	AdvancedPath string
}

type SettingsPanel struct {
	Name        translation.Text
	Description translation.Text
	Settings    []Property
	Order       int
}
