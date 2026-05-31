package outline

import "testing"

func TestNewDefaultsBaseURL(t *testing.T) {
	provider := New(Config{APIKey: "secret"})
	if provider.cfg.BaseURL != "https://app.getoutline.com/api" {
		t.Fatalf("BaseURL = %q, want Outline cloud API default", provider.cfg.BaseURL)
	}
}

func TestOutlineAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 2 {
		t.Fatalf("auth fields = %d, want 2", len(fields))
	}
	if fields[0].Name != "api_key" || fields[0].Type != "password" || !fields[0].Required {
		t.Fatalf("first auth field = %+v, want required api_key password", fields[0])
	}
}
