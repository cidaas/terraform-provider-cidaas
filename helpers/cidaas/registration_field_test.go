// helpers/cidaas/registration_field_test.go
package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRegField(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	regField := NewRegField(config)

	if regField == nil {
		t.Fatal("Expected registration field instance, got nil")
	}

	if regField.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, regField.BaseURL)
	}

	if regField.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, regField.AccessToken)
	}
}

func TestRegField_Upsert_Success(t *testing.T) {
	minLength := int64(3)
	maxLength := int64(50)
	minDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)

	expectedRegField := RegistrationFieldConfig{
		ID:                                       "field-123",
		FieldKey:                                 "email",
		FieldType:                                "EMAIL",
		DataType:                                 "TEXT",
		BaseDataType:                             "STRING",
		Order:                                    1,
		Internal:                                 false,
		ReadOnly:                                 false,
		Claimable:                                true,
		Required:                                 true,
		Unique:                                   true,
		IsSearchable:                             true,
		OverwriteWithNullValueFromSocialProvider: false,
		Enabled:                                  true,
		IsGroup:                                  false,
		IsList:                                   false,
		ParentGroupID:                            "",
		ClassName:                                "form-control",
		ConsentRefs:                              []string{"consent-1", "consent-2"},
		Scopes:                                   []string{"profile", "email"},
		FieldDefinition: &FieldDefinition{
			MinLength:       &minLength,
			MaxLength:       &maxLength,
			MinDate:         &minDate,
			MaxDate:         &maxDate,
			InitialDateView: "month",
			Regex:           "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			AttributesKeys:  []string{"placeholder", "class"},
		},
		LocaleTexts: []*LocaleText{
			{
				Locale:            "en-US",
				Name:              "Email Address",
				MinLengthErrorMsg: "Email must be at least 3 characters",
				MaxLengthErrorMsg: "Email must not exceed 50 characters",
				RequiredMsg:       "Email is required",
				Attributes: []*Attribute{
					{
						Key:   "placeholder",
						Value: "Enter your email",
					},
					{
						Key:   "class",
						Value: "email-input",
					},
				},
				ConsentLabel: &ConsentLabel{
					Label:     "email_consent",
					LabelText: "I agree to receive emails",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "fieldsetup-srv/fields") {
			t.Errorf("Expected fieldsetup-srv/fields endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		if receivedRegField.FieldKey != expectedRegField.FieldKey {
			t.Errorf("Expected FieldKey %s, got %s", expectedRegField.FieldKey, receivedRegField.FieldKey)
		}

		if receivedRegField.Required != expectedRegField.Required {
			t.Errorf("Expected Required %t, got %t", expectedRegField.Required, receivedRegField.Required)
		}

		if len(receivedRegField.Scopes) != len(expectedRegField.Scopes) {
			t.Errorf("Expected %d scopes, got %d", len(expectedRegField.Scopes), len(receivedRegField.Scopes))
		}

		// Verify FieldDefinition
		if receivedRegField.FieldDefinition == nil {
			t.Error("Expected FieldDefinition to be present")
		} else {
			if *receivedRegField.FieldDefinition.MinLength != *expectedRegField.FieldDefinition.MinLength {
				t.Errorf("Expected MinLength %d, got %d", *expectedRegField.FieldDefinition.MinLength, *receivedRegField.FieldDefinition.MinLength)
			}
		}

		// Verify LocaleTexts
		if len(receivedRegField.LocaleTexts) != 1 {
			t.Errorf("Expected 1 locale text, got %d", len(receivedRegField.LocaleTexts))
		}

		// Send success response
		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    expectedRegField,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), expectedRegField)

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.FieldKey != expectedRegField.FieldKey {
		t.Errorf("Expected FieldKey %s, got %s", expectedRegField.FieldKey, result.Data.FieldKey)
	}

	if result.Data.Required != expectedRegField.Required {
		t.Errorf("Expected Required %t, got %t", expectedRegField.Required, result.Data.Required)
	}
}

func TestRegField_Upsert_MinimalData(t *testing.T) {
	minimalRegField := RegistrationFieldConfig{
		FieldKey:  "username",
		FieldType: "TEXT",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		if receivedRegField.FieldKey != minimalRegField.FieldKey {
			t.Errorf("Expected FieldKey %s, got %s", minimalRegField.FieldKey, receivedRegField.FieldKey)
		}

		if receivedRegField.Required != minimalRegField.Required {
			t.Errorf("Expected Required %t, got %t", minimalRegField.Required, receivedRegField.Required)
		}

		if receivedRegField.FieldType != minimalRegField.FieldType {
			t.Errorf("Expected FieldType %s, got %s", minimalRegField.FieldType, receivedRegField.FieldType)
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    minimalRegField,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), minimalRegField)

	if err != nil {
		t.Fatalf("Upsert with minimal data failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.FieldKey != minimalRegField.FieldKey {
		t.Errorf("Expected FieldKey %s, got %s", minimalRegField.FieldKey, result.Data.FieldKey)
	}

	if result.Data.Enabled != minimalRegField.Enabled {
		t.Errorf("Expected Enabled %t, got %t", minimalRegField.Enabled, result.Data.Enabled)
	}
}

func TestRegField_Upsert_WithComplexFieldDefinition(t *testing.T) {
	minLength := int64(8)
	maxLength := int64(128)
	initialDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	regFieldWithComplexDef := RegistrationFieldConfig{
		FieldKey:  "password",
		FieldType: "PASSWORD",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
		FieldDefinition: &FieldDefinition{
			MinLength:       &minLength,
			MaxLength:       &maxLength,
			InitialDate:     &initialDate,
			InitialDateView: "year",
			Regex:           "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d@$!%*?&]{8,}$",
			AttributesKeys:  []string{"autocomplete", "pattern"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		// Verify complex FieldDefinition
		if receivedRegField.FieldDefinition == nil {
			t.Fatal("Expected FieldDefinition to be present")
		}

		if *receivedRegField.FieldDefinition.MinLength != minLength {
			t.Errorf("Expected MinLength %d, got %d", minLength, *receivedRegField.FieldDefinition.MinLength)
		}

		if *receivedRegField.FieldDefinition.MaxLength != maxLength {
			t.Errorf("Expected MaxLength %d, got %d", maxLength, *receivedRegField.FieldDefinition.MaxLength)
		}

		if len(receivedRegField.FieldDefinition.AttributesKeys) != 2 {
			t.Errorf("Expected 2 attributes keys, got %d", len(receivedRegField.FieldDefinition.AttributesKeys))
		}

		if receivedRegField.FieldDefinition.InitialDateView != "year" {
			t.Errorf("Expected InitialDateView 'year', got %s", receivedRegField.FieldDefinition.InitialDateView)
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    regFieldWithComplexDef,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), regFieldWithComplexDef)

	if err != nil {
		t.Fatalf("Upsert with complex field definition failed: %v", err)
	}

	if result.Data.FieldDefinition == nil {
		t.Error("Expected FieldDefinition in response")
	}

	if *result.Data.FieldDefinition.MinLength != minLength {
		t.Errorf("Expected MinLength %d in response, got %d", minLength, *result.Data.FieldDefinition.MinLength)
	}

	if len(result.Data.FieldDefinition.AttributesKeys) != 2 {
		t.Errorf("Expected 2 attributes keys in response, got %d", len(result.Data.FieldDefinition.AttributesKeys))
	}
}

func TestRegField_Upsert_WithMultipleLocaleTexts(t *testing.T) {
	regFieldWithLocales := RegistrationFieldConfig{
		FieldKey:  "full_name",
		FieldType: "TEXT",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
		LocaleTexts: []*LocaleText{
			{
				Locale:            "en-US",
				Name:              "Full Name",
				RequiredMsg:       "Full name is required",
				MinLengthErrorMsg: "Name too short",
				MaxLengthErrorMsg: "Name too long",
				Attributes: []*Attribute{
					{Key: "placeholder", Value: "Enter your full name"},
					{Key: "class", Value: "name-input"},
				},
				ConsentLabel: &ConsentLabel{
					Label:     "name_consent",
					LabelText: "I agree to provide my name",
				},
			},
			{
				Locale:            "de-DE",
				Name:              "Vollständiger Name",
				RequiredMsg:       "Vollständiger Name ist erforderlich",
				MinLengthErrorMsg: "Name zu kurz",
				MaxLengthErrorMsg: "Name zu lang",
				Attributes: []*Attribute{
					{Key: "placeholder", Value: "Geben Sie Ihren vollständigen Namen ein"},
					{Key: "class", Value: "name-input-de"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		if len(receivedRegField.LocaleTexts) != 2 {
			t.Errorf("Expected 2 locale texts, got %d", len(receivedRegField.LocaleTexts))
		}

		// Verify English locale
		enLocale := receivedRegField.LocaleTexts[0]
		if enLocale.Name != "Full Name" {
			t.Errorf("Expected first locale name 'Full Name', got %s", enLocale.Name)
		}

		if len(enLocale.Attributes) != 2 {
			t.Errorf("Expected 2 attributes for English locale, got %d", len(enLocale.Attributes))
		}

		// Verify German locale
		deLocale := receivedRegField.LocaleTexts[1]
		if deLocale.Name != "Vollständiger Name" {
			t.Errorf("Expected second locale name 'Vollständiger Name', got %s", deLocale.Name)
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    regFieldWithLocales,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), regFieldWithLocales)

	if err != nil {
		t.Fatalf("Upsert with multiple locales failed: %v", err)
	}

	if len(result.Data.LocaleTexts) != 2 {
		t.Errorf("Expected 2 locale texts in response, got %d", len(result.Data.LocaleTexts))
	}

	// Verify locale names in response
	if result.Data.LocaleTexts[0].Name != "Full Name" {
		t.Errorf("Expected first locale name 'Full Name' in response, got %s", result.Data.LocaleTexts[0].Name)
	}
	if result.Data.LocaleTexts[1].Name != "Vollständiger Name" {
		t.Errorf("Expected second locale name 'Vollständiger Name' in response, got %s", result.Data.LocaleTexts[1].Name)
	}
}

func TestRegField_Upsert_WithBooleanFlags(t *testing.T) {
	regFieldWithFlags := RegistrationFieldConfig{
		FieldKey:                                 "phone",
		FieldType:                                "PHONE",
		DataType:                                 "TEXT",
		Internal:                                 true,
		ReadOnly:                                 false,
		Claimable:                                true,
		Required:                                 false,
		Unique:                                   true,
		IsSearchable:                             false,
		OverwriteWithNullValueFromSocialProvider: true,
		Enabled:                                  true,
		IsGroup:                                  false,
		IsList:                                   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		// Verify all boolean flags
		if receivedRegField.Internal != true {
			t.Error("Expected Internal to be true")
		}
		if receivedRegField.ReadOnly != false {
			t.Error("Expected ReadOnly to be false")
		}
		if receivedRegField.Claimable != true {
			t.Error("Expected Claimable to be true")
		}
		if receivedRegField.Required != false {
			t.Error("Expected Required to be false")
		}
		if receivedRegField.Unique != true {
			t.Error("Expected Unique to be true")
		}
		if receivedRegField.IsSearchable != false {
			t.Error("Expected IsSearchable to be false")
		}
		if receivedRegField.OverwriteWithNullValueFromSocialProvider != true {
			t.Error("Expected OverwriteWithNullValueFromSocialProvider to be true")
		}
		if receivedRegField.IsList != true {
			t.Error("Expected IsList to be true")
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    regFieldWithFlags,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), regFieldWithFlags)

	if err != nil {
		t.Fatalf("Upsert with boolean flags failed: %v", err)
	}

	// Verify boolean flags in response
	if result.Data.Internal != true {
		t.Error("Expected Internal to be true in response")
	}
	if result.Data.IsList != true {
		t.Error("Expected IsList to be true in response")
	}
	if result.Data.OverwriteWithNullValueFromSocialProvider != true {
		t.Error("Expected OverwriteWithNullValueFromSocialProvider to be true in response")
	}
	if result.Data.Unique != true {
		t.Error("Expected Unique to be true in response")
	}
}

func TestRegField_Upsert_WithConsentRefs(t *testing.T) {
	regFieldWithConsent := RegistrationFieldConfig{
		FieldKey:    "marketing_email",
		FieldType:   "CHECKBOX",
		DataType:    "BOOLEAN",
		Required:    false,
		Enabled:     true,
		ConsentRefs: []string{"marketing_consent", "newsletter_consent", "promotional_consent"},
		Scopes:      []string{"marketing", "communications"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedRegField RegistrationFieldConfig
		json.Unmarshal(body, &receivedRegField)

		if len(receivedRegField.ConsentRefs) != 3 {
			t.Errorf("Expected 3 consent refs, got %d", len(receivedRegField.ConsentRefs))
		}

		expectedConsents := []string{"marketing_consent", "newsletter_consent", "promotional_consent"}
		for i, consent := range receivedRegField.ConsentRefs {
			if consent != expectedConsents[i] {
				t.Errorf("Expected consent ref %s at index %d, got %s", expectedConsents[i], i, consent)
			}
		}

		if len(receivedRegField.Scopes) != 2 {
			t.Errorf("Expected 2 scopes, got %d", len(receivedRegField.Scopes))
		}

		if receivedRegField.FieldType != "CHECKBOX" {
			t.Errorf("Expected FieldType CHECKBOX, got %s", receivedRegField.FieldType)
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    regFieldWithConsent,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Upsert(context.Background(), regFieldWithConsent)

	if err != nil {
		t.Fatalf("Upsert with consent refs failed: %v", err)
	}

	if len(result.Data.ConsentRefs) != 3 {
		t.Errorf("Expected 3 consent refs in response, got %d", len(result.Data.ConsentRefs))
	}

	if len(result.Data.Scopes) != 2 {
		t.Errorf("Expected 2 scopes in response, got %d", len(result.Data.Scopes))
	}

	// Verify specific consent refs
	expectedConsents := []string{"marketing_consent", "newsletter_consent", "promotional_consent"}
	for i, consent := range result.Data.ConsentRefs {
		if consent != expectedConsents[i] {
			t.Errorf("Expected consent ref %s at index %d in response, got %s", expectedConsents[i], i, consent)
		}
	}
}

func TestRegField_Upsert_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	regFieldConfig := RegistrationFieldConfig{
		FieldKey:  "test_field",
		FieldType: "TEXT",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
	}

	_, err := regField.Upsert(context.Background(), regFieldConfig)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	// The error should come from util.HandleResponseError or util.ProcessResponse
	// We don't check for specific error message since it depends on the util implementation
}

func TestRegField_Upsert_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	regFieldConfig := RegistrationFieldConfig{
		FieldKey:  "test_field",
		FieldType: "TEXT",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
	}

	_, err := regField.Upsert(context.Background(), regFieldConfig)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// The error should come from util.ProcessResponse when trying to parse invalid JSON
}

func TestRegField_Get_Success(t *testing.T) {
	expectedRegField := RegistrationFieldConfig{
		ID:        "field-123",
		FieldKey:  "email",
		FieldType: "EMAIL",
		DataType:  "TEXT",
		Required:  true,
		Enabled:   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "fieldsetup-srv/fields/email") {
			t.Errorf("Expected fieldsetup-srv/fields/email endpoint, got %s", r.URL.Path)
		}

		response := RegistrationFieldResponse{
			Success: true,
			Status:  200,
			Data:    expectedRegField,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.Get(context.Background(), "email")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.FieldKey != expectedRegField.FieldKey {
		t.Errorf("Expected FieldKey %s, got %s", expectedRegField.FieldKey, result.Data.FieldKey)
	}
}

func TestRegField_Get_Error(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "field not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	_, err := regField.Get(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error for not found field, got nil")
	}
}

func TestRegField_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "fieldsetup-srv/fields/email") {
			t.Errorf("Expected fieldsetup-srv/fields/email endpoint, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	err := regField.Delete(context.Background(), "email")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestRegField_Delete_Error(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	err := regField.Delete(context.Background(), "test-field")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestRegField_GetAll_Success(t *testing.T) {
	expectedRegFields := []RegistrationFieldConfig{
		{
			ID:        "field-1",
			FieldKey:  "email",
			FieldType: "EMAIL",
			DataType:  "TEXT",
			Required:  true,
			Enabled:   true,
		},
		{
			ID:        "field-2",
			FieldKey:  "username",
			FieldType: "TEXT",
			DataType:  "TEXT",
			Required:  true,
			Enabled:   true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint (fieldsetup-srv/graph/fields expects POST)
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "fieldsetup-srv/graph/fields") {
			t.Errorf("Expected fieldsetup-srv/graph/fields endpoint, got %s", r.URL.Path)
		}

		response := AllRegFieldResponse{
			Success: true,
			Status:  200,
			Data:    expectedRegFields,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 registration fields, got %d", len(result))
	}

	if result[0].FieldKey != "email" {
		t.Errorf("Expected first field key 'email', got %s", result[0].FieldKey)
	}

	if result[1].FieldKey != "username" {
		t.Errorf("Expected second field key 'username', got %s", result[1].FieldKey)
	}
}

func TestRegField_GetAll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AllRegFieldResponse{
			Success: true,
			Status:  200,
			Data:    []RegistrationFieldConfig{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	result, err := regField.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 registration fields, got %d", len(result))
	}
}

func TestRegField_GetAll_Error(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	_, err := regField.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestRegField_UpdateOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "fieldsetup-srv/fields/order") {
			t.Errorf("Expected fieldsetup-srv/fields/order endpoint, got %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		var order RegistrationFieldOrder
		if err := json.Unmarshal(body, &order); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		if order.FieldKey != "sample_field" || order.CurrentOrder != 10 || order.PreviousOrder != 49 {
			t.Errorf("Unexpected order payload: %+v", order)
		}
		if order.ParentGroupID != "DEFAULT" {
			t.Errorf("Expected parent_group_id DEFAULT, got %s", order.ParentGroupID)
		}

		response := RegistrationFieldOrderResponse{Success: true, Status: 200, Data: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	err := regField.UpdateOrder(context.Background(), RegistrationFieldOrder{
		ParentGroupID: "DEFAULT",
		FieldKey:      "sample_field",
		CurrentOrder:  10,
		PreviousOrder: 49,
	})
	if err != nil {
		t.Fatalf("UpdateOrder failed: %v", err)
	}
}

func TestRegField_UpdateOrder_Error(t *testing.T) {
	server := NewMockServer(http.StatusBadRequest, `{"success":false,"status":400,"data":false}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	regField := NewRegField(config)

	err := regField.UpdateOrder(context.Background(), RegistrationFieldOrder{
		ParentGroupID: "DEFAULT",
		FieldKey:      "sample_field",
		CurrentOrder:  10,
		PreviousOrder: 99,
	})
	if err == nil {
		t.Fatal("Expected error for bad request, got nil")
	}
}
