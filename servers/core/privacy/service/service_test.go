package service

import (
	"testing"

	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/standaloneModule"
	"github.com/stretchr/testify/assert"
)

func TestExternalModules(t *testing.T) {
	aiModule := standaloneModule.Module{Name: "AI", BaseURL: "https://prompt.example.org/ai/api"}
	service := NewPrivacyService(db.Queries{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, aiModule)

	modules := service.externalModules([]coursePhaseTypeDTO.CoursePhaseType{
		{Name: "Assessment", BaseUrl: "https://prompt.example.org/assessment/api"},
		{Name: "Application", BaseUrl: "core"},
	})

	assert.Equal(t, []standaloneModule.Module{
		{Name: "Assessment", BaseURL: "https://prompt.example.org/assessment/api"},
		aiModule,
	}, modules, "core implemented phase types have no module, standalone modules are always asked")
}
