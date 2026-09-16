# System templates cannot be imported using the standard Terraform import command.
# Instead, users must create a configuration that matches the existing system template and run terraform apply.

# V3 Change Note: The format of the command is changed in V3. In V2, the import identifier was joined by the chracter "-"
# However in V3, it is replaced by the chracter ":". Example: TERRAFORM_TEMPLATE:SMS:en-us 

# Below is the command to import a custom template
# Here, template_key:template_type:locale is a combination of template_key, template_type and locale, joined by the special character ":".
# For example, if the resource name is "sample" with template_key as "TERRAFORM_TEMPLATE", template_type as "SMS" and locale as "de-de", the import statement would be:

terraform import cidaas_template.sample TERRAFORM_TEMPLATE:SMS:de-de
