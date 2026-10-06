package ai

import (
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	"github.com/prompt-edu/prompt/servers/core/standaloneModule"
)

const ServiceName = "AI"

// Module is the AI server, which keeps audit records per person but no data per course phase.
func Module(environment, coreHost string) standaloneModule.Module {
	baseURL := resolution.NormaliseHost(coreHost) + "/ai/api"
	if environment == "development" {
		baseURL = "http://localhost:8092/ai/api"
	}
	return standaloneModule.Module{Name: ServiceName, BaseURL: baseURL}
}
