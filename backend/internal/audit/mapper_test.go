package audit

import "testing"

func TestMapMetadataRedactsSensitiveKeysRecursively(t *testing.T) {
	metadata, err := mapMetadata([]byte(`{
		"reason":"password reset",
		"password":"one",
		"nested":{"client-secret":"two","safe":true},
		"items":[{"refreshToken":"three"},{"count":2}],
		"authorization_version":7,
		"refresh_token_hash":"four"
	}`))
	if err != nil {
		t.Fatalf("mapMetadata() error = %v", err)
	}
	if metadata["password"] != redactedValue {
		t.Fatalf("password = %v", metadata["password"])
	}
	nested := metadata["nested"].(map[string]any)
	if nested["client-secret"] != redactedValue || nested["safe"] != true {
		t.Fatalf("nested metadata = %+v", nested)
	}
	items := metadata["items"].([]any)
	if items[0].(map[string]any)["refreshToken"] != redactedValue {
		t.Fatalf("array metadata = %+v", items)
	}
	if metadata["reason"] != "password reset" {
		t.Fatalf("non-sensitive value was changed: %v", metadata["reason"])
	}
	if metadata["authorization_version"] != float64(7) {
		t.Fatalf("authorization version was redacted: %v", metadata["authorization_version"])
	}
	if metadata["refresh_token_hash"] != redactedValue {
		t.Fatalf("refresh token hash was exposed: %v", metadata["refresh_token_hash"])
	}
}

func TestMapMetadataRejectsNonObject(t *testing.T) {
	if _, err := mapMetadata([]byte(`["not-an-object"]`)); err == nil {
		t.Fatal("expected non-object metadata to fail closed")
	}
}
