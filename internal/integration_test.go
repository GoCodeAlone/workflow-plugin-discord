package internal_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow/wftest"
)

// TestIntegration_SendMessage verifies a pipeline that mocks discord_send_message
// and then sets a confirmation value via the built-in step.set step.
func TestIntegration_SendMessage(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  notify:
    steps:
      - name: send
        type: step.discord_send_message
        config:
          channel_id: "123456789"
          content: "hello from wftest"
      - name: confirm
        type: step.set
        config:
          values:
            sent: true
`),
		wftest.MockStep("step.discord_send_message", wftest.Returns(map[string]any{
			"id":         "msg-001",
			"message_id": "msg-001",
			"channel_id": "123456789",
		})),
	)

	result := h.ExecutePipeline("notify", nil)
	if result.Error != nil {
		t.Fatalf("pipeline error: %v", result.Error)
	}
	if result.Output["sent"] != true {
		t.Errorf("expected sent=true, got %v", result.Output["sent"])
	}
	sendOut := result.StepOutput("send")
	if sendOut["message_id"] != "msg-001" {
		t.Errorf("expected message_id=msg-001, got %v", sendOut["message_id"])
	}
}

// TestIntegration_SendEmbed verifies a pipeline that mocks discord_send_embed
// and captures the returned embed message ID.
func TestIntegration_SendEmbed(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  embed-notify:
    steps:
      - name: embed
        type: step.discord_send_embed
        config:
          channel_id: "987654321"
          title: "Alert"
          description: "Something happened"
          color: 16711680
      - name: confirm
        type: step.set
        config:
          values:
            embed_sent: true
`),
		wftest.MockStep("step.discord_send_embed", wftest.Returns(map[string]any{
			"id":         "embed-msg-002",
			"message_id": "embed-msg-002",
			"channel_id": "987654321",
		})),
	)

	result := h.ExecutePipeline("embed-notify", nil)
	if result.Error != nil {
		t.Fatalf("pipeline error: %v", result.Error)
	}
	if result.Output["embed_sent"] != true {
		t.Errorf("expected embed_sent=true, got %v", result.Output["embed_sent"])
	}
	embedOut := result.StepOutput("embed")
	if embedOut["message_id"] != "embed-msg-002" {
		t.Errorf("expected message_id=embed-msg-002, got %v", embedOut["message_id"])
	}
}

// TestIntegration_AddReaction verifies a pipeline that mocks discord_add_reaction
// and checks the reaction confirmation output.
func TestIntegration_AddReaction(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  react:
    steps:
      - name: reaction
        type: step.discord_add_reaction
        config:
          channel_id: "111222333"
          message_id: "msg-abc"
          emoji: "👍"
      - name: confirm
        type: step.set
        config:
          values:
            reaction_done: true
`),
		wftest.MockStep("step.discord_add_reaction", wftest.Returns(map[string]any{
			"reacted": true,
			"success": true,
		})),
	)

	result := h.ExecutePipeline("react", nil)
	if result.Error != nil {
		t.Fatalf("pipeline error: %v", result.Error)
	}
	if result.Output["reaction_done"] != true {
		t.Errorf("expected reaction_done=true, got %v", result.Output["reaction_done"])
	}
	reactionOut := result.StepOutput("reaction")
	if reactionOut["reacted"] != true {
		t.Errorf("expected reacted=true, got %v", reactionOut["reacted"])
	}
	if reactionOut["success"] != true {
		t.Errorf("expected success=true, got %v", reactionOut["success"])
	}
}
