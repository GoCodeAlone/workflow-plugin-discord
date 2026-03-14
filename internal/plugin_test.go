package internal

import (
	"context"
	"testing"
)

func TestManifest(t *testing.T) {
	p := New()
	m := p.Manifest()
	if m.Name != "discord" {
		t.Errorf("expected name discord, got %s", m.Name)
	}
	if m.Version == "" {
		t.Error("version must not be empty")
	}
}

func TestModuleTypes(t *testing.T) {
	p := New()
	types := p.ModuleTypes()
	if len(types) != 1 || types[0] != "discord.provider" {
		t.Errorf("unexpected module types: %v", types)
	}
}

func TestStepTypes(t *testing.T) {
	p := New()
	want := map[string]bool{
		"step.discord_send_message":   true,
		"step.discord_send_embed":     true,
		"step.discord_edit_message":   true,
		"step.discord_delete_message": true,
		"step.discord_add_reaction":   true,
		"step.discord_upload_file":    true,
		"step.discord_create_thread":  true,
		"step.discord_voice_join":     true,
		"step.discord_voice_leave":    true,
		"step.discord_voice_play":     true,
	}
	for _, st := range p.StepTypes() {
		if !want[st] {
			t.Errorf("unexpected step type: %s", st)
		}
		delete(want, st)
	}
	for missing := range want {
		t.Errorf("missing step type: %s", missing)
	}
}

func TestTriggerTypes(t *testing.T) {
	p := New()
	types := p.TriggerTypes()
	if len(types) != 1 || types[0] != "trigger.discord" {
		t.Errorf("unexpected trigger types: %v", types)
	}
}

func TestCreateModuleUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateModule("unknown.type", "test", nil)
	if err == nil {
		t.Error("expected error for unknown module type")
	}
}

func TestCreateStepUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateStep("unknown.step", "test", nil)
	if err == nil {
		t.Error("expected error for unknown step type")
	}
}

func TestCreateTriggerUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateTrigger("unknown.trigger", nil, nil)
	if err == nil {
		t.Error("expected error for unknown trigger type")
	}
}

func TestCreateModuleMissingToken(t *testing.T) {
	p := New()
	_, err := p.CreateModule("discord.provider", "test", map[string]any{})
	if err == nil {
		t.Error("expected error when token is missing")
	}
}

func TestCreateStepAllTypes(t *testing.T) {
	p := New()
	for _, st := range p.StepTypes() {
		step, err := p.CreateStep(st, "test", map[string]any{})
		if err != nil {
			t.Errorf("CreateStep(%s) error: %v", st, err)
		}
		if step == nil {
			t.Errorf("CreateStep(%s) returned nil", st)
		}
	}
}

func TestStepTypesMatchManifest(t *testing.T) {
	// Verify the step count matches plugin.json capabilities
	p := New()
	if len(p.StepTypes()) != 10 {
		t.Errorf("expected 10 step types to match plugin.json, got %d", len(p.StepTypes()))
	}
}

func TestCreateTriggerMissingToken(t *testing.T) {
	p := New()
	_, err := p.CreateTrigger("trigger.discord", map[string]any{}, func(string, map[string]any) error { return nil })
	if err == nil {
		t.Error("expected error when token is missing")
	}
}

// withTestProvider registers a fake discordProvider for Execute-level tests.
func withTestProvider(t *testing.T, name string) {
	t.Helper()
	prov := &discordProvider{name: name}
	providerRegistry.Store(name, prov)
	t.Cleanup(func() { providerRegistry.Delete(name) })
}

func TestStepExecuteMissingProvider(t *testing.T) {
	_, err := (&sendMessageStep{}).Execute(context.Background(), nil, nil,
		map[string]any{"provider": "nonexistent", "channel_id": "C1", "content": "hi"},
		nil, nil)
	if err == nil {
		t.Error("expected error for unregistered provider")
	}
}

func TestStepExecuteMissingRequiredFields(t *testing.T) {
	withTestProvider(t, "test-discord")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"send_message", func() error {
			_, err := (&sendMessageStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"send_embed", func() error {
			_, err := (&sendEmbedStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"edit_message", func() error {
			_, err := (&editMessageStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"delete_message", func() error {
			_, err := (&deleteMessageStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"add_reaction", func() error {
			_, err := (&addReactionStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"upload_file", func() error {
			_, err := (&uploadFileStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"create_thread", func() error {
			_, err := (&createThreadStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"voice_join", func() error {
			_, err := (&voiceJoinStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"voice_leave", func() error {
			_, err := (&voiceLeaveStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
		{"voice_play", func() error {
			_, err := (&voicePlayStep{}).Execute(context.Background(), nil, nil,
				map[string]any{"provider": "test-discord"}, nil, nil)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err == nil {
				t.Errorf("%s: expected validation error for missing required fields", tt.name)
			}
		})
	}
}

func TestVoicePlayReturnsError(t *testing.T) {
	withTestProvider(t, "test-discord-play")
	// voice_play is a stub that always returns an error (requires opus encoding)
	_, err := (&voicePlayStep{}).Execute(context.Background(), nil, nil,
		map[string]any{"provider": "test-discord-play", "guild_id": "G123"}, nil, nil)
	if err == nil {
		t.Error("expected voice_play to return an error (opus stub)")
	}
}
