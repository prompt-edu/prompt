package providerconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig/providerconfigDTO"
	log "github.com/sirupsen/logrus"
)

// validateUpsertRequest ensures the request describes a known provider and includes
// every required credential field for it.
func validateUpsertRequest(req providerconfigDTO.UpsertRequest) error {
	if strings.TrimSpace(req.ProviderType) == "" {
		return logAndReturnError("providerType is required")
	}
	fields, err := GetAuthFields(req.ProviderType)
	if err != nil {
		return err
	}
	return validateRequiredCredentials(fields, req.Credentials)
}

// validateProviderType returns a wrapped "unknown provider" error when the type
// has no registered auth fields.
func validateProviderType(providerType string) error {
	if strings.TrimSpace(providerType) == "" {
		return logAndReturnError("providerType is required")
	}
	if _, err := GetAuthFields(providerType); err != nil {
		return err
	}
	return nil
}

// validateRequiredCredentials checks that every required field is present and that all
// supplied values are strings. Providers unmarshal the stored JSON into string fields,
// so a nested object or number would be encrypted happily and only blow up much later,
// when the worker builds the provider.
func validateRequiredCredentials(fields []provider.AuthField, creds map[string]interface{}) error {
	known := make(map[string]provider.AuthField, len(fields))
	for _, field := range fields {
		known[field.Name] = field
	}

	for name, raw := range creds {
		if _, ok := known[name]; !ok {
			return logAndReturnError("unknown credential field: " + name)
		}
		if _, isString := raw.(string); !isString {
			return logAndReturnError("credential " + name + " must be a string")
		}
	}

	for _, field := range fields {
		if !field.Required {
			continue
		}
		value, ok := creds[field.Name].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return logAndReturnError("missing required credential: " + field.Name)
		}
	}
	return nil
}

// ErrValidation marks a rejected request so the router can answer 400 instead of 500.
var ErrValidation = errors.New("invalid provider configuration")

func logAndReturnError(msg string) error {
	log.Error(msg)
	return fmt.Errorf("%w: %s", ErrValidation, msg)
}
