package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// captureServer records the last request and answers with data.
func captureServer(t *testing.T, data interface{}) (*httptest.Server, *map[string]interface{}, *http.Request) {
	t.Helper()
	body := map[string]interface{}{}
	last := &http.Request{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = map[string]interface{}{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
		*last = *r.Clone(context.Background())
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": 200, "data": data})
	}))
	t.Cleanup(server.Close)
	return server, &body, last
}

func TestFederationProvider_CatalogSocialPayloadHasNoEndpoints(t *testing.T) {
	t.Parallel()
	server, body, last := captureServer(t, map[string]interface{}{
		"id": "fed-apple", "provider_name": "default_apple", "provider": "apple", "standard_type": "OAUTH2",
		"oauth2": map[string]interface{}{"client_id": "apple-client"},
	})

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	res, err := client.Create(context.Background(), &FederationProviderModel{
		ProviderName: "default_apple",
		Provider:     "apple",
		DisplayName:  "Apple",
		StandardType: FederationStandardTypeOAuth2,
		OAuth2:       &FederationOAuth2Config{ClientID: "apple-client", ClientSecret: "apple-secret"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if res.Data.ID != "fed-apple" {
		t.Fatalf("unexpected id: %q", res.Data.ID)
	}
	if last.Method != http.MethodPost || last.URL.Path != "/federation/providers" {
		t.Fatalf("unexpected request %s %s", last.Method, last.URL.Path)
	}
	b := *body
	if _, ok := b["id"]; ok {
		t.Fatalf("create must not send a client-supplied id: %v", b)
	}
	if b["provider"] != "apple" || b["owner"] != "client" {
		t.Fatalf("unexpected base payload: %v", b)
	}
	oauth2, ok := b["oauth2"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected oauth2 nest, got %v", b)
	}
	for _, key := range []string{"authorization_endpoint", "token_endpoint", "userinfo_endpoint", "userinfoSource"} {
		if _, present := oauth2[key]; present {
			t.Fatalf("catalog social payload must not contain %s: %v", key, oauth2)
		}
	}
	if oauth2["client_id"] != "apple-client" || oauth2["client_secret"] != "apple-secret" {
		t.Fatalf("unexpected oauth2 credentials: %v", oauth2)
	}
}

func TestFederationProvider_CustomPayloadKeepsEndpointsAndUserinfoSource(t *testing.T) {
	t.Parallel()
	server, body, last := captureServer(t, map[string]interface{}{
		"_id": "fed-custom", "provider_name": "my_oidc", "provider": "custom", "standard_type": "OPENID_CONNECT",
		"createdTime": "2026-01-01T00:00:00Z",
	})

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	model := &FederationProviderModel{
		ID:           "ignored",
		ProviderName: "my_oidc",
		Provider:     "custom",
		StandardType: FederationStandardTypeOIDC,
		CreatedTime:  "2026-01-01T00:00:00Z",
		OAuth2: &FederationOAuth2Config{
			ClientID:          "cid",
			ClientSecret:      "secret",
			WellknownEndpoint: "https://idp.example.com/.well-known/openid-configuration",
			UserinfoSource:    "IDTOKEN",
			TokenEndpointAuth: json.RawMessage(`{"apiAccessType":"BASIC"}`),
		},
		// must be dropped: does not match standard_type
		SAML: &FederationSAMLConfig{MetaData: "<xml/>"},
	}
	res, err := client.Update(context.Background(), "fed-custom", model)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.Data.ID != "fed-custom" {
		t.Fatalf("expected id from _id fallback, got %q", res.Data.ID)
	}
	if last.Method != http.MethodPut || last.URL.Path != "/federation/providers/fed-custom" {
		t.Fatalf("unexpected request %s %s", last.Method, last.URL.Path)
	}
	b := *body
	if b["id"] != "fed-custom" {
		t.Fatalf("update must send the path id, got %v", b["id"])
	}
	if _, ok := b["createdTime"]; ok {
		t.Fatalf("read-only createdTime must not be sent: %v", b)
	}
	if _, ok := b["saml"]; ok {
		t.Fatalf("saml must not be sent for OPENID_CONNECT: %v", b)
	}
	oauth2 := b["oauth2"].(map[string]interface{})
	if oauth2["userinfoSource"] != "IDTOKEN" || oauth2["wellknown_endpoint"] == nil {
		t.Fatalf("unexpected oauth2 payload: %v", oauth2)
	}
	if auth, ok := oauth2["token_endpoint_auth"].(map[string]interface{}); !ok || auth["apiAccessType"] != "BASIC" {
		t.Fatalf("token_endpoint_auth must be sent as an object: %v", oauth2["token_endpoint_auth"])
	}
	if model.ID != "ignored" || model.SAML == nil {
		t.Fatal("payload builder must not mutate the caller's model")
	}
}

func TestFederationProvider_SAMLPayloadDropsSPCertificates(t *testing.T) {
	t.Parallel()
	server, body, _ := captureServer(t, map[string]interface{}{
		"id": "fed-saml", "provider_name": "corp_saml", "provider": "custom", "standard_type": "SAML",
		"saml": map[string]interface{}{
			"meta_data":               "<xml/>",
			"sp_signing_certificates": []interface{}{map[string]interface{}{"certificate": "PEM", "sha256_fingerprint": "AA:BB"}},
		},
	})

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	res, err := client.Create(context.Background(), &FederationProviderModel{
		ProviderName: "corp_saml",
		Provider:     "custom",
		DisplayName:  "Corp",
		StandardType: FederationStandardTypeSAML,
		SAML: &FederationSAMLConfig{
			MetaData:              "<xml/>",
			SPSigningCertificates: []FederationSAMLCertificate{{Certificate: "PEM"}},
		},
		OAuth2: &FederationOAuth2Config{ClientID: "dropped"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if len(res.Data.SAML.SPSigningCertificates) != 1 || res.Data.SAML.SPSigningCertificates[0].Sha256Fingerprint != "AA:BB" {
		t.Fatalf("SP certificates not decoded: %+v", res.Data.SAML)
	}
	b := *body
	if _, ok := b["oauth2"]; ok {
		t.Fatalf("oauth2 must not be sent for SAML: %v", b)
	}
	saml := b["saml"].(map[string]interface{})
	if _, ok := saml["sp_signing_certificates"]; ok {
		t.Fatalf("read-only SP certificates must not be sent: %v", saml)
	}
	if saml["meta_data"] != "<xml/>" || saml["sign_auth_request"] != false {
		t.Fatalf("unexpected saml payload: %v", saml)
	}
}

func TestFederationProvider_LDAPPayload(t *testing.T) {
	t.Parallel()
	server, body, _ := captureServer(t, map[string]interface{}{"id": "fed-ldap"})

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	useSSL := true
	if _, err := client.Create(context.Background(), &FederationProviderModel{
		ProviderName: "corp_ldap",
		Provider:     "custom",
		StandardType: FederationStandardTypeLDAP,
		LDAP: &FederationLDAPConfig{
			ConnectionString: "ldaps://ldap.example.com",
			BindDN:           "cn=admin",
			BindCredentials:  "pw",
			BaseDN:           "dc=example,dc=com",
			UseSSL:           &useSSL,
			DefaultGroups:    []FederationLDAPGroup{{GroupID: "CIDAAS_USERS", Roles: []string{"USER"}}},
		},
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	ldap := (*body)["ldap"].(map[string]interface{})
	if ldap["bind_credentials"] != "pw" || ldap["use_ssl"] != true {
		t.Fatalf("unexpected ldap payload: %v", ldap)
	}
	groups := ldap["default_groups"].([]interface{})
	if groups[0].(map[string]interface{})["groupId"] != "CIDAAS_USERS" {
		t.Fatalf("unexpected default_groups: %v", groups)
	}
}

func TestFederationProvider_ListAndDelete(t *testing.T) {
	t.Parallel()
	server, _, last := captureServer(t, []interface{}{
		map[string]interface{}{"_id": "a", "provider_name": "my_google"},
	})

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	list, err := client.List(context.Background(), FederationProviderListParams{ProviderName: "my_google", StandardType: "OAUTH2", Size: 100})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "a" {
		t.Fatalf("unexpected list: %+v", list)
	}
	q := last.URL.Query()
	if last.URL.Path != "/federation/providers" || q.Get("provider_name") != "my_google" || q.Get("standard_type") != "OAUTH2" || q.Get("size") != "100" || q.Has("from") {
		t.Fatalf("unexpected list request: %s", last.URL.String())
	}

	if err := client.Delete(context.Background(), "a"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if last.Method != http.MethodDelete || last.URL.Path != "/federation/providers/a" {
		t.Fatalf("unexpected delete request %s %s", last.Method, last.URL.Path)
	}
}

func TestIsFederationSocialProvider(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"apple", "google", "netid_qa", "x"} {
		if !IsFederationSocialProvider(p) {
			t.Errorf("%s should be a catalog social provider", p)
		}
	}
	for _, p := range []string{"custom", "", "keycloak"} {
		if IsFederationSocialProvider(p) {
			t.Errorf("%s should not be a catalog social provider", p)
		}
	}
}
