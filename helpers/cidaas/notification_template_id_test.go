package cidaas

import "testing"

func TestSyntheticTemplateDocumentID(t *testing.T) {
	n := 2
	tests := []struct {
		name string
		m    NotificationsSrvTemplateModel
		want string
	}{
		{
			name: "verify account link en",
			m: NotificationsSrvTemplateModel{
				GroupID: "default", TemplateKey: "VERIFY_ACCOUNT",
				CommunicationMethod: "email", Locale: "en", ProcessingType: "LINK",
			},
			want: "default:VERIFY_ACCOUNT:email:en:LINK",
		},
		{
			name: "locale casing",
			m: NotificationsSrvTemplateModel{
				GroupID: "default", TemplateKey: "WELCOME_USER",
				CommunicationMethod: "email", Locale: "en-US",
			},
			want: "default:WELCOME_USER:email:en-us",
		},
		{
			name: "with number suffix",
			m: NotificationsSrvTemplateModel{
				GroupID: "default", TemplateKey: "FOO",
				CommunicationMethod: "sms", Locale: "de",
				Number: &n,
			},
			want: "default:FOO:sms:de#2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SyntheticTemplateDocumentID(tt.m); got != tt.want {
				t.Fatalf("SyntheticTemplateDocumentID() = %q, want %q", got, tt.want)
			}
		})
	}
}
