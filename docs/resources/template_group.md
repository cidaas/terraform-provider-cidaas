**Deprecated for new designs:** this resource uses legacy **templates-srv/groups**. For notification-srv template groups, use **`cidaas_notifications_template_group`**.

The cidaas_template_group resource in the provider is used to define and manage templates groups within the Cidaas system. Template Groups categorize your communication templates allowing you to map preferred templates to specific clients effectively.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:
- cidaas:templates_read
- cidaas:templates_write
- cidaas:templates_delete