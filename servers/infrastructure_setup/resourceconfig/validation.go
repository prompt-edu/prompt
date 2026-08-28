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
	if err := validateNameTemplate(req.NameTemplate); err != nil {
		return err
	}
	return validateTemplatedExtraConfig(req.ProviderType, req.ResourceExtraConfig)
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
	if err := validateNameTemplate(req.NameTemplate); err != nil {
		return err
	}
	return validateTemplatedExtraConfig(providerType, req.ResourceExtraConfig)
}

// validateTemplatedExtraConfig applies the name-template rules to the extra-config keys
// the provider declares as templates, so an unresolvable placeholder is rejected here
// rather than reaching a real resource name during a run.
func validateTemplatedExtraConfig(providerType string, extra map[string]interface{}) error {
	if len(extra) == 0 {
		return nil
	}
	keys, err := providerconfig.TemplatedExtraConfigKeys(providerType)
	if err != nil {
		return logAndReturnError(err.Error())
	}
	for _, key := range keys {
		raw, ok := extra[key]
		if !ok {
			continue
		}
		text, isString := raw.(string)
		if !isString {
			return logAndReturnError(fmt.Sprintf("extra config %q must be a string", key))
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		if err := validateTemplate(key, text); err != nil {
			return err
		}
	}
	return nil
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
	if strings.TrimSpace(nameTemplate) == "" {
		return logAndReturnError("nameTemplate is required")
	}
	return validateTemplate("nameTemplate", nameTemplate)
}

func validateTemplate(field, template string) error {
	if len(template) > maxNameTemplateLength {
		return logAndReturnError(field + " is too long")
	}
	if strings.Count(template, "{{") != strings.Count(template, "}}") {
		return logAndReturnError("invalid " + field + ": unbalanced {{ }} delimiters")
	}
	if _, err := execution.ResolveName(template, execution.TemplateData{}); err != nil {
		return logAndReturnError(fmt.Sprintf("invalid %s: %v (supported: %s)",
			field, err, strings.Join(execution.SupportedPlaceholders(), ", ")))
	}
	return nil
}

// logAndReturnError records a rejected request. These are ordinary bad requests answered
// with 400, so they are logged at warn level and do not raise error-level alerts.
func logAndReturnError(msg string) error {
	log.Warn(msg)
	return fmt.Errorf("%w: %s", ErrValidation, msg)
}
