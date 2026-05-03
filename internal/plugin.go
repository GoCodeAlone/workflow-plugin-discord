package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// Version is set at build time via -ldflags
// "-X github.com/GoCodeAlone/workflow-plugin-discord/internal.Version=X.Y.Z".
// Default is a bare semver so plugin loaders that validate semver accept
// unreleased dev builds; goreleaser overrides with the real release tag.
var Version = "0.0.0"

type discordPlugin struct{}

type contractDescriptor struct {
	Kind   string
	Type   string
	Config string
	Output string
}

type moduleRegistration struct {
	typeName string
	config   string
	create   func(name string, config map[string]any) (sdk.ModuleInstance, error)
}

type stepRegistration struct {
	typeName string
	config   string
	output   string
	create   func() sdk.StepInstance
}

type triggerRegistration struct {
	typeName string
	config   string
	output   string
	create   func(config map[string]any, cb sdk.TriggerCallback) (sdk.TriggerInstance, error)
}

var discordModuleRegistrations = []moduleRegistration{
	{
		typeName: "discord.provider",
		config:   "discord.v1.ProviderConfig",
		create: func(name string, config map[string]any) (sdk.ModuleInstance, error) {
			return newDiscordProvider(name, config)
		},
	},
}

var discordStepRegistrations = []stepRegistration{
	{"step.discord_send_message", "discord.v1.SendMessageConfig", "discord.v1.SendMessageOutput", func() sdk.StepInstance { return &sendMessageStep{} }},
	{"step.discord_send_embed", "discord.v1.SendEmbedConfig", "discord.v1.SendEmbedOutput", func() sdk.StepInstance { return &sendEmbedStep{} }},
	{"step.discord_edit_message", "discord.v1.EditMessageConfig", "discord.v1.EditMessageOutput", func() sdk.StepInstance { return &editMessageStep{} }},
	{"step.discord_delete_message", "discord.v1.DeleteMessageConfig", "discord.v1.DeleteMessageOutput", func() sdk.StepInstance { return &deleteMessageStep{} }},
	{"step.discord_add_reaction", "discord.v1.AddReactionConfig", "discord.v1.AddReactionOutput", func() sdk.StepInstance { return &addReactionStep{} }},
	{"step.discord_upload_file", "discord.v1.UploadFileConfig", "discord.v1.UploadFileOutput", func() sdk.StepInstance { return &uploadFileStep{} }},
	{"step.discord_create_thread", "discord.v1.CreateThreadConfig", "discord.v1.CreateThreadOutput", func() sdk.StepInstance { return &createThreadStep{} }},
	{"step.discord_voice_join", "discord.v1.VoiceJoinConfig", "discord.v1.VoiceJoinOutput", func() sdk.StepInstance { return &voiceJoinStep{} }},
	{"step.discord_voice_leave", "discord.v1.VoiceLeaveConfig", "discord.v1.VoiceLeaveOutput", func() sdk.StepInstance { return &voiceLeaveStep{} }},
	{"step.discord_voice_play", "discord.v1.VoicePlayConfig", "discord.v1.VoicePlayOutput", func() sdk.StepInstance { return &voicePlayStep{} }},
}

var discordTriggerRegistrations = []triggerRegistration{
	{
		typeName: "trigger.discord",
		config:   "discord.v1.TriggerConfig",
		output:   "discord.v1.TriggerPayload",
		create: func(config map[string]any, cb sdk.TriggerCallback) (sdk.TriggerInstance, error) {
			return newDiscordTrigger(config, cb)
		},
	},
}

// New returns a new discordPlugin instance.
func New() *discordPlugin { return &discordPlugin{} }

// Manifest returns plugin metadata.
func (p *discordPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "discord",
		Version:     Version,
		Author:      "GoCodeAlone",
		Description: "Discord messaging, bots, and voice",
	}
}

// ModuleTypes returns the module type names this plugin provides.
func (p *discordPlugin) ModuleTypes() []string {
	types := make([]string, 0, len(discordModuleRegistrations))
	for _, registration := range discordModuleRegistrations {
		types = append(types, registration.typeName)
	}
	return types
}

// StepTypes returns the step type names this plugin provides.
func (p *discordPlugin) StepTypes() []string {
	types := make([]string, 0, len(discordStepRegistrations))
	for _, registration := range discordStepRegistrations {
		types = append(types, registration.typeName)
	}
	return types
}

// TriggerTypes returns the trigger type names this plugin provides.
func (p *discordPlugin) TriggerTypes() []string {
	types := make([]string, 0, len(discordTriggerRegistrations))
	for _, registration := range discordTriggerRegistrations {
		types = append(types, registration.typeName)
	}
	return types
}

func (p *discordPlugin) contractDescriptors() []contractDescriptor {
	descriptors := make([]contractDescriptor, 0, len(discordModuleRegistrations)+len(discordStepRegistrations)+len(discordTriggerRegistrations))
	for _, registration := range discordModuleRegistrations {
		descriptors = append(descriptors, contractDescriptor{
			Kind:   "module",
			Type:   registration.typeName,
			Config: registration.config,
		})
	}
	for _, registration := range discordStepRegistrations {
		descriptors = append(descriptors, contractDescriptor{
			Kind:   "step",
			Type:   registration.typeName,
			Config: registration.config,
			Output: registration.output,
		})
	}
	for _, registration := range discordTriggerRegistrations {
		descriptors = append(descriptors, contractDescriptor{
			Kind:   "trigger",
			Type:   registration.typeName,
			Config: registration.config,
			Output: registration.output,
		})
	}
	return descriptors
}

// CreateModule creates a module instance of the given type.
func (p *discordPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	for _, registration := range discordModuleRegistrations {
		if registration.typeName == typeName {
			return registration.create(name, config)
		}
	}
	return nil, fmt.Errorf("discord plugin: unknown module type %q", typeName)
}

// CreateStep creates a step instance of the given type.
func (p *discordPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	// Steps need access to a provider session; they'll resolve it at Execute time
	// via a shared registry keyed by module name.
	for _, registration := range discordStepRegistrations {
		if registration.typeName == typeName {
			return registration.create(), nil
		}
	}
	return nil, fmt.Errorf("discord plugin: unknown step type %q", typeName)
}

// CreateTrigger creates a trigger instance of the given type.
func (p *discordPlugin) CreateTrigger(typeName string, config map[string]any, cb sdk.TriggerCallback) (sdk.TriggerInstance, error) {
	for _, registration := range discordTriggerRegistrations {
		if registration.typeName == typeName {
			return registration.create(config, cb)
		}
	}
	return nil, fmt.Errorf("discord plugin: unknown trigger type %q", typeName)
}
