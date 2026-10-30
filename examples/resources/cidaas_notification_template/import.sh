# The import ID is the template document _id: {group_id}:{template_key}:{communication_method}:{locale}[:{processing_type}[:{usage_type}[:{verification_type}]]]
# Language-only locales (e.g. de, en) are valid, as used by seeded templates.

terraform import cidaas_notification_template.verify_user_email_de "default:VERIFY_USER:email:de:CODE:MULTIFACTOR_AUTHENTICATION:EMAIL"
