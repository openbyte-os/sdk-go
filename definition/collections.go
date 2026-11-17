package definition

import "github.com/openbyte-os/sdk-go/app"

// WithPermissions appends to the definition's Permissions collection.
func (b *Builder) WithPermissions(values ...app.Permission) *Builder {
	b.Definition.Permissions = append(b.Definition.Permissions, values...)
	return b
}

// WithOperations appends to the definition's Operations collection.
func (b *Builder) WithOperations(values ...app.Operation) *Builder {
	b.Definition.Operations = append(b.Definition.Operations, values...)
	return b
}

// WithDependencies appends to the definition's Dependencies collection.
func (b *Builder) WithDependencies(values ...app.Dependency) *Builder {
	b.Definition.Dependencies = append(b.Definition.Dependencies, values...)
	return b
}

// WithProviders appends to the definition's Providers collection.
func (b *Builder) WithProviders(values ...app.ProviderRegistration) *Builder {
	b.Definition.Providers = append(b.Definition.Providers, values...)
	return b
}

// WithProviderDependencies appends to the definition's ProviderDependencies collection.
func (b *Builder) WithProviderDependencies(values ...app.ProviderDependency) *Builder {
	b.Definition.ProviderDependencies = append(b.Definition.ProviderDependencies, values...)
	return b
}

// WithRoles appends to the definition's Roles collection.
func (b *Builder) WithRoles(values ...app.PermissionPolicy) *Builder {
	b.Definition.Roles = append(b.Definition.Roles, values...)
	return b
}

// WithPaths appends to the definition's Paths collection.
func (b *Builder) WithPaths(values ...app.Path) *Builder {
	b.Definition.Paths = append(b.Definition.Paths, values...)
	return b
}

// WithUnify appends to the definition's Unify collection.
func (b *Builder) WithUnify(values ...app.IntegrationPoint) *Builder {
	b.Definition.Unify = append(b.Definition.Unify, values...)
	return b
}

// WithActivationSteps appends to the definition's ActivationSteps collection.
func (b *Builder) WithActivationSteps(values ...app.ActivationStep) *Builder {
	b.Definition.ActivationSteps = append(b.Definition.ActivationSteps, values...)
	return b
}

// WithAuthProviders appends to the definition's AuthProviders collection.
func (b *Builder) WithAuthProviders(values ...app.AuthProvider) *Builder {
	b.Definition.AuthProviders = append(b.Definition.AuthProviders, values...)
	return b
}

// WithConnections appends to the definition's Connections collection.
func (b *Builder) WithConnections(values ...app.Connection) *Builder {
	b.Definition.Connections = append(b.Definition.Connections, values...)
	return b
}

// WithScheduledTasks appends to the definition's ScheduledTasks collection.
func (b *Builder) WithScheduledTasks(values ...app.ScheduledTask) *Builder {
	b.Definition.ScheduledTasks = append(b.Definition.ScheduledTasks, values...)
	return b
}

// WithListenToEvents appends to the definition's ListenToEvents collection.
func (b *Builder) WithListenToEvents(values ...app.ScopedKey) *Builder {
	b.Definition.ListenToEvents = append(b.Definition.ListenToEvents, values...)
	return b
}

// WithNavigation appends to the definition's Navigation collection.
func (b *Builder) WithNavigation(values ...app.Navigation) *Builder {
	b.Definition.Navigation = append(b.Definition.Navigation, values...)
	return b
}

// WithConfiguration appends to the definition's Configuration collection.
func (b *Builder) WithConfiguration(values ...app.SettingsPage) *Builder {
	b.Definition.Configuration = append(b.Definition.Configuration, values...)
	return b
}

// WithConfigurationNav appends to the definition's ConfigurationNav collection.
func (b *Builder) WithConfigurationNav(values ...app.EntryPoint) *Builder {
	b.Definition.ConfigurationNav = append(b.Definition.ConfigurationNav, values...)
	return b
}

// WithSearchPatterns appends to the definition's SearchPatterns collection.
func (b *Builder) WithSearchPatterns(values ...app.SearchPattern) *Builder {
	b.Definition.SearchPatterns = append(b.Definition.SearchPatterns, values...)
	return b
}

// WithQuickActions appends to the definition's QuickActions collection.
func (b *Builder) WithQuickActions(values ...app.EntryPoint) *Builder {
	b.Definition.QuickActions = append(b.Definition.QuickActions, values...)
	return b
}

// WithActions appends to the definition's Actions collection.
func (b *Builder) WithActions(values ...app.Action) *Builder {
	b.Definition.Actions = append(b.Definition.Actions, values...)
	return b
}

// WithSearchResults appends to the definition's SearchResults collection.
func (b *Builder) WithSearchResults(values ...app.EntryPoint) *Builder {
	b.Definition.SearchResults = append(b.Definition.SearchResults, values...)
	return b
}

// WithMCPCapabilities appends to the definition's MCPCapabilities collection.
func (b *Builder) WithMCPCapabilities(values ...app.MCPCapability) *Builder {
	b.Definition.MCPCapabilities = append(b.Definition.MCPCapabilities, values...)
	return b
}
