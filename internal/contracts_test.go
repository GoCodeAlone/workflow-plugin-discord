package internal

import (
	"testing"

	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
)

func TestContractRegistry_NonNil(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()
	if reg == nil {
		t.Fatal("ContractRegistry() returned nil")
	}
	if reg.FileDescriptorSet == nil {
		t.Fatal("ContractRegistry().FileDescriptorSet is nil")
	}
	if len(reg.FileDescriptorSet.File) == 0 {
		t.Fatal("ContractRegistry().FileDescriptorSet.File is empty")
	}
	if len(reg.Contracts) == 0 {
		t.Fatal("ContractRegistry().Contracts is empty")
	}
}

func TestContractRegistry_CoversModuleType(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()
	found := false
	for _, c := range reg.Contracts {
		if c.Kind == pb.ContractKind_CONTRACT_KIND_MODULE && c.ModuleType == "discord.provider" {
			if c.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
				t.Errorf("discord.provider contract mode = %v, want STRICT_PROTO", c.Mode)
			}
			if c.ConfigMessage == "" {
				t.Error("discord.provider contract has empty ConfigMessage")
			}
			found = true
		}
	}
	if !found {
		t.Error("no contract descriptor found for module type discord.provider")
	}
}

func TestContractRegistry_CoversAllStepTypes(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()

	stepContracts := make(map[string]*pb.ContractDescriptor)
	for _, c := range reg.Contracts {
		if c.Kind == pb.ContractKind_CONTRACT_KIND_STEP {
			stepContracts[c.StepType] = c
		}
	}

	expected := []string{
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

	for _, stepType := range expected {
		c, ok := stepContracts[stepType]
		if !ok {
			t.Errorf("no contract descriptor found for step type %q", stepType)
			continue
		}
		if c.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
			t.Errorf("%s contract mode = %v, want STRICT_PROTO", stepType, c.Mode)
		}
		if c.ConfigMessage == "" {
			t.Errorf("%s contract has empty ConfigMessage", stepType)
		}
		if c.InputMessage == "" {
			t.Errorf("%s contract has empty InputMessage", stepType)
		}
		if c.OutputMessage == "" {
			t.Errorf("%s contract has empty OutputMessage", stepType)
		}
	}
}

func TestContractRegistry_CoversTriggerType(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()
	found := false
	for _, c := range reg.Contracts {
		if c.Kind == pb.ContractKind_CONTRACT_KIND_TRIGGER && c.TriggerType == "trigger.discord" {
			if c.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
				t.Errorf("trigger.discord contract mode = %v, want STRICT_PROTO", c.Mode)
			}
			if c.ConfigMessage == "" {
				t.Error("trigger.discord contract has empty ConfigMessage")
			}
			if c.OutputMessage == "" {
				t.Error("trigger.discord contract has empty OutputMessage")
			}
			found = true
		}
	}
	if !found {
		t.Error("no contract descriptor found for trigger type trigger.discord")
	}
}

func TestContractRegistry_FileDescriptorSetContainsDiscordFile(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()

	found := false
	for _, f := range reg.FileDescriptorSet.File {
		if f.GetName() == "discord.proto" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FileDescriptorSet does not contain discord.proto")
	}
}

func TestContractRegistry_FileDescriptorSetContainsStructFile(t *testing.T) {
	p := New()
	reg := p.ContractRegistry()

	found := false
	for _, f := range reg.FileDescriptorSet.File {
		if f.GetName() == "google/protobuf/struct.proto" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FileDescriptorSet does not contain google/protobuf/struct.proto")
	}
}
