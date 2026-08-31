// helpers/cidaas/password_policy_test.go
package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewPasswordPolicy(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	passwordPolicy := NewPasswordPolicy(config)

	if passwordPolicy == nil {
		t.Fatal("Expected password policy instance, got nil")
	}

	if passwordPolicy.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, passwordPolicy.BaseURL)
	}

	if passwordPolicy.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, passwordPolicy.AccessToken)
	}
}

func TestPasswordPolicy_Get_Success(t *testing.T) {
	expectedPolicy := PasswordPolicyModel{
		ID:         "policy-123",
		PolicyName: "Strong Password Policy",
		PasswordPolicy: &Policy{
			BlockCompromised: true,
			DenyUsageCount:   5,
			StrengthRegexes:  []string{"^(?=.*[a-z])", "^(?=.*[A-Z])", "^(?=.*\\d)"},
			ChangeEnforcement: ChangeEnforcement{
				ExpirationInDays:       90,
				NotifyUserBeforeInDays: 7,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "verification-actions-srv/policies/policy-123") {
			t.Errorf("Expected verification-actions-srv/policies/policy-123 endpoint, got %s", r.URL.Path)
		}

		response := PasswordPolicyResponse{
			Success: true,
			Status:  200,
			Data:    expectedPolicy,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	result, err := passwordPolicy.Get(context.Background(), "policy-123")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ID != expectedPolicy.ID {
		t.Errorf("Expected ID %s, got %s", expectedPolicy.ID, result.Data.ID)
	}

	if result.Data.PolicyName != expectedPolicy.PolicyName {
		t.Errorf("Expected PolicyName %s, got %s", expectedPolicy.PolicyName, result.Data.PolicyName)
	}

	if result.Data.PasswordPolicy.BlockCompromised != expectedPolicy.PasswordPolicy.BlockCompromised {
		t.Errorf("Expected BlockCompromised %t, got %t", expectedPolicy.PasswordPolicy.BlockCompromised, result.Data.PasswordPolicy.BlockCompromised)
	}
}

func TestPasswordPolicy_Get_Error(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "policy not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	_, err := passwordPolicy.Get(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error for not found policy, got nil")
	}
}

func TestPasswordPolicy_Create_Success(t *testing.T) {
	policyPayload := PasswordPolicyModel{
		PolicyName: "New Password Policy",
		PasswordPolicy: &Policy{
			BlockCompromised: false,
			DenyUsageCount:   3,
			StrengthRegexes:  []string{"^(?=.*[a-z])", "^(?=.*[A-Z])"},
			ChangeEnforcement: ChangeEnforcement{
				ExpirationInDays:       60,
				NotifyUserBeforeInDays: 5,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "verification-actions-srv/policies") {
			t.Errorf("Expected verification-actions-srv/policies endpoint, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var receivedPolicy PasswordPolicyModel
		json.Unmarshal(body, &receivedPolicy)

		if receivedPolicy.PolicyName != policyPayload.PolicyName {
			t.Errorf("Expected PolicyName %s, got %s", policyPayload.PolicyName, receivedPolicy.PolicyName)
		}

		if receivedPolicy.PasswordPolicy.DenyUsageCount != policyPayload.PasswordPolicy.DenyUsageCount {
			t.Errorf("Expected DenyUsageCount %d, got %d", policyPayload.PasswordPolicy.DenyUsageCount, receivedPolicy.PasswordPolicy.DenyUsageCount)
		}

		response := PasswordPolicyResponse{
			Success: true,
			Status:  201,
			Data:    policyPayload,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	result, err := passwordPolicy.Create(context.Background(), policyPayload)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.PolicyName != policyPayload.PolicyName {
		t.Errorf("Expected PolicyName %s, got %s", policyPayload.PolicyName, result.Data.PolicyName)
	}
}

func TestPasswordPolicy_Create_Error(t *testing.T) {
	server := NewMockServer(http.StatusBadRequest, `{"error": "invalid policy"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	policyPayload := PasswordPolicyModel{PolicyName: "Invalid Policy"}
	_, err := passwordPolicy.Create(context.Background(), policyPayload)

	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestPasswordPolicy_Update_Success(t *testing.T) {
	policyPayload := PasswordPolicyModel{
		ID:         "policy-123",
		PolicyName: "Updated Password Policy",
		PasswordPolicy: &Policy{
			BlockCompromised: true,
			DenyUsageCount:   10,
			StrengthRegexes:  []string{"^(?=.*[a-z])", "^(?=.*[A-Z])", "^(?=.*\\d)", "^(?=.*[@$!%*?&])"},
			ChangeEnforcement: ChangeEnforcement{
				ExpirationInDays:       120,
				NotifyUserBeforeInDays: 14,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "verification-actions-srv/policies") {
			t.Errorf("Expected verification-actions-srv/policies endpoint, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var receivedPolicy PasswordPolicyModel
		json.Unmarshal(body, &receivedPolicy)

		if receivedPolicy.ID != policyPayload.ID {
			t.Errorf("Expected ID %s, got %s", policyPayload.ID, receivedPolicy.ID)
		}

		if len(receivedPolicy.PasswordPolicy.StrengthRegexes) != 4 {
			t.Errorf("Expected 4 strength regexes, got %d", len(receivedPolicy.PasswordPolicy.StrengthRegexes))
		}

		// Note: PasswordPolicyUpdateResponse.Data is bool, not PasswordPolicyModel
		response := PasswordPolicyUpdateResponse{
			Success: true,
			Status:  200,
			Data:    true, // This is bool, not the policy model
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	result, err := passwordPolicy.Update(context.Background(), policyPayload)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if !result.Data {
		t.Error("Expected Data to be true")
	}
}

func TestPasswordPolicy_Update_Error(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	policyPayload := PasswordPolicyModel{ID: "policy-123"}
	_, err := passwordPolicy.Update(context.Background(), policyPayload)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestPasswordPolicy_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "verification-actions-srv/policies/policy-123") {
			t.Errorf("Expected verification-actions-srv/policies/policy-123 endpoint, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	err := passwordPolicy.Delete(context.Background(), "policy-123")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestPasswordPolicy_Delete_Error(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "policy not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	passwordPolicy := NewPasswordPolicy(config)

	err := passwordPolicy.Delete(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error for not found policy, got nil")
	}
}
