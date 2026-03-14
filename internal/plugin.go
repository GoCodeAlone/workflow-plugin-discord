package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type discordPlugin struct{}

// New returns a new discordPlugin instance.
func New() *discordPlugin { return &discordPlugin{} }

// Manifest returns plugin metadata.
func (p *discordPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "discord",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "Discord messaging, bots, and voice",
	}
}

// ModuleTypes returns the module type names this plugin provides.
func (p *discordPlugin) ModuleTypes() []string {
	return []string{"discord.provider"}
}

// StepTypes returns the step type names this plugin provides.
func (p *discordPlugin) StepTypes() []string {
	return []string{
		"step.discord_send_message",
		"step.discord_send_embed",
		"step.discord_edit_message",
		"step.discord_delete_message",
		"step.discord_add_reaction",
		"step.discord_upload_file",
		"step.discord_create_thread",
		"step.discord_voice_join",
		"step.discord_voice_leave",
		"step.discord_voice_play",
	}
}

// TriggerTypes returns the trigger type names this plugin provides.
func (p *discordPlugin) TriggerTypes() []string {
	return []string{"trigger.discord"}
}

// CreateModule creates a module instance of the given type.
func (p *discordPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "discord.provider":
		return newDiscordProvider(name, config)
	default:
		return nil, fmt.Errorf("discord plugin: unknown module type %q", typeName)
	}
}

// CreateStep creates a step instance of the given type.
func (p *discordPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	// Steps need access to a provider session; they'll resolve it at Execute time
	// via a shared registry keyed by module name.
	switch typeName {
	case "step.discord_send_message":
		return &sendMessageStep{}, nil
	case "step.discord_send_embed":
		return &sendEmbedStep{}, nil
	case "step.discord_edit_message":
		return &editMessageStep{}, nil
	case "step.discord_delete_message":
		return &deleteMessageStep{}, nil
	case "step.discord_add_reaction":
		return &addReactionStep{}, nil
	case "step.discord_upload_file":
		return &uploadFileStep{}, nil
	case "step.discord_create_thread":
		return &createThreadStep{}, nil
	case "step.discord_voice_join":
		return &voiceJoinStep{}, nil
	case "step.discord_voice_leave":
		return &voiceLeaveStep{}, nil
	case "step.discord_voice_play":
		return &voicePlayStep{}, nil
	default:
		return nil, fmt.Errorf("discord plugin: unknown step type %q", typeName)
	}
}

// CreateTrigger creates a trigger instance of the given type.
func (p *discordPlugin) CreateTrigger(typeName string, config map[string]any, cb sdk.TriggerCallback) (sdk.TriggerInstance, error) {
	switch typeName {
	case "trigger.discord":
		return newDiscordTrigger(config, cb)
	default:
		return nil, fmt.Errorf("discord plugin: unknown trigger type %q", typeName)
	}
}
