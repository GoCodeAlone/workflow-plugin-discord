package internal

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// providerRegistry holds active discord provider sessions keyed by module name.
var providerRegistry = &sync.Map{}

type discordProvider struct {
	name    string
	token   string
	session *discordgo.Session
}

func newDiscordProvider(name string, config map[string]any) (*discordProvider, error) {
	token, _ := config["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("discord.provider: token is required")
	}
	return &discordProvider{name: name, token: token}, nil
}

func (m *discordProvider) Init() error {
	dg, err := discordgo.New("Bot " + m.token)
	if err != nil {
		return fmt.Errorf("discord session: %w", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessageReactions
	m.session = dg
	providerRegistry.Store(m.name, m)
	return nil
}

func (m *discordProvider) Start(ctx context.Context) error {
	return m.session.Open()
}

func (m *discordProvider) Stop(ctx context.Context) error {
	providerRegistry.Delete(m.name)
	return m.session.Close()
}

// Name returns the platform identifier.
func (m *discordProvider) Name() string { return "discord" }

// SendMessage sends a plain text message to a channel.
func (m *discordProvider) SendMessage(ctx context.Context, channelID, content string, opts interface{}) (string, error) {
	msg, err := m.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return "", fmt.Errorf("discord send: %w", err)
	}
	return msg.ID, nil
}

// EditMessage updates an existing message.
func (m *discordProvider) EditMessage(ctx context.Context, channelID, messageID, content string) error {
	_, err := m.session.ChannelMessageEdit(channelID, messageID, content)
	return err
}

// DeleteMessage removes a message.
func (m *discordProvider) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	return m.session.ChannelMessageDelete(channelID, messageID)
}

// React adds a reaction emoji to a message.
func (m *discordProvider) React(ctx context.Context, channelID, messageID, emoji string) error {
	return m.session.MessageReactionAdd(channelID, messageID, emoji)
}

// UploadFile sends a file to a channel.
func (m *discordProvider) UploadFile(ctx context.Context, channelID string, file io.Reader, filename string) (string, error) {
	msg, err := m.session.ChannelFileSend(channelID, filename, file)
	if err != nil {
		return "", fmt.Errorf("discord upload: %w", err)
	}
	return msg.ID, nil
}

// JoinVoice connects the bot to a voice channel.
func (m *discordProvider) JoinVoice(ctx context.Context, guildID, channelID string) error {
	_, err := m.session.ChannelVoiceJoin(guildID, channelID, false, false)
	return err
}

// LeaveVoice disconnects from a voice channel in the given guild.
func (m *discordProvider) LeaveVoice(ctx context.Context, guildID string) error {
	vc, ok := m.session.VoiceConnections[guildID]
	if !ok {
		return nil
	}
	return vc.Disconnect()
}

// PlayAudio streams audio to a connected voice channel (stub — requires opus encoding).
func (m *discordProvider) PlayAudio(ctx context.Context, guildID string, audio io.Reader) error {
	return fmt.Errorf("discord voice play: not implemented (requires opus-encoded frames)")
}

// getProvider resolves a provider from the registry by module name.
func getProvider(providerName string) (*discordProvider, error) {
	v, ok := providerRegistry.Load(providerName)
	if !ok {
		return nil, fmt.Errorf("discord: provider %q not found (is the discord.provider module configured?)", providerName)
	}
	p, ok := v.(*discordProvider)
	if !ok {
		return nil, fmt.Errorf("discord: invalid provider type for %q", providerName)
	}
	return p, nil
}
