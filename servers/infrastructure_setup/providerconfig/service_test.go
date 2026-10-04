package providerconfig

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/encryption"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig/providerconfigDTO"
)

func setupProviderConfigTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

func setProviderConfigEncryptionKey(t *testing.T) {
	t.Helper()
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
}

func TestUpsertProviderConfigEncryptsAndRedactsCredentials(t *testing.T) {
	setProviderConfigEncryptionKey(t)
	testDB, cleanup := setupProviderConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	service := NewService(testDB.Conn)

	resp, err := service.UpsertProviderConfig(context.Background(), coursePhaseID, providerconfigDTO.UpsertRequest{
		ProviderType: "gitlab",
		Credentials: map[string]interface{}{
			"base_url":      "https://gitlab.example.com",
			"private_token": "secret-token",
		},
	})
	if err != nil {
		t.Fatalf("UpsertProviderConfig returned error: %v", err)
	}
	if resp.ID == uuid.Nil || resp.ProviderType != "gitlab" {
		t.Fatalf("response = %+v, want redacted gitlab provider config", resp)
	}

	stored, err := testDB.Queries.GetProviderConfig(context.Background(), db.GetProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
	})
	if err != nil {
		t.Fatalf("GetProviderConfig returned error: %v", err)
	}
	if string(stored.Credentials) == "secret-token" {
		t.Fatal("credentials were stored in plaintext")
	}

	raw, err := encryption.Decrypt(stored.Credentials)
	if err != nil {
		t.Fatalf("decrypt stored credentials: %v", err)
	}
	var decrypted map[string]string
	if err := json.Unmarshal(raw, &decrypted); err != nil {
		t.Fatalf("decode decrypted credentials: %v", err)
	}
	if decrypted["private_token"] != "secret-token" {
		t.Fatalf("private_token = %q, want secret-token", decrypted["private_token"])
	}
}

func TestListAndDeleteProviderConfigs(t *testing.T) {
	setProviderConfigEncryptionKey(t)
	testDB, cleanup := setupProviderConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	service := NewService(testDB.Conn)

	if _, err := service.UpsertProviderConfig(context.Background(), coursePhaseID, providerconfigDTO.UpsertRequest{ProviderType: "gitlab", Credentials: map[string]interface{}{"base_url": "https://gitlab.example.com", "private_token": "secret"}}); err != nil {
		t.Fatalf("upsert gitlab: %v", err)
	}
	if _, err := service.UpsertProviderConfig(context.Background(), coursePhaseID, providerconfigDTO.UpsertRequest{ProviderType: "slack", Credentials: map[string]interface{}{"bot_token": "secret"}}); err != nil {
		t.Fatalf("upsert slack: %v", err)
	}

	configs, err := service.ListProviderConfigs(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("ListProviderConfigs returned error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("configs = %d, want 2", len(configs))
	}

	if err := service.DeleteProviderConfig(context.Background(), coursePhaseID, "gitlab"); err != nil {
		t.Fatalf("DeleteProviderConfig returned error: %v", err)
	}
	configs, err = service.ListProviderConfigs(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("ListProviderConfigs after delete returned error: %v", err)
	}
	if len(configs) != 1 || configs[0].ProviderType != "slack" {
		t.Fatalf("configs after delete = %+v, want only slack", configs)
	}
}

func TestGetAuthFieldsRejectsUnknownProvider(t *testing.T) {
	if _, err := GetAuthFields("unknown"); err == nil {
		t.Fatal("GetAuthFields returned nil error for unknown provider")
	}
}

// Credentials are stored as JSON and unmarshalled into string fields by the providers,
// so a non-string value has to be rejected up front rather than blowing up in the
// worker much later.
func TestValidateUpsertRequestRejectsNonStringCredentials(t *testing.T) {
	tests := map[string]map[string]interface{}{
		"nested object": {"base_url": map[string]interface{}{"nested": true}, "private_token": "t"},
		"array":         {"base_url": []interface{}{"a"}, "private_token": "t"},
		"number":        {"base_url": "https://gitlab.test", "private_token": 42},
	}

	for name, credentials := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateUpsertRequest(providerconfigDTO.UpsertRequest{
				ProviderType: "gitlab",
				Credentials:  credentials,
			})
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestValidateUpsertRequestRejectsUnknownField(t *testing.T) {
	err := validateUpsertRequest(providerconfigDTO.UpsertRequest{
		ProviderType: "gitlab",
		Credentials: map[string]interface{}{
			"base_url":      "https://gitlab.test",
			"private_token": "token",
			"typo_field":    "value",
		},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation for an unknown field", err)
	}
}

func TestValidateUpsertRequestAcceptsValidCredentials(t *testing.T) {
	if err := validateUpsertRequest(providerconfigDTO.UpsertRequest{
		ProviderType: "gitlab",
		Credentials: map[string]interface{}{
			"base_url":      "https://gitlab.test",
			"private_token": "token",
		},
	}); err != nil {
		t.Fatalf("validateUpsertRequest: %v", err)
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	for providerType, want := range map[string][]string{
		"gitlab":   {"group", "project"},
		"slack":    {"channel"},
		"outline":  {"collection"},
		"rancher":  {"project"},
		"keycloak": {"group"},
	} {
		got, err := SupportedResourceTypes(providerType)
		if err != nil {
			t.Fatalf("SupportedResourceTypes(%q): %v", providerType, err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("SupportedResourceTypes(%q) = %v, want %v", providerType, got, want)
		}
	}

	if _, err := SupportedResourceTypes("unknown"); err == nil {
		t.Fatal("SupportedResourceTypes returned no error for an unknown provider")
	}
}

// Only the keys a provider actually treats as templates may be resolved; anything else
// would change the meaning of a literal provider setting.
func TestTemplatedExtraConfigKeys(t *testing.T) {
	for providerType, want := range map[string][]string{
		"gitlab":   {"parent_group_template"},
		"outline":  {"group_name_template"},
		"slack":    nil,
		"rancher":  nil,
		"keycloak": nil,
	} {
		got, err := TemplatedExtraConfigKeys(providerType)
		if err != nil {
			t.Fatalf("TemplatedExtraConfigKeys(%q): %v", providerType, err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("TemplatedExtraConfigKeys(%q) = %v, want %v", providerType, got, want)
		}
	}
}
