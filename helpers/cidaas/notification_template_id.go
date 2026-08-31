package cidaas

import (
	"fmt"
	"strings"
)

// Matches notification-srv internal/base.OptInPrefix (template keys that skip processing/usage/verification in id).
const optInReminderPrefix = "OPTIN_REMINDER"

func templateIsOptInReminder(groupID, templateKey string) bool {
	if strings.EqualFold(groupID, "developer") {
		return false
	}
	return strings.HasPrefix(templateKey, optInReminderPrefix)
}

// SyntheticTemplateDocumentID builds the template document _id the same way as
// notification-srv internal/model/template.Template.CreateID when POST /templates/ has no _id.
// Used to PUT after HTTP 409 "template already found".
func SyntheticTemplateDocumentID(m NotificationsSrvTemplateModel) string {
	if m.TemplateKey == "" {
		return ""
	}
	id := m.GroupID + ":" + m.TemplateKey
	comm := strings.TrimSpace(m.CommunicationMethod)
	if comm == "" {
		return id
	}
	id += ":" + strings.ToLower(comm)
	if strings.TrimSpace(m.Locale) != "" {
		id += ":" + strings.ToLower(strings.TrimSpace(m.Locale))
		if strings.TrimSpace(m.ProcessingType) != "" && !templateIsOptInReminder(m.GroupID, m.TemplateKey) {
			id += ":" + strings.TrimSpace(m.ProcessingType)
			if strings.TrimSpace(m.UsageType) != "" {
				id += ":" + strings.TrimSpace(m.UsageType)
				if strings.TrimSpace(m.VerificationType) != "" {
					id += ":" + strings.TrimSpace(m.VerificationType)
				}
			}
		}
	}
	if m.Number != nil {
		id += fmt.Sprintf("#%d", *m.Number)
	}
	return id
}
