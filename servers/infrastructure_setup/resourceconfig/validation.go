package resourceconfig

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
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

// ErrConfirmationRequired marks a destructive request the caller has to confirm, so the
// router can answer 409 rather than carrying it out.
var ErrConfirmationRequired = errors.New("confirmation required")

// resourceConfigIdentityConstraint is the unique constraint on a config's identity
// within a phase (migration 0003).
const resourceConfigIdentityConstraint = "uq_resource_config_identity"

// uniqueViolationCode is Postgres' SQLSTATE for a unique constraint violation.
const uniqueViolationCode = "23505"

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
	scope := db.ResourceScope(req.Scope)
	if err := validateNameTemplate(req.NameTemplate, scope); err != nil {
		return err
	}
	if err := validateRequiredExtraConfig(req.ProviderType, req.ResourceType, req.ResourceExtraConfig); err != nil {
		return err
	}
	return validateTemplatedExtraConfig(req.ProviderType, req.ResourceExtraConfig, scope)
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
	scope := db.ResourceScope(req.Scope)
	if err := validateNameTemplate(req.NameTemplate, scope); err != nil {
		return err
	}
	if err := validateRequiredExtraConfig(providerType, req.ResourceType, req.ResourceExtraConfig); err != nil {
		return err
	}
	return validateTemplatedExtraConfig(providerType, req.ResourceExtraConfig, scope)
}

// validateTemplatedExtraConfig applies the name-template rules to the extra-config keys
// the provider declares as templates, so an unresolvable placeholder is rejected here
// rather than reaching a real resource name during a run.
func validateTemplatedExtraConfig(providerType string, extra map[string]interface{}, scope db.ResourceScope) error {
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
		if err := validateTemplate(key, text, scope); err != nil {
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
func validateNameTemplate(nameTemplate string, scope db.ResourceScope) error {
	if strings.TrimSpace(nameTemplate) == "" {
		return logAndReturnError("nameTemplate is required")
	}
	return validateTemplate("nameTemplate", nameTemplate, scope)
}

func validateTemplate(field, template string, scope db.ResourceScope) error {
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
	// A placeholder the scope never fills resolves to an empty string, so every target
	// of this config would end up with the same name and share one external resource.
	if unfillable := execution.UnfillablePlaceholders(template, scope); len(unfillable) > 0 {
		return logAndReturnError(fmt.Sprintf("invalid %s: %s cannot be resolved for scope %s (available: %s)",
			field, strings.Join(unfillable, ", "), scope,
			strings.Join(execution.PlaceholdersForScope(scope), ", ")))
	}
	return nil
}

// validateRequiredExtraConfig rejects a config that omits an extra-config key the
// provider needs for the resource type, which would otherwise fail once per instance
// in the middle of a run instead of when the lecturer saves it.
func validateRequiredExtraConfig(providerType, resourceType string, extra map[string]interface{}) error {
	required, err := providerconfig.RequiredExtraConfigKeys(providerType, resourceType)
	if err != nil {
		return logAndReturnError(err.Error())
	}
	for _, key := range required {
		text, isString := extra[key].(string)
		if !isString || strings.TrimSpace(text) == "" {
			return logAndReturnError(fmt.Sprintf("extra config %q is required for %s %s",
				key, providerType, resourceType))
		}
	}
	return nil
}

// logAndReturnError records a rejected request. These are ordinary bad requests answered
// with 400, so they are logged at warn level and do not raise error-level alerts.
func logAndReturnError(msg string) error {
	log.Warn(msg)
	return fmt.Errorf("%w: %s", ErrValidation, msg)
}
