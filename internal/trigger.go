package internal

import (
	"context"
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/bwmarrin/discordgo"
)

type discordTrigger struct {
	session  *discordgo.Session
	callback sdk.TriggerCallback
	cancel   context.CancelFunc
}

func newDiscordTrigger(config map[string]any, cb sdk.TriggerCallback) (*discordTrigger, error) {
	token, _ := config["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("trigger.discord: token is required")
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("trigger.discord: create session: %w", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMembers
	return &discordTrigger{session: dg, callback: cb}, nil
}

func (t *discordTrigger) Start(ctx context.Context) error {
	_, t.cancel = context.WithCancel(ctx)

	t.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author != nil && m.Author.Bot {
			return
		}
		_ = t.callback("fire", map[string]any{
			"type":       "message_create",
			"channel_id": m.ChannelID,
			"message_id": m.ID,
			"content":    m.Content,
			"author_id":  m.Author.ID,
			"guild_id":   m.GuildID,
		})
	})

	t.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		_ = t.callback("fire", map[string]any{
			"type":       "message_update",
			"channel_id": m.ChannelID,
			"message_id": m.ID,
			"content":    m.Content,
			"guild_id":   m.GuildID,
		})
	})

	t.session.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		_ = t.callback("fire", map[string]any{
			"type":       "reaction_add",
			"channel_id": r.ChannelID,
			"message_id": r.MessageID,
			"user_id":    r.UserID,
			"emoji":      r.Emoji.Name,
			"guild_id":   r.GuildID,
		})
	})

	t.session.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		_ = t.callback("fire", map[string]any{
			"type":     "member_join",
			"guild_id": m.GuildID,
			"user_id":  m.User.ID,
		})
	})

	return t.session.Open()
}

func (t *discordTrigger) Stop(ctx context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}
	return t.session.Close()
}
