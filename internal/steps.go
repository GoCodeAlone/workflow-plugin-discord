package internal

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/bwmarrin/discordgo"
)

// resolveProvider looks up the discord.provider name from current map and returns the session.
func resolveProvider(current map[string]any) (*discordProvider, error) {
	name, _ := current["provider"].(string)
	if name == "" {
		name = "discord" // default provider name
	}
	return getProvider(name)
}

// --- step.discord_send_message ---

type sendMessageStep struct{}

func (s *sendMessageStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	content, _ := current["content"].(string)
	if channelID == "" || content == "" {
		return nil, fmt.Errorf("discord_send_message: channel_id and content are required")
	}
	msg, err := p.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return nil, fmt.Errorf("discord_send_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":         msg.ID,
		"message_id": msg.ID,
		"channel_id": msg.ChannelID,
	}}, nil
}

// --- step.discord_send_embed ---

type sendEmbedStep struct{}

func (s *sendEmbedStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	title, _ := current["title"].(string)
	description, _ := current["description"].(string)
	color, _ := current["color"].(int)
	if channelID == "" {
		return nil, fmt.Errorf("discord_send_embed: channel_id is required")
	}
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}
	msg, err := p.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return nil, fmt.Errorf("discord_send_embed: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":         msg.ID,
		"message_id": msg.ID,
		"channel_id": msg.ChannelID,
	}}, nil
}

// --- step.discord_edit_message ---

type editMessageStep struct{}

func (s *editMessageStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	messageID, _ := current["message_id"].(string)
	content, _ := current["content"].(string)
	if channelID == "" || messageID == "" || content == "" {
		return nil, fmt.Errorf("discord_edit_message: channel_id, message_id, and content are required")
	}
	msg, err := p.session.ChannelMessageEdit(channelID, messageID, content)
	if err != nil {
		return nil, fmt.Errorf("discord_edit_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{
		"message_id": msg.ID,
	}}, nil
}

// --- step.discord_delete_message ---

type deleteMessageStep struct{}

func (s *deleteMessageStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	messageID, _ := current["message_id"].(string)
	if channelID == "" || messageID == "" {
		return nil, fmt.Errorf("discord_delete_message: channel_id and message_id are required")
	}
	if err := p.session.ChannelMessageDelete(channelID, messageID); err != nil {
		return nil, fmt.Errorf("discord_delete_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"deleted": true}}, nil
}

// --- step.discord_add_reaction ---

type addReactionStep struct{}

func (s *addReactionStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	messageID, _ := current["message_id"].(string)
	emoji, _ := current["emoji"].(string)
	if channelID == "" || messageID == "" || emoji == "" {
		return nil, fmt.Errorf("discord_add_reaction: channel_id, message_id, and emoji are required")
	}
	if err := p.session.MessageReactionAdd(channelID, messageID, emoji); err != nil {
		return nil, fmt.Errorf("discord_add_reaction: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"reacted": true, "success": true}}, nil
}

// --- step.discord_upload_file ---

type uploadFileStep struct{}

func (s *uploadFileStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	filename, _ := current["filename"].(string)
	content, _ := current["content"].(string)
	if channelID == "" || filename == "" {
		return nil, fmt.Errorf("discord_upload_file: channel_id and filename are required")
	}
	msg, err := p.session.ChannelFileSend(channelID, filename, strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("discord_upload_file: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{
		"message_id": msg.ID,
		"channel_id": msg.ChannelID,
	}}, nil
}

// --- step.discord_create_thread ---

type createThreadStep struct{}

func (s *createThreadStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	channelID, _ := current["channel_id"].(string)
	messageID, _ := current["message_id"].(string)
	name, _ := current["name"].(string)
	if channelID == "" || name == "" {
		return nil, fmt.Errorf("discord_create_thread: channel_id and name are required")
	}

	var thread *discordgo.Channel
	if messageID != "" {
		thread, err = p.session.MessageThreadStartComplex(channelID, messageID, &discordgo.ThreadStart{
			Name: name,
		})
	} else {
		thread, err = p.session.ThreadStartComplex(channelID, &discordgo.ThreadStart{
			Name: name,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("discord_create_thread: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{
		"thread_id": thread.ID,
		"name":      thread.Name,
	}}, nil
}

// --- step.discord_voice_join ---

type voiceJoinStep struct{}

func (s *voiceJoinStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	guildID, _ := current["guild_id"].(string)
	channelID, _ := current["channel_id"].(string)
	if guildID == "" || channelID == "" {
		return nil, fmt.Errorf("discord_voice_join: guild_id and channel_id are required")
	}
	_, err = p.session.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		return nil, fmt.Errorf("discord_voice_join: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"joined": true}}, nil
}

// --- step.discord_voice_play ---

type voicePlayStep struct{}

func (s *voicePlayStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	guildID, _ := current["guild_id"].(string)
	if guildID == "" {
		return nil, fmt.Errorf("discord_voice_play: guild_id is required")
	}
	if err := p.PlayAudio(ctx, guildID, strings.NewReader("")); err != nil {
		return nil, fmt.Errorf("discord_voice_play: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"playing": true}}, nil
}

// --- step.discord_voice_leave ---

type voiceLeaveStep struct{}

func (s *voiceLeaveStep) Execute(ctx context.Context, triggerData map[string]any, stepOutputs map[string]map[string]any, current map[string]any, metadata map[string]any, cfg map[string]any) (*sdk.StepResult, error) {
	p, err := resolveProvider(current)
	if err != nil {
		return nil, err
	}
	guildID, _ := current["guild_id"].(string)
	if guildID == "" {
		return nil, fmt.Errorf("discord_voice_leave: guild_id is required")
	}
	vc, ok := p.session.VoiceConnections[guildID]
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"left": false, "reason": "not connected"}}, nil
	}
	if err := vc.Disconnect(); err != nil {
		return nil, fmt.Errorf("discord_voice_leave: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"left": true}}, nil
}

