//nolint:revive
package identity_provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = (*FederationProviderResource)(nil)
	_ resource.ResourceWithConfigure      = (*FederationProviderResource)(nil)
	_ resource.ResourceWithImportState    = (*FederationProviderResource)(nil)
	_ resource.ResourceWithValidateConfig = (*FederationProviderResource)(nil)
	_ resource.ResourceWithUpgradeState   = (*FederationProviderResource)(nil)
)

var (
	federationStandardTypes     = []string{cidaas.FederationStandardTypeOAuth2, cidaas.FederationStandardTypeOIDC, cidaas.FederationStandardTypeSAML, cidaas.FederationStandardTypeLDAP}
	federationOwners            = []string{"admin", "client", "core", "system"}
	federationUserinfoSources   = []string{"USERINFOENDPOINT", "IDTOKEN", "ACCESSTOKEN"}
	federationSigningAlgorithms = []string{"RSA-SHA256", "RSA-SHA384", "RSA-SHA512", "RSA-SHA1"}
	federationSignTargets       = []string{"authnrequest", "response", "logoutrequest", "logoutresponse"}
	federationEncryptTargets    = []string{"assertion"}
	federationContentEncAlgs    = []string{"AES-128-CBC", "AES-256-CBC", "AES-128-GCM", "AES-256-GCM"}
	federationKeyTransportAlgs  = []string{"RSA-OAEP-MGF1P", "RSA-OAEP-SHA256"}
	federationOAuth2DataTypes   = []string{"string", "integer", "boolean", "dateTime"}
	federationSAMLDataTypes     = append(append([]string{}, federationOAuth2DataTypes...),
		"xsd:string", "xsd:integer", "xsd:boolean", "xsd:dateTime", "xsd:base64Binary", "saml:NameID")
	federationProviderNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// FederationProviderResource is the cidaas_federation_provider resource (v4 / Trustdesk only).
//
//nolint:revive
type FederationProviderResource struct {
	base.BaseResource
}

// NewFederationProviderResource returns the cidaas_federation_provider resource.
func NewFederationProviderResource() resource.Resource {
	return &FederationProviderResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name: base.RESOURCE_FEDERATION_PROVIDER,
			},
		),
	}
}

// Configure blocks the resource unless the provider targets cidaas 4.x.
func (r *FederationProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	if !c.RequireV4("cidaas_federation_provider", &resp.Diagnostics) {
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
}

// Schema returns the resource schema.
func (r *FederationProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = federationProviderResourceSchema()
}

// federationProviderResourceSchema builds the schema (version 1: nested oauth2/saml/ldap).
func federationProviderResourceSchema() schema.Schema {
	return schema.Schema{
		Version: 1,
		MarkdownDescription: "Manages federated identity providers on cidaas v4 (Trustdesk) via `/federation/providers`: " +
			"catalog social providers (Apple, Google, …), custom OAuth2 / OpenID Connect, SAML and LDAP. " +
			"Preferred replacement for the deprecated `cidaas_social_provider` and `cidaas_custom_provider`." +
			"\n\nSet exactly one protocol block matching `standard_type`: `oauth2` for `OAUTH2` / `OPENID_CONNECT`, `saml` for `SAML`, `ldap` for `LDAP`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the federation provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"provider_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique technical name of the provider per tenant. Only letters, digits, `_` and `-` are allowed. Changing it forces a new resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.RegexMatches(federationProviderNameRegex, "may only contain letters, digits, '_' and '-'"),
				},
			},
			"provider_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Platform of the provider (API field `provider`; `provider` itself is a reserved Terraform meta-argument): a catalog social key (`" + strings.Join(cidaas.FederationSocialProviders, "`, `") + "`) or `custom`. " +
					"Catalog social providers do not need endpoint URLs or `userinfo_source`.",
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"standard_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Protocol of the provider. One of `OAUTH2`, `OPENID_CONNECT`, `SAML`, `LDAP`.",
				Validators:          []validator.String{stringvalidator.OneOf(federationStandardTypes...)},
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Display name shown in the UI. Required when `standard_type` is `SAML`.",
			},
			"logo_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "URL of the provider logo.",
			},
			"owner": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("client"),
				MarkdownDescription: "Owner of the provider. One of `admin`, `client`, `core`, `system`. Defaults to `client` (visible in the Admin UI).",
				Validators:          []validator.String{stringvalidator.OneOf(federationOwners...)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the provider.",
			},
			"domains": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Allowed domains for this provider.",
			},
			"enabled_for_admin_portal": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the provider is available for the admin portal login.",
			},
			"attribute_mapping": federationAttributeMappingSchema(),
			"oauth2":            federationOAuth2Schema(),
			"saml":              federationSAMLSchema(),
			"ldap":              federationLDAPSchema(),
			"created_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp.",
			},
		},
	}
}

// federationAttributeMappingSchema builds the attribute_mapping attribute.
func federationAttributeMappingSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		Optional: true,
		MarkdownDescription: "Attribute mapping from the external provider to cidaas fields. The map key must equal `field_key`: a system field " +
			"(e.g. `email`, `given_name`), `customFields.<key>` or `identityCustomFields.<key>`.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"field_key": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Canonical cidaas field. Must equal the map key.",
				},
				"external_key": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Attribute name on the external provider (claim name or SAML attribute).",
				},
				"data_type": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Data type. Required for SAML. OAuth2/OIDC: `string`, `integer`, `boolean`, `dateTime`. " +
						"SAML additionally: `xsd:string`, `xsd:integer`, `xsd:boolean`, `xsd:dateTime`, `xsd:base64Binary`, `saml:NameID`.",
					Validators: []validator.String{stringvalidator.OneOf(federationSAMLDataTypes...)},
				},
				"default": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Default value used when the external attribute is missing.",
				},
				"ignore": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "Do not map this attribute. Defaults to `false`.",
				},
				"transform": schema.SingleNestedAttribute{
					Optional:            true,
					MarkdownDescription: "Optional value transformation.",
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Transformation type (e.g. `regex`, `format`, `function`).",
						},
						"config": schema.MapAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Transformation specific configuration.",
						},
					},
				},
				"email_verified_key": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "External attribute holding the email-verified flag (legacy AD migration).",
				},
			},
		},
	}
}

// federationOAuth2Schema builds the oauth2 block.
func federationOAuth2Schema() schema.SingleNestedAttribute {
	claimNames := func(desc string) schema.SingleNestedAttribute {
		return schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: desc,
			Attributes: map[string]schema.Attribute{
				"user_info": schema.ListAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Claim names read from the userinfo response.",
				},
				"id_token": schema.ListAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Claim names read from the ID token.",
				},
			},
		}
	}
	httpsURL := []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^https://`), "must be an https:// URL")}

	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "OAuth2 / OpenID Connect configuration. Required when `standard_type` is `OAUTH2` or `OPENID_CONNECT`.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Client ID issued by the external provider.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Client secret. Stored in state. Required unless `pkce` is `true`; conflicts with `client_secret_wo`.",
			},
			"client_secret_wo": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				MarkdownDescription: "Write-only client secret (Terraform >= 1.11), never stored in state. Requires `client_secret_wo_version`; change the version to send a new secret.",
			},
			"client_secret_wo_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Version marker for `client_secret_wo`. Changing it triggers an update that re-sends the secret.",
			},
			"pkce": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Use PKCE for the authorization code flow. When `true`, no client secret is required. Defaults to `false`.",
			},
			"userinfo_source": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Where user claims are read from: `USERINFOENDPOINT`, `IDTOKEN` or `ACCESSTOKEN`. Required for custom providers.",
				Validators:          []validator.String{stringvalidator.OneOf(federationUserinfoSources...)},
			},
			"authorization_endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Authorization endpoint (https). Required for custom providers unless `wellknown_endpoint` is set; optional override for catalog social providers.",
				Validators:          httpsURL,
			},
			"token_endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Token endpoint (https). Required for custom providers unless `wellknown_endpoint` is set.",
				Validators:          httpsURL,
			},
			"userinfo_endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Userinfo endpoint (https). Required for custom providers without `wellknown_endpoint` when `userinfo_source` is `USERINFOENDPOINT`.",
				Validators:          httpsURL,
			},
			"wellknown_endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OIDC discovery URL (https). Alternative to `authorization_endpoint` + `token_endpoint`. The API does not persist it; the configured value is kept in state.",
				Validators:          httpsURL,
			},
			"authorization_parameters": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional query parameters appended to the authorization request.",
			},
			"token_endpoint_auth": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Authentication used when calling the token endpoint, as JSON (`jsonencode({ apiAccessType = \"...\", ... })`).",
			},
			"scopes": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Scopes requested from the provider.",
			},
			"scopes_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Full scope configuration (label and per-scope flags).",
				Attributes: map[string]schema.Attribute{
					"display_label": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Label shown for the scopes (e.g. `Permission`).",
					},
					"scopes": schema.ListNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Scope setups.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"scope_name": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Scope name (e.g. `email`).",
								},
								"required": schema.BoolAttribute{
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
									MarkdownDescription: "Whether the scope is required. Defaults to `false`.",
								},
								"recommended": schema.BoolAttribute{
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
									MarkdownDescription: "Whether the scope is recommended. Defaults to `false`.",
								},
							},
						},
					},
				},
			},
			"claims": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Claims required or optionally read from the provider.",
				Attributes: map[string]schema.Attribute{
					"required_claims": claimNames("Claims that must be present."),
					"optional_claims": claimNames("Claims mapped when present."),
				},
			},
			"amr_config": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Mapping between external and cidaas AMR values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"amr_value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "cidaas AMR value.",
						},
						"ext_amr_value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "External AMR value.",
						},
					},
				},
			},
			"groups": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Provider specific group attribute / location.",
			},
		},
	}
}

// federationSAMLSchema builds the saml block.
func federationSAMLSchema() schema.SingleNestedAttribute {
	certificates := func(desc string) schema.ListNestedAttribute {
		return schema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: desc,
			PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"certificate": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "PEM encoded certificate.",
					},
					"sha256_fingerprint": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "SHA-256 fingerprint.",
					},
				},
			},
		}
	}
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "SAML configuration. Required when `standard_type` is `SAML`.",
		Attributes: map[string]schema.Attribute{
			"meta_data": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IdP metadata XML.",
			},
			"sign_auth_request": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Sign the SAML authentication request. Defaults to `false`.",
			},
			"user_id_attribute": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "SAML attribute used as user ID. Defaults to the NameID when unset.",
			},
			"protocol_binding_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Protocol binding method of the authentication request.",
			},
			"idp_initiated_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "IdP-initiated login configuration.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Identifier assigned by the API.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"client_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Client that receives the IdP-initiated login.",
					},
					"protocol": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Protocol used (typically `SAML`).",
					},
					"query_string": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Additional query parameters.",
					},
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Enable IdP-initiated login. Defaults to `false`.",
					},
				},
			},
			"signature_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Signing of outgoing SAML messages.",
				Attributes: map[string]schema.Attribute{
					"signing_algorithm": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "One of `RSA-SHA256`, `RSA-SHA384`, `RSA-SHA512`, `RSA-SHA1`.",
						Validators:          []validator.String{stringvalidator.OneOf(federationSigningAlgorithms...)},
					},
					"sign_targets": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Messages to sign: `authnrequest`, `response`, `logoutrequest`, `logoutresponse`.",
						Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(federationSignTargets...))},
					},
				},
			},
			"encryption_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Encryption of SAML messages.",
				Attributes: map[string]schema.Attribute{
					"encrypt_targets": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Messages to encrypt when sending: `assertion`.",
						Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(federationEncryptTargets...))},
					},
					"require_encrypted_for": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Messages that must be encrypted when received: `assertion`.",
						Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(federationEncryptTargets...))},
					},
					"content_encryption_algorithm": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "One of `AES-128-CBC`, `AES-256-CBC`, `AES-128-GCM`, `AES-256-GCM`.",
						Validators:          []validator.String{stringvalidator.OneOf(federationContentEncAlgs...)},
					},
					"key_transport_algorithm": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "One of `RSA-OAEP-MGF1P`, `RSA-OAEP-SHA256`.",
						Validators:          []validator.String{stringvalidator.OneOf(federationKeyTransportAlgs...)},
					},
				},
			},
			"sp_signing_certificates":    certificates("SP signing certificates generated by cidaas (read-only)."),
			"sp_encryption_certificates": certificates("SP encryption certificates generated by cidaas (read-only)."),
		},
	}
}

// federationLDAPSchema builds the ldap block.
func federationLDAPSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "LDAP configuration. Required when `standard_type` is `LDAP`.",
		Attributes: map[string]schema.Attribute{
			"connection_string": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "LDAP URL (`ldap://` or `ldaps://`). At least one of `connection_string` or `host` is required.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^ldaps?://`), "must start with ldap:// or ldaps://"),
				},
			},
			"host": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "LDAP host. At least one of `connection_string` or `host` is required.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "LDAP port (1-65535).",
				Validators:          []validator.Int64{int64validator.Between(1, 65535)},
			},
			"use_ssl": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use SSL/TLS.",
			},
			"bind_dn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "DN used to bind to the directory.",
			},
			"bind_credentials": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Bind password. Stored in state; conflicts with `bind_credentials_wo`.",
			},
			"bind_credentials_wo": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				MarkdownDescription: "Write-only bind password (Terraform >= 1.11), never stored in state. Requires `bind_credentials_wo_version`.",
			},
			"bind_credentials_wo_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Version marker for `bind_credentials_wo`. Changing it re-sends the password.",
			},
			"base_dn": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base DN. At least one of `base_dn` or `search_base` is required.",
			},
			"search_base": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Search base. At least one of `base_dn` or `search_base` is required.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Search filter.",
			},
			"user_filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User filter.",
			},
			"unique_field": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Attribute that uniquely identifies a user.",
			},
			"user_sync": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Synchronize users.",
			},
			"default_groups": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Groups (and roles) assigned to users from this directory.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Group ID.",
						},
						"roles": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Roles within the group.",
						},
					},
				},
			},
		},
	}
}

// ValidateConfig enforces the cross-field rules of provider-model (nest per standard_type, SAML display name,
// custom OAuth2 endpoints, secrets, LDAP connection). Unknown values are skipped; the API validates again on apply.
func (r *FederationProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	cfg := configReader{ctx: ctx, get: req.Config.GetAttribute}
	validateFederationProviderConfig(cfg, &resp.Diagnostics)
}

// configReader reads single config values by path, returning unknown when a value cannot be read.
type configReader struct {
	ctx context.Context
	get func(context.Context, path.Path, interface{}) diag.Diagnostics
}

// str reads a string attribute.
func (c configReader) str(p path.Path) types.String {
	var v types.String
	if c.get(c.ctx, p, &v).HasError() {
		return types.StringUnknown()
	}
	return v
}

// boolean reads a bool attribute.
func (c configReader) boolean(p path.Path) types.Bool {
	var v types.Bool
	if c.get(c.ctx, p, &v).HasError() {
		return types.BoolUnknown()
	}
	return v
}

// object reads an object attribute.
func (c configReader) object(p path.Path) types.Object {
	var v types.Object
	if c.get(c.ctx, p, &v).HasError() {
		return types.ObjectUnknown(nil)
	}
	return v
}

// mapping reads a map attribute.
func (c configReader) mapping(p path.Path) types.Map {
	var v types.Map
	if c.get(c.ctx, p, &v).HasError() {
		return types.MapUnknown(types.StringType)
	}
	return v
}

// validateFederationProviderConfig checks that the protocol block matches standard_type and runs the per-protocol checks.
func validateFederationProviderConfig(cfg configReader, diags *diag.Diagnostics) {
	standardType := cfg.str(path.Root("standard_type"))
	if !known(standardType) {
		return
	}
	st := standardType.ValueString()

	nests := map[string]bool{
		"oauth2": st == cidaas.FederationStandardTypeOAuth2 || st == cidaas.FederationStandardTypeOIDC,
		"saml":   st == cidaas.FederationStandardTypeSAML,
		"ldap":   st == cidaas.FederationStandardTypeLDAP,
	}
	for _, name := range []string{"oauth2", "saml", "ldap"} {
		obj := cfg.object(path.Root(name))
		if obj.IsUnknown() {
			continue
		}
		if nests[name] && obj.IsNull() {
			diags.AddAttributeError(path.Root(name), "Missing protocol configuration",
				fmt.Sprintf("`%s` is required when standard_type is %q.", name, st))
		}
		if !nests[name] && !obj.IsNull() {
			diags.AddAttributeError(path.Root(name), "Protocol configuration does not match standard_type",
				fmt.Sprintf("`%s` must not be set when standard_type is %q.", name, st))
		}
	}

	switch {
	case nests["oauth2"]:
		if !cfg.object(path.Root("oauth2")).IsNull() {
			validateFederationOAuth2Config(cfg, diags)
		}
	case nests["saml"]:
		if dn := cfg.str(path.Root("display_name")); dn.IsNull() || (known(dn) && strings.TrimSpace(dn.ValueString()) == "") {
			diags.AddAttributeError(path.Root("display_name"), "Missing display_name", "`display_name` is required when standard_type is \"SAML\".")
		}
	case nests["ldap"]:
		if !cfg.object(path.Root("ldap")).IsNull() {
			validateFederationLDAPConfig(cfg, diags)
		}
	}

	validateFederationAttributeMapping(cfg, st, diags)
}

// validateFederationOAuth2Config checks secrets and, for custom providers, userinfo_source and endpoints.
func validateFederationOAuth2Config(cfg configReader, diags *diag.Diagnostics) {
	o := path.Root("oauth2")
	secret := cfg.str(o.AtName("client_secret"))
	secretWO := cfg.str(o.AtName("client_secret_wo"))
	secretWOVersion := cfg.str(o.AtName("client_secret_wo_version"))
	pkce := cfg.boolean(o.AtName("pkce"))

	if !secret.IsNull() && !secretWO.IsNull() {
		diags.AddAttributeError(o.AtName("client_secret_wo"), "Conflicting client secrets", "Set only one of `client_secret` or `client_secret_wo`.")
	}
	if !secretWO.IsNull() && secretWOVersion.IsNull() {
		diags.AddAttributeError(o.AtName("client_secret_wo_version"), "Missing client_secret_wo_version", "`client_secret_wo_version` is required with `client_secret_wo`.")
	}
	pkceEnabled := known(pkce) && pkce.ValueBool()
	if !pkce.IsUnknown() && !pkceEnabled && secret.IsNull() && secretWO.IsNull() {
		diags.AddAttributeError(o.AtName("client_secret"), "Missing client secret",
			"One of `client_secret` or `client_secret_wo` is required when `pkce` is not enabled.")
	}

	provider := cfg.str(path.Root("provider_type"))
	if !known(provider) || cidaas.IsFederationSocialProvider(provider.ValueString()) {
		return
	}

	// Custom provider: userinfo_source and endpoints are mandatory.
	source := cfg.str(o.AtName("userinfo_source"))
	if source.IsNull() {
		diags.AddAttributeError(o.AtName("userinfo_source"), "Missing userinfo_source",
			fmt.Sprintf("`userinfo_source` is required for custom providers (provider %q is not a catalog social provider).", provider.ValueString()))
	}
	wellknown := cfg.str(o.AtName("wellknown_endpoint"))
	if !wellknown.IsNull() {
		return
	}
	for _, name := range []string{"authorization_endpoint", "token_endpoint"} {
		if cfg.str(o.AtName(name)).IsNull() {
			diags.AddAttributeError(o.AtName(name), "Missing endpoint",
				fmt.Sprintf("`%s` is required for custom providers unless `wellknown_endpoint` is set.", name))
		}
	}
	if known(source) && source.ValueString() == "USERINFOENDPOINT" && cfg.str(o.AtName("userinfo_endpoint")).IsNull() {
		diags.AddAttributeError(o.AtName("userinfo_endpoint"), "Missing endpoint",
			"`userinfo_endpoint` is required when `userinfo_source` is \"USERINFOENDPOINT\" and `wellknown_endpoint` is not set.")
	}
}

// validateFederationLDAPConfig checks the LDAP connection, search base and bind credentials.
func validateFederationLDAPConfig(cfg configReader, diags *diag.Diagnostics) {
	l := path.Root("ldap")
	if cfg.str(l.AtName("connection_string")).IsNull() && cfg.str(l.AtName("host")).IsNull() {
		diags.AddAttributeError(l, "Missing LDAP connection", "At least one of `connection_string` or `host` is required.")
	}
	if cfg.str(l.AtName("base_dn")).IsNull() && cfg.str(l.AtName("search_base")).IsNull() {
		diags.AddAttributeError(l, "Missing LDAP search base", "At least one of `base_dn` or `search_base` is required.")
	}
	creds := cfg.str(l.AtName("bind_credentials"))
	credsWO := cfg.str(l.AtName("bind_credentials_wo"))
	if !creds.IsNull() && !credsWO.IsNull() {
		diags.AddAttributeError(l.AtName("bind_credentials_wo"), "Conflicting bind credentials", "Set only one of `bind_credentials` or `bind_credentials_wo`.")
	}
	if !credsWO.IsNull() && cfg.str(l.AtName("bind_credentials_wo_version")).IsNull() {
		diags.AddAttributeError(l.AtName("bind_credentials_wo_version"), "Missing bind_credentials_wo_version", "`bind_credentials_wo_version` is required with `bind_credentials_wo`.")
	}
}

// validateFederationAttributeMapping checks that map keys equal field_key and data_type fits the protocol.
func validateFederationAttributeMapping(cfg configReader, standardType string, diags *diag.Diagnostics) {
	mappings := cfg.mapping(path.Root("attribute_mapping"))
	if !known(mappings) {
		return
	}
	for key := range mappings.Elements() {
		p := path.Root("attribute_mapping").AtMapKey(key)
		fieldKey := cfg.str(p.AtName("field_key"))
		if known(fieldKey) && fieldKey.ValueString() != key {
			diags.AddAttributeError(p.AtName("field_key"), "attribute_mapping key mismatch",
				fmt.Sprintf("field_key %q must equal the map key %q.", fieldKey.ValueString(), key))
		}
		dataType := cfg.str(p.AtName("data_type"))
		if standardType == cidaas.FederationStandardTypeSAML && dataType.IsNull() {
			diags.AddAttributeError(p.AtName("data_type"), "Missing data_type", "`data_type` is required for SAML attribute mappings.")
		}
		if standardType != cidaas.FederationStandardTypeSAML && known(dataType) && !util.Contains(federationOAuth2DataTypes, dataType.ValueString()) {
			diags.AddAttributeError(p.AtName("data_type"), "Invalid data_type",
				fmt.Sprintf("data_type %q is only valid for SAML; use one of %s.", dataType.ValueString(), strings.Join(federationOAuth2DataTypes, ", ")))
		}
	}
}

// Create creates the provider and stores the API response in state.
func (r *FederationProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := expandFederationProvider(&plan, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create federation provider", util.FormatErrorMessage(err))
		return
	}

	state := flattenFederationProvider(&res.Data, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state from the API and removes the resource from state when it no longer exists.
func (r *FederationProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state federationProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Get(ctx, state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("Failed to read federation provider", util.FormatErrorMessage(err))
		return
	}

	newState := flattenFederationProvider(&res.Data, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update replaces the provider via PUT and stores the API response in state.
func (r *FederationProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := expandFederationProvider(&plan, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Update(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update federation provider", util.FormatErrorMessage(err))
		return
	}

	newState := flattenFederationProvider(&res.Data, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete removes the provider. A provider that is already gone is not an error.
func (r *FederationProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state federationProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.FederationProvider.Delete(ctx, state.ID.ValueString())
	if err != nil && !util.IsResourceNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete federation provider", util.FormatErrorMessage(err))
		return
	}
}

// ImportState imports a provider by its id.
func (r *FederationProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// State upgrade: v0 (flat OAuth2 attributes, <= 4.1.x) -> v1 (nested oauth2/saml/ldap)
// ---------------------------------------------------------------------------

// federationProviderModelV0 is the flat state model of schema version 0.
type federationProviderModelV0 struct {
	ID                    types.String `tfsdk:"id"`
	ProviderName          types.String `tfsdk:"provider_name"`
	DisplayName           types.String `tfsdk:"display_name"`
	StandardType          types.String `tfsdk:"standard_type"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	UserinfoEndpoint      types.String `tfsdk:"userinfo_endpoint"`
	LogoURL               types.String `tfsdk:"logo_url"`
	Domains               types.List   `tfsdk:"domains"`
	Owner                 types.String `tfsdk:"owner"`
}

// federationProviderSchemaV0 is the flat schema of version 0, used to read prior state.
func federationProviderSchemaV0() schema.Schema {
	str := func() schema.StringAttribute { return schema.StringAttribute{Optional: true} }
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                       schema.StringAttribute{Computed: true},
			"provider_name":            schema.StringAttribute{Required: true},
			"display_name":             schema.StringAttribute{Required: true},
			"standard_type":            schema.StringAttribute{Required: true},
			"client_id":                schema.StringAttribute{Required: true},
			"client_secret":            schema.StringAttribute{Optional: true, Sensitive: true},
			"client_secret_wo":         schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true},
			"client_secret_wo_version": str(),
			"authorization_endpoint":   str(),
			"token_endpoint":           str(),
			"userinfo_endpoint":        str(),
			"logo_url":                 str(),
			"domains":                  schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"owner":                    schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

// UpgradeState registers the v0 -> v1 state upgrader.
func (r *FederationProviderResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	prior := federationProviderSchemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &prior,
			StateUpgrader: upgradeFederationProviderStateV0,
		},
	}
}

// upgradeFederationProviderStateV0 reads v0 state and writes the upgraded v1 state.
func upgradeFederationProviderStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior federationProviderModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	upgraded := upgradeFederationProviderV0(prior)
	resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
}

// upgradeFederationProviderV0 moves the flat OAuth2 attributes into the oauth2 nest. v0 always sent
// provider = "custom" and userinfoSource = USERINFOENDPOINT, so those are carried over explicitly.
func upgradeFederationProviderV0(prior federationProviderModelV0) federationProviderModel {
	domains := prior.Domains
	if domains.IsNull() || domains.IsUnknown() {
		domains = types.ListNull(types.StringType)
	}
	return federationProviderModel{
		ID:                    prior.ID,
		ProviderName:          prior.ProviderName,
		Provider:              types.StringValue("custom"),
		StandardType:          prior.StandardType,
		DisplayName:           prior.DisplayName,
		LogoURL:               prior.LogoURL,
		Owner:                 prior.Owner,
		Description:           types.StringNull(),
		Domains:               domains,
		EnabledForAdminPortal: types.BoolNull(),
		CreatedTime:           types.StringNull(),
		UpdatedTime:           types.StringNull(),
		OAuth2: &federationOAuth2Model{
			ClientID:                prior.ClientID,
			ClientSecret:            prior.ClientSecret,
			ClientSecretWO:          types.StringNull(),
			ClientSecretWOVersion:   prior.ClientSecretWOVersion,
			Pkce:                    types.BoolValue(false),
			UserinfoSource:          types.StringValue("USERINFOENDPOINT"),
			AuthorizationEndpoint:   prior.AuthorizationEndpoint,
			TokenEndpoint:           prior.TokenEndpoint,
			UserinfoEndpoint:        prior.UserinfoEndpoint,
			WellknownEndpoint:       types.StringNull(),
			AuthorizationParameters: types.MapNull(types.StringType),
			TokenEndpointAuth:       types.StringNull(),
			Scopes:                  types.ListNull(types.StringType),
			Groups:                  types.StringNull(),
		},
	}
}

// ---------------------------------------------------------------------------
// Data source support
// ---------------------------------------------------------------------------

// FederationProviderResourceSchema exposes the resource schema so the data source can derive its
// (computed, write-only free) schema from it and stay in sync.
func FederationProviderResourceSchema() schema.Schema {
	return federationProviderResourceSchema()
}

// federationProviderObjectType returns the object type of the resource schema.
func federationProviderObjectType() types.ObjectType {
	return federationProviderResourceSchema().Type().(types.ObjectType) //nolint:forcetypeassert
}

// FederationProviderDataSourceValue converts an API provider into a value matching dsType (the data source
// schema type, which is the resource schema without write-only attributes).
func FederationProviderDataSourceValue(ctx context.Context, api *cidaas.FederationProviderModel, dsType attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := flattenFederationProvider(api, nil)
	obj, d := types.ObjectValueFrom(ctx, federationProviderObjectType().AttrTypes, model)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	out, d := conformValue(obj, dsType)
	diags.Append(d...)
	return out, diags
}
