package resourceconfig

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/execution"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig/resourceconfigDTO"
	log "github.com/sirupsen/logrus"
)

const maxNameTemplateLength = 255

// ErrValidation marks a rejected request so the router can answer 400 instead of 500.
var ErrValidation = errors.New("invalid resource configuration")

// ErrProviderNotConfigured is returned when the provider has no stored credentials.
var ErrProviderNotConfigured = errors.New("provider credentials are missing")

// validateResourceType rejects a resource kind the provider cannot create.
func validateResourceType(providerType, resourceType string) error {
	supported, err := providerconfig.SupportedResourceTypes(providerType)
	if err != nil {
		return logAndReturnError(err.Error())
	}
	if !slices.Contains(supported, resourceType) {
		return logAndReturnError(fmt.Sprintf(
			"resourceType %q is not supported by %s (supported: %s)",
			resourceType, providerType, strings.Join(supported, ", ")))
	}
	return nil
}

// validateCreateResourceConfig checks the request beyond the Gin binding tags.
func validateCreateResourceConfig(req resourceconfigDTO.CreateRequest) error {
	if strings.TrimSpace(req.ProviderType) == "" {
		return logAndReturnError("providerType is required")
	}
	if strings.TrimSpace(req.ResourceType) == "" {
		return logAndReturnError("resourceType is required")
	}
	if err := validateResourceType(req.ProviderType, req.ResourceType); err != nil {
		return err
	}
	if err := validateScope(req.Scope); err != nil {
		return err
	}
	return validateNameTemplate(req.NameTemplate)
}

// validateUpdateResourceConfig checks the update request beyond the Gin binding tags.
// providerType comes from the stored row: an update cannot change it.
func validateUpdateResourceConfig(providerType string, req resourceconfigDTO.UpdateRequest) error {
	if strings.TrimSpace(req.ResourceType) == "" {
		return logAndReturnError("resourceType is required")
	}
	if err := validateResourceType(providerType, req.ResourceType); err != nil {
		return err
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

// validateNameTemplate ensures the template is non-empty, within size limits, has
// balanced {{ }} delimiters, and uses only placeholders the worker can resolve.
// An unknown placeholder would otherwise only surface as a failed instance once
// provisioning runs.
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
	if _, err := execution.ResolveName(nameTemplate, execution.TemplateData{}); err != nil {
		return logAndReturnError(fmt.Sprintf("%v (supported: %s)",
			err, strings.Join(execution.SupportedPlaceholders(), ", ")))
	}
	return nil
}

func logAndReturnError(msg string) error {
	log.Error(msg)
	return fmt.Errorf("%w: %s", ErrValidation, msg)
}
