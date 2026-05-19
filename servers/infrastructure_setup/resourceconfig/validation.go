package resourceconfig

import (
	"errors"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig/resourceconfigDTO"
	log "github.com/sirupsen/logrus"
)

const maxNameTemplateLength = 255

// validateCreateResourceConfig checks the request beyond the Gin binding tags.
func validateCreateResourceConfig(req resourceconfigDTO.CreateRequest) error {
	if strings.TrimSpace(req.ProviderType) == "" {
		return logAndReturnError("providerType is required")
	}
	if strings.TrimSpace(req.ResourceType) == "" {
		return logAndReturnError("resourceType is required")
	}
	if err := validateScope(req.Scope); err != nil {
		return err
	}
	return validateNameTemplate(req.NameTemplate)
}

// validateUpdateResourceConfig checks the update request beyond the Gin binding tags.
func validateUpdateResourceConfig(req resourceconfigDTO.UpdateRequest) error {
	if strings.TrimSpace(req.ResourceType) == "" {
		return logAndReturnError("resourceType is required")
	}
	if err := validateScope(req.Scope); err != nil {
		return err
	}
	return validateNameTemplate(req.NameTemplate)
}

func validateScope(scope string) error {
	switch scope {
	case "per_team", "per_student":
		return nil
	}
	return logAndReturnError("scope must be per_team or per_student")
}

// validateNameTemplate ensures the template is non-empty, within size limits,
// and has balanced {{ }} placeholder delimiters. The execution worker renders
// names via plain string replacement (see execution.ResolveName), so we only
// need to catch malformed delimiter pairs here.
func validateNameTemplate(nameTemplate string) error {
	trimmed := strings.TrimSpace(nameTemplate)
	if trimmed == "" {
		return logAndReturnError("nameTemplate is required")
	}
	if len(nameTemplate) > maxNameTemplateLength {
		return logAndReturnError("nameTemplate is too long")
	}
	if strings.Count(nameTemplate, "{{") != strings.Count(nameTemplate, "}}") {
		return logAndReturnError("invalid nameTemplate: unbalanced {{ }} delimiters")
	}
	return nil
}

func logAndReturnError(msg string) error {
	log.Error(msg)
	return errors.New(msg)
}
