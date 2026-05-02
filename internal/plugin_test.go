package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

func TestRuntimeTypesMatchManifestCapabilities(t *testing.T) {
	p := New()
	manifest := loadPluginManifest(t)

	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{"module types", p.ModuleTypes(), manifest.Capabilities.ModuleTypes},
		{"step types", p.StepTypes(), manifest.Capabilities.StepTypes},
		{"trigger types", p.TriggerTypes(), manifest.Capabilities.TriggerTypes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := stringSet(tt.got), stringSet(tt.want); !setsEqual(got, want) {
				t.Fatalf("%s = %v, want manifest %v", tt.name, got, want)
			}
		})
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
	wantDescriptors := map[string]struct {
		config string
		output string
	}{
		"module:discord.provider":          {"discord.v1.ProviderConfig", ""},
		"step:step.discord_send_message":   {"discord.v1.SendMessageConfig", "discord.v1.SendMessageOutput"},
		"step:step.discord_send_embed":     {"discord.v1.SendEmbedConfig", "discord.v1.SendEmbedOutput"},
		"step:step.discord_edit_message":   {"discord.v1.EditMessageConfig", "discord.v1.EditMessageOutput"},
		"step:step.discord_delete_message": {"discord.v1.DeleteMessageConfig", "discord.v1.DeleteMessageOutput"},
		"step:step.discord_add_reaction":   {"discord.v1.AddReactionConfig", "discord.v1.AddReactionOutput"},
		"step:step.discord_upload_file":    {"discord.v1.UploadFileConfig", "discord.v1.UploadFileOutput"},
		"step:step.discord_create_thread":  {"discord.v1.CreateThreadConfig", "discord.v1.CreateThreadOutput"},
		"step:step.discord_voice_join":     {"discord.v1.VoiceJoinConfig", "discord.v1.VoiceJoinOutput"},
		"step:step.discord_voice_leave":    {"discord.v1.VoiceLeaveConfig", "discord.v1.VoiceLeaveOutput"},
		"step:step.discord_voice_play":     {"discord.v1.VoicePlayConfig", "discord.v1.VoicePlayOutput"},
		"trigger:trigger.discord":          {"discord.v1.TriggerConfig", "discord.v1.TriggerPayload"},
	}

	got := map[string]bool{}
	for _, contract := range manifest.Contracts {
		key := contract.Kind + ":" + contract.Type
		got[key] = true
		if contract.Mode != "strict" {
			t.Fatalf("%s mode = %q, want strict", key, contract.Mode)
		}
		wantDescriptor, ok := wantDescriptors[key]
		if !ok {
			t.Fatalf("%s has no expected descriptor", key)
		}
		if contract.Config != wantDescriptor.config {
			t.Fatalf("%s config = %q, want %q", key, contract.Config, wantDescriptor.config)
		}
		if contract.Output != wantDescriptor.output {
			t.Fatalf("%s output = %q, want %q", key, contract.Output, wantDescriptor.output)
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
	if runtime.GOOS == "windows" {
		t.Skip("GoReleaser hooks are shell commands")
	}

	config := loadGoReleaserConfig(t)
	if len(config.Before.Hooks) == 0 {
		t.Fatal(".goreleaser.yaml must define before hooks")
	}

	releaseVersion := "9.8.7"
	tmp := t.TempDir()
	copyFile(t, filepath.Join(repoRoot(t), "plugin.json"), filepath.Join(tmp, "plugin.json"))
	binDir := t.TempDir()
	validationLog := filepath.Join(tmp, "validation.log")
	writeExecutable(t, filepath.Join(binDir, "go"), `#!/bin/sh
if [ "$1" = "run" ]; then
	shift
	shift
	exec wfctl "$@"
fi
echo "unexpected go invocation: $*" >&2
exit 42
`)
	writeExecutable(t, filepath.Join(binDir, "wfctl"), `#!/bin/sh
file=""
strict=0
prev=""
for arg in "$@"; do
	if [ "$prev" = "--file" ]; then
		file="$arg"
	fi
	if [ "$arg" = "--strict-contracts" ]; then
		strict=1
	fi
	prev="$arg"
done
if [ "$1" != "plugin" ] || [ "$2" != "validate" ]; then
	echo "unexpected wfctl invocation: $*" >&2
	exit 43
fi
if [ "$strict" != "1" ]; then
	echo "missing --strict-contracts" >&2
	exit 44
fi
if [ -z "$file" ]; then
	echo "missing --file" >&2
	exit 45
fi
grep -q "\"version\": \"$RELEASE_VERSION\"" "$file" || exit 46
grep -q "/releases/download/v$RELEASE_VERSION/" "$file" || exit 47
if grep -q "/releases/download/v0.1.0/" "$file"; then
	exit 48
fi
grep -q "\"mode\": \"strict\"" "$file" || exit 49
printf 'validated %s\n' "$file" >> "$VALIDATION_LOG"
`)

	for _, hook := range config.Before.Hooks {
		cmd := exec.CommandContext(context.Background(), "sh", "-c", strings.ReplaceAll(hook, "{{ .Version }}", releaseVersion))
		cmd.Dir = tmp
		cmd.Env = append(os.Environ(),
			"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
			"RELEASE_VERSION="+releaseVersion,
			"VALIDATION_LOG="+validationLog,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("run GoReleaser hook %q: %v\n%s", hook, err, output)
		}
	}

	manifest := loadPluginManifestFrom(t, filepath.Join(tmp, "plugin.json"))
	if manifest.Version != releaseVersion {
		t.Fatalf("manifest version = %q, want %q", manifest.Version, releaseVersion)
	}
	for _, download := range manifest.Downloads {
		wantPrefix := fmt.Sprintf("https://github.com/GoCodeAlone/workflow-plugin-discord/releases/download/v%s/", releaseVersion)
		if !strings.HasPrefix(download.URL, wantPrefix) {
			t.Fatalf("download URL %q does not use release version %s", download.URL, releaseVersion)
		}
	}
	if data, err := os.ReadFile(validationLog); err != nil || !strings.Contains(string(data), "validated plugin.json") {
		t.Fatalf("strict validation was not invoked; log=%q err=%v", data, err)
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
		Output string `json:"output"`
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
	return loadPluginManifestFrom(t, filepath.Join(repoRoot(t), "plugin.json"))
}

func loadPluginManifestFrom(t *testing.T, path string) pluginManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var manifest pluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return manifest
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
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
