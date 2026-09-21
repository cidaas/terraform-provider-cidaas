#!/bin/bash

# Import script for cidaas_notification_template_type resource
# 
# Usage:
#   terraform import cidaas_notification_template_type.<resource_name> <template_key>
#
# Examples:
#   # Import a system template type
#   terraform import cidaas_notification_template_type.verify_user VERIFY_USER
#
#   # Import a custom template type
#   terraform import cidaas_notification_template_type.custom_notification CUSTOM_NOTIFICATION
#
# Note:
#   - The template_key must be uppercase (e.g., VERIFY_USER, not verify_user)
#   - System template types are pre-provisioned and can only have custom_attributes modified
#   - Custom template types can be fully managed after import

# Example import commands (uncomment and modify as needed):
# terraform import cidaas_notification_template_type.verify_user_system VERIFY_USER
# terraform import cidaas_notification_template_type.custom_notification CUSTOM_NOTIFICATION
# terraform import cidaas_notification_template_type.simple_notification SIMPLE_NOTIFICATION

