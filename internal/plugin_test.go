package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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
	p := New()
	manifest := loadPluginManifest(t)
	if got, want := stringSet(p.StepTypes()), stringSet(manifest.Capabilities.StepTypes); !setsEqual(got, want) {
		t.Fatalf("step types = %v, want manifest %v", got, want)
	}
}

func TestPluginContractsMatchRuntimeTypes(t *testing.T) {
	p := New()
	manifest := loadPluginManifest(t)

	want := map[string]bool{}
	for _, moduleType := range p.ModuleTypes() {
		want["module:"+moduleType] = true
	}
	for _, stepType := range p.StepTypes() {
		want["step:"+stepType] = true
	}
	for _, triggerType := range p.TriggerTypes() {
		want["trigger:"+triggerType] = true
	}

	got := map[string]bool{}
	for _, contract := range manifest.Contracts {
		key := contract.Kind + ":" + contract.Type
		got[key] = true
		if contract.Mode != "strict" {
			t.Fatalf("%s mode = %q, want strict", key, contract.Mode)
		}
		if contract.Config == "" {
			t.Fatalf("%s missing config message", key)
		}
	}
	if !setsEqual(got, want) {
		t.Fatalf("contracts = %v, want runtime types %v", got, want)
	}
}

func TestPluginDownloadsMatchGoReleaserMatrix(t *testing.T) {
	manifest := loadPluginManifest(t)
	release := loadGoReleaserConfig(t)
	if len(release.Builds) != 1 {
		t.Fatalf("goreleaser builds = %d, want 1", len(release.Builds))
	}
	build := release.Builds[0]

	want := map[string]string{}
	for _, goos := range build.Goos {
		for _, goarch := range build.Goarch {
			key := goos + "/" + goarch
			want[key] = fmt.Sprintf("https://github.com/GoCodeAlone/workflow-plugin-discord/releases/download/v%s/workflow-plugin-discord-%s-%s.tar.gz", manifest.Version, goos, goarch)
		}
	}

	got := map[string]string{}
	for _, download := range manifest.Downloads {
		got[download.OS+"/"+download.Arch] = download.URL
	}
	if !setsEqual(stringSetFromMap(got), stringSetFromMap(want)) {
		t.Fatalf("download matrix = %v, want %v", got, want)
	}
	for key, wantURL := range want {
		if gotURL := got[key]; gotURL != wantURL {
			t.Fatalf("download %s = %q, want %q", key, gotURL, wantURL)
		}
	}
}

func TestGoReleaserValidatesRewrittenPluginManifest(t *testing.T) {
	config := loadGoReleaserConfig(t)
	hooks := strings.Join(config.Before.Hooks, "\n")
	for _, want := range []string{
		`sed -i.bak 's/"version": ".*"/"version": "{{ .Version }}"/' plugin.json`,
		`sed -i.bak 's|/releases/download/v[^/]*/|/releases/download/v{{ .Version }}/|g' plugin.json`,
		`wfctl@v0.20.1 plugin validate --file plugin.json --strict-contracts`,
	} {
		if !strings.Contains(hooks, want) {
			t.Fatalf(".goreleaser.yaml missing %q", want)
		}
	}
}

func TestCreateTriggerMissingToken(t *testing.T) {
	p := New()
	_, err := p.CreateTrigger("trigger.discord", map[string]any{}, func(string, map[string]any) error { return nil })
	if err == nil {
		t.Error("expected error when token is missing")
	}
}

type pluginManifest struct {
	Version      string `json:"version"`
	Capabilities struct {
		ModuleTypes  []string `json:"moduleTypes"`
		StepTypes    []string `json:"stepTypes"`
		TriggerTypes []string `json:"triggerTypes"`
	} `json:"capabilities"`
	Contracts []struct {
		Kind   string `json:"kind"`
		Type   string `json:"type"`
		Mode   string `json:"mode"`
		Config string `json:"config"`
	} `json:"contracts"`
	Downloads []struct {
		OS   string `json:"os"`
		Arch string `json:"arch"`
		URL  string `json:"url"`
	} `json:"downloads"`
}

type goReleaserConfig struct {
	Before struct {
		Hooks []string `yaml:"hooks"`
	} `yaml:"before"`
	Builds []struct {
		Goos   []string `yaml:"goos"`
		Goarch []string `yaml:"goarch"`
	} `yaml:"builds"`
}

func loadPluginManifest(t *testing.T) pluginManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "plugin.json"))
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var manifest pluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}
	return manifest
}

func loadGoReleaserConfig(t *testing.T) goReleaserConfig {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}
	var config goReleaserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse .goreleaser.yaml: %v", err)
	}
	return config
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func stringSetFromMap(values map[string]string) map[string]bool {
	set := make(map[string]bool, len(values))
	for value := range values {
		set[value] = true
	}
	return set
}

func setsEqual(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if !right[key] {
			return false
		}
	}
	return true
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
