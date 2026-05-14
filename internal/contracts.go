package internal

import (
	discordv1 "github.com/GoCodeAlone/workflow-plugin-discord/gen"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// ContractRegistry returns the typed contract descriptors for the discord.provider
// module, all 10 step types, and the trigger.discord trigger. The workflow engine
// calls this via the sdk.ContractProvider interface to resolve proto message types
// for strict validation.
func (p *discordPlugin) ContractRegistry() *pb.ContractRegistry {
	return discordContractRegistry
}

// discordContractRegistry declares STRICT_PROTO contracts for all discord
// module, step, and trigger types. The FileDescriptorSet includes
// google.protobuf.Struct (used in Input fields) so the engine can resolve
// all message types by full name.
var discordContractRegistry = &pb.ContractRegistry{
	FileDescriptorSet: &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			protodesc.ToFileDescriptorProto(structpb.File_google_protobuf_struct_proto),
			protodesc.ToFileDescriptorProto(discordv1.File_discord_proto),
		},
	},
	Contracts: []*pb.ContractDescriptor{
		// ── module ───────────────────────────────────────────────────────────────
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_MODULE,
			ModuleType:    "discord.provider",
			ConfigMessage: discordProtoPkg + "ProviderConfig",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		// ── steps ────────────────────────────────────────────────────────────────
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_send_message",
			ConfigMessage: discordProtoPkg + "SendMessageConfig",
			InputMessage:  discordProtoPkg + "SendMessageInput",
			OutputMessage: discordProtoPkg + "SendMessageOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_send_embed",
			ConfigMessage: discordProtoPkg + "SendEmbedConfig",
			InputMessage:  discordProtoPkg + "SendEmbedInput",
			OutputMessage: discordProtoPkg + "SendEmbedOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_edit_message",
			ConfigMessage: discordProtoPkg + "EditMessageConfig",
			InputMessage:  discordProtoPkg + "EditMessageInput",
			OutputMessage: discordProtoPkg + "EditMessageOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_delete_message",
			ConfigMessage: discordProtoPkg + "DeleteMessageConfig",
			InputMessage:  discordProtoPkg + "DeleteMessageInput",
			OutputMessage: discordProtoPkg + "DeleteMessageOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_add_reaction",
			ConfigMessage: discordProtoPkg + "AddReactionConfig",
			InputMessage:  discordProtoPkg + "AddReactionInput",
			OutputMessage: discordProtoPkg + "AddReactionOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_upload_file",
			ConfigMessage: discordProtoPkg + "UploadFileConfig",
			InputMessage:  discordProtoPkg + "UploadFileInput",
			OutputMessage: discordProtoPkg + "UploadFileOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_create_thread",
			ConfigMessage: discordProtoPkg + "CreateThreadConfig",
			InputMessage:  discordProtoPkg + "CreateThreadInput",
			OutputMessage: discordProtoPkg + "CreateThreadOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_voice_join",
			ConfigMessage: discordProtoPkg + "VoiceJoinConfig",
			InputMessage:  discordProtoPkg + "VoiceJoinInput",
			OutputMessage: discordProtoPkg + "VoiceJoinOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_voice_leave",
			ConfigMessage: discordProtoPkg + "VoiceLeaveConfig",
			InputMessage:  discordProtoPkg + "VoiceLeaveInput",
			OutputMessage: discordProtoPkg + "VoiceLeaveOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
			StepType:      "step.discord_voice_play",
			ConfigMessage: discordProtoPkg + "VoicePlayConfig",
			InputMessage:  discordProtoPkg + "VoicePlayInput",
			OutputMessage: discordProtoPkg + "VoicePlayOutput",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		// ── trigger ───────────────────────────────────────────────────────────────
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_TRIGGER,
			TriggerType:   "trigger.discord",
			ConfigMessage: discordProtoPkg + "TriggerConfig",
			OutputMessage: discordProtoPkg + "TriggerPayload",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
	},
}

// discordProtoPkg is the proto package prefix for all discord typed messages.
const discordProtoPkg = "workflow.plugin.discord.v1."

// Compile-time assertion: discordPlugin implements sdk.ContractProvider.
var _ interface{ ContractRegistry() *pb.ContractRegistry } = (*discordPlugin)(nil)
