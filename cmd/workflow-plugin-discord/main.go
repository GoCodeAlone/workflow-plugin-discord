// Command workflow-plugin-discord is a workflow engine external plugin that
// provides Discord messaging, bot events, and voice channel capabilities.
// It runs as a subprocess and communicates with the host workflow engine via
// the go-plugin protocol.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-discord/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.New())
}
