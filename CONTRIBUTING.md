# Contributing

## Documentation

Provider docs are Registry-native and generated with [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs).

### Source of truth

| Content | Where to edit |
|---------|----------------|
| Attribute / resource descriptions | Go `MarkdownDescription` on schemas under `internal/` |
| Copy-paste HCL | `examples/provider/` and `examples/resources/<type>/` |
| Page layout / subcategory | `templates/index.md.tmpl`, `templates/resources.md.tmpl` |
| Guides | `templates/guides/*.md` |

Do **not** hand-edit generated schema sections under `docs/` — they are overwritten by generate.

### Minimum bar for a resource change

1. Schema descriptions for new/changed attributes
2. Working example under `examples/resources/<resource_type>/`
3. Run documentation generation and commit `docs/`

### Commands

```bash
make generate   # fmt examples + tfplugindocs generate
```

### Subcategories

Resources are grouped for the Registry navigation (Apps, Consent, Hosted Pages, Notifications, Groups, Security, Identity Providers, Webhooks). Mapping lives in `templates/resources.md.tmpl`.
