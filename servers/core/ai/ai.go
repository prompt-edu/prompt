package ai

import (
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	"github.com/prompt-edu/prompt/servers/core/standaloneModule"
)

const ServiceName = "AI"

// Module is the AI server, which keeps audit records per person but no data per course phase.
// Without an explicit aiHost (SERVER_AI_HOST) it sits behind the same host as core.
func Module(aiHost, coreHost string) standaloneModule.Module {
	if aiHost == "" {
		aiHost = coreHost
	}
	return standaloneModule.Module{Name: ServiceName, BaseURL: resolution.NormaliseHost(aiHost) + "/ai/api"}
}
