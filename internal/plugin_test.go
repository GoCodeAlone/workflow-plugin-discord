package internal

import (
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
