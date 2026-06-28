package providerconfig

import (
	"errors"
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

func validateRequiredCredentials(fields []provider.AuthField, creds map[string]interface{}) error {
	for _, field := range fields {
		if !field.Required {
			continue
		}
		raw, ok := creds[field.Name]
		if !ok {
			return logAndReturnError("missing required credential: " + field.Name)
		}
		if value, isString := raw.(string); isString && strings.TrimSpace(value) == "" {
			return logAndReturnError("missing required credential: " + field.Name)
		}
	}
	return nil
}

func logAndReturnError(msg string) error {
	log.Error(msg)
	return errors.New(msg)
}
