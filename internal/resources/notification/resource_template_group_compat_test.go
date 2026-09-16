package notification

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestTemplateGroupSenderFieldsRemainOptional(t *testing.T) {
	t.Parallel()
	email, ok := templateGroupSchema.Attributes["email_sender_config"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("email_sender_config missing")
	}
	for _, name := range []string{"from_email", "from_name", "reply_to"} {
		attr, ok := email.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("%s missing", name)
		}
		if attr.Required {
			t.Fatalf("%s must remain Optional for existing customer configs", name)
		}
		if !attr.Optional {
			t.Fatalf("%s must be Optional", name)
		}
	}
	sms, ok := templateGroupSchema.Attributes["sms_sender_config"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("sms_sender_config missing")
	}
	fromName, ok := sms.Attributes["from_name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("sms from_name missing")
	}
	if fromName.Required || !fromName.Optional {
		t.Fatal("sms from_name must remain Optional")
	}
	if !fromName.Computed {
		t.Fatal("sms from_name must be Computed")
	}
}
