package app

type UIMode string

const (
	UIModeFull        UIMode = "full"        // Standard application
	UIModeDextral     UIMode = "dextral"     // Dextral application
	UIModeBasic       UIMode = "basic"       // simple application - used for non-native applications
	UIModeIntegration UIMode = "integration" // Integrated into existing pages
	UIModeHelp        UIMode = "help"        // Help App
	UIModeNone        UIMode = "none"        // No UI - flow only
)
