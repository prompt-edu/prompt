package rancher

import "testing"

func TestRancherAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 4 {
		t.Fatalf("auth fields = %d, want 4", len(fields))
	}
	required := map[string]bool{}
	for _, field := range fields {
		required[field.Name] = field.Required
	}
	for _, name := range []string{"rancher_url", "access_key", "secret_key", "cluster_id"} {
		if !required[name] {
			t.Fatalf("field %q is not marked required in %+v", name, fields)
		}
	}
}

func TestNewStoresConfig(t *testing.T) {
	provider := New(Config{
		RancherURL: "https://rancher.example.com",
		AccessKey:  "access",
		SecretKey:  "secret",
		ClusterID:  "c-abcde",
	})
	if provider.cfg.ClusterID != "c-abcde" {
		t.Fatalf("ClusterID = %q, want c-abcde", provider.cfg.ClusterID)
	}
}
