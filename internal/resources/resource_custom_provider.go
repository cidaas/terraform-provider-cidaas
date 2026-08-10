package resources

import (
	"context"
	"regexp"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type CustomProvider struct {
	BaseResource
}

func NewCustomProvider() resource.Resource {
	return &CustomProvider{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_CUSTOM_PROVIDER,
				Schema: &customProviderSchema,
			},
		),
	}
}

type ProviderConfig struct {
	ID                    types.String `tfsdk:"id"`
	ProviderName          types.String `tfsdk:"provider_name"`
	DisplayName           types.String `tfsdk:"display_name"`
	LogoURL               types.String `tfsdk:"logo_url"`
	StandardType          types.String `tfsdk:"standard_type"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	UserinfoEndpoint      types.String `tfsdk:"userinfo_endpoint"`
	ScopeDisplayLabel     types.String `tfsdk:"scope_display_label"`
	Domains               types.Set    `tfsdk:"domains"`
	Scopes                types.List   `tfsdk:"scopes"`
	UserinfoFields        types.Object `tfsdk:"userinfo_fields"`
	scopes                []*CpScope
	userinfoFields        *UserInfoField
	AmrConfig             types.List   `tfsdk:"amr_config"`
	UserinfoSource        types.String `tfsdk:"userinfo_source"`
	Pkce                  types.Bool   `tfsdk:"pkce"`
	AuthType              types.String `tfsdk:"auth_type"`
	APIKeyDetails         types.Object `tfsdk:"apikey_details"`
	TOTPDetails           types.Object `tfsdk:"totp_details"`
	CidaasAuthDetails     types.Object `tfsdk:"cidaas_auth_details"`
	apiKeyDetails         *AuthConfig
	totpDetails           *AuthConfig
	cidaasAuthDetails     *CidaasAuthConfig
}

type CpScope struct {
	ScopeName   types.String `tfsdk:"scope_name"`
	Required    types.Bool   `tfsdk:"required"`
	Recommended types.Bool   `tfsdk:"recommended"`
}

type UfNestedObject struct {
	ExtFieldKey types.String `tfsdk:"ext_field_key"`
	Default     types.String `tfsdk:"default"`
}

type UfEmailVerifiedNestedObject struct {
	ExtFieldKey types.String `tfsdk:"ext_field_key"`
	Default     types.Bool   `tfsdk:"default"`
}

type UserInfoField struct {
	Name              types.Object `tfsdk:"name"`
	FamilyName        types.Object `tfsdk:"family_name"`
	GivenName         types.Object `tfsdk:"given_name"`
	MiddleName        types.Object `tfsdk:"middle_name"`
	Nickname          types.Object `tfsdk:"nickname"`
	PreferredUsername types.Object `tfsdk:"preferred_username"`
	Profile           types.Object `tfsdk:"profile"`
	Picture           types.Object `tfsdk:"picture"`
	Website           types.Object `tfsdk:"website"`
	Gender            types.Object `tfsdk:"gender"`
	Birthdate         types.Object `tfsdk:"birthdate"`
	Zoneinfo          types.Object `tfsdk:"zoneinfo"`
	Locale            types.Object `tfsdk:"locale"`
	UpdatedAt         types.Object `tfsdk:"updated_at"`
	Email             types.Object `tfsdk:"email"`
	EmailVerified     types.Object `tfsdk:"email_verified"`
	PhoneNumber       types.Object `tfsdk:"phone_number"`
	MobileNumber      types.Object `tfsdk:"mobile_number"`
	Address           types.Object `tfsdk:"address"`
	Sub               types.Object `tfsdk:"sub"`
	Groups            types.Object `tfsdk:"groups"`

	name              *UfNestedObject
	familyName        *UfNestedObject
	givenName         *UfNestedObject
	middleName        *UfNestedObject
	nickname          *UfNestedObject
	preferredUsername *UfNestedObject
	profile           *UfNestedObject
	picture           *UfNestedObject
	website           *UfNestedObject
	gender            *UfNestedObject
	birthdate         *UfNestedObject
	zoneinfo          *UfNestedObject
	locale            *UfNestedObject
	updatedAt         *UfNestedObject
	email             *UfNestedObject
	emailVerified     *UfEmailVerifiedNestedObject
	phoneNumber       *UfNestedObject
	mobileNumber      *UfNestedObject
	address           *UfNestedObject
	sub               *UfNestedObject
	groups            *UfNestedObject

	CustomFields types.Map `tfsdk:"custom_fields"`
}

type AmrConfig struct {
	AmrValue    types.String `tfsdk:"amr_value"`
	ExtAmrValue types.String `tfsdk:"ext_amr_value"`
}

func (pc *ProviderConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !pc.UserinfoFields.IsNull() {
		pc.userinfoFields = &UserInfoField{}
		diags = pc.UserinfoFields.As(ctx, pc.userinfoFields, basetypes.ObjectAsOptions{})

		extractField := func(field types.Object, target **UfNestedObject) {
			if !field.IsNull() && !field.IsUnknown() {
				*target = &UfNestedObject{}
				diags.Append(field.As(ctx, *target, basetypes.ObjectAsOptions{})...)
			}
		}

		extractField(pc.userinfoFields.Name, &pc.userinfoFields.name)
		extractField(pc.userinfoFields.FamilyName, &pc.userinfoFields.familyName)
		extractField(pc.userinfoFields.GivenName, &pc.userinfoFields.givenName)
		extractField(pc.userinfoFields.MiddleName, &pc.userinfoFields.middleName)
		extractField(pc.userinfoFields.Nickname, &pc.userinfoFields.nickname)
		extractField(pc.userinfoFields.PreferredUsername, &pc.userinfoFields.preferredUsername)
		extractField(pc.userinfoFields.Profile, &pc.userinfoFields.profile)
		extractField(pc.userinfoFields.Picture, &pc.userinfoFields.picture)
		extractField(pc.userinfoFields.Website, &pc.userinfoFields.website)
		extractField(pc.userinfoFields.Gender, &pc.userinfoFields.gender)
		extractField(pc.userinfoFields.Birthdate, &pc.userinfoFields.birthdate)
		extractField(pc.userinfoFields.Zoneinfo, &pc.userinfoFields.zoneinfo)
		extractField(pc.userinfoFields.Locale, &pc.userinfoFields.locale)
		extractField(pc.userinfoFields.UpdatedAt, &pc.userinfoFields.updatedAt)
		extractField(pc.userinfoFields.Email, &pc.userinfoFields.email)
		extractField(pc.userinfoFields.PhoneNumber, &pc.userinfoFields.phoneNumber)
		extractField(pc.userinfoFields.MobileNumber, &pc.userinfoFields.mobileNumber)
		extractField(pc.userinfoFields.Address, &pc.userinfoFields.address)
		extractField(pc.userinfoFields.Sub, &pc.userinfoFields.sub)
		extractField(pc.userinfoFields.Groups, &pc.userinfoFields.groups)

		if !pc.userinfoFields.EmailVerified.IsNull() && !pc.userinfoFields.EmailVerified.IsUnknown() {
			pc.userinfoFields.emailVerified = &UfEmailVerifiedNestedObject{}
			diags.Append(pc.userinfoFields.EmailVerified.As(ctx, &pc.userinfoFields.emailVerified, basetypes.ObjectAsOptions{})...)
		}
	}

	if !pc.Scopes.IsNull() {
		pc.scopes = make([]*CpScope, 0, len(pc.Scopes.Elements()))
		diags.Append(pc.Scopes.ElementsAs(ctx, &pc.scopes, false)...)
	}

	if !pc.APIKeyDetails.IsNull() {
		pc.apiKeyDetails = &AuthConfig{}
		diags = pc.APIKeyDetails.As(ctx, pc.apiKeyDetails, basetypes.ObjectAsOptions{})
	}
	if !pc.TOTPDetails.IsNull() {
		pc.totpDetails = &AuthConfig{}
		diags = pc.TOTPDetails.As(ctx, pc.totpDetails, basetypes.ObjectAsOptions{})
	}
	if !pc.CidaasAuthDetails.IsNull() {
		pc.cidaasAuthDetails = &CidaasAuthConfig{}
		diags = pc.CidaasAuthDetails.As(ctx, pc.cidaasAuthDetails, basetypes.ObjectAsOptions{})
	}
	return diags
}

var customProviderSchema = schema.Schema{
	MarkdownDescription: "This example demonstrates the configuration of a custom provider resource for interacting with Cidaas." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:providers_read" +
		"\n- cidaas:providers_write" +
		"\n- cidaas:providers_delete" +
		"\n\n-> **Note:** Write-Only argument `client_secret_wo` is available to use in place of `client_secret`." +
		" Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later." +
		" [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The ID of the resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"provider_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The unique identifier of the custom provider. This cannot be updated for an existing state.",
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"display_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The display name of the provider.",
		},
		"logo_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL for the provider's logo.",
		},
		"standard_type": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Type of standard. Allowed values `OAUTH2` and `OPENID_CONNECT`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"OPENID_CONNECT", "OAUTH2"}...),
			},
		},
		"client_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The client ID of the provider.",
		},
		"client_secret": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "The client secret of the provider. Exactly one of `client_secret` or `client_secret_wo` must be set. Note that this will be stored in the state file.",
		},
		"client_secret_wo": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			WriteOnly:           true,
			MarkdownDescription: "Write-Only equivalent of `client_secret`. The value is sent to cidaas on create and update but is not stored in plan or state. Must be set together with `client_secret_wo_version`. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later.",
		},
		"client_secret_wo_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Used together with `client_secret_wo` to trigger an update. Increment this value when an update to `client_secret_wo` is required.",
		},
		"authorization_endpoint": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The URL for authorization of the provider.",
		},
		"token_endpoint": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The URL to generate token with this provider.",
		},
		"userinfo_endpoint": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The URL to fetch user details using this provider.",
		},
		// In plan Set deletes an existing record and create a whole new one, so preferred list. However, to allow only unique values use set
		"scopes": schema.ListNestedAttribute{
			MarkdownDescription: "List of scopes of the provider with details",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"scope_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The name of the scope, e.g., `openid`, `profile`.",
					},
					"required": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Indicates if the scope is required.",
						Default:             booldefault.StaticBool(false),
					},
					"recommended": schema.BoolAttribute{
						MarkdownDescription: "Indicates if the scope is recommended.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			Optional: true,
		},
		"scope_display_label": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Display label for the scope of the provider.",
		},
		"userinfo_fields": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			MarkdownDescription: "Object containing various user information fields with their values." +
				" The userinfo_fields section includes specific fields such as name, family_name, address, etc., along with custom_fields allowing additional user information customization",
			Attributes: map[string]schema.Attribute{
				"name":               createStandardNestedAttribute(),
				"family_name":        createStandardNestedAttribute(),
				"given_name":         createStandardNestedAttribute(),
				"middle_name":        createStandardNestedAttribute(),
				"nickname":           createStandardNestedAttribute(),
				"preferred_username": createStandardNestedAttribute(),
				"profile":            createStandardNestedAttribute(),
				"picture":            createStandardNestedAttribute(),
				"website":            createStandardNestedAttribute(),
				"gender":             createStandardNestedAttribute(),
				"birthdate":          createStandardNestedAttribute(),
				"zoneinfo":           createStandardNestedAttribute(),
				"locale":             createStandardNestedAttribute(),
				"updated_at":         createStandardNestedAttribute(),
				"email":              createStandardNestedAttribute(),
				"phone_number":       createStandardNestedAttribute(),
				"mobile_number":      createStandardNestedAttribute(),
				"address":            createStandardNestedAttribute(),
				"sub":                createStandardNestedAttribute(),
				"groups":             createStandardNestedAttribute(),

				"email_verified": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"ext_field_key": schema.StringAttribute{
							Optional: true,
						},
						"default": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							// API returns false when unset; keep schema default aligned to avoid drift.
							Default: booldefault.StaticBool(false),
						},
					},
				},

				"custom_fields": schema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
			Default: objectdefault.StaticValue(userInfoDefaultValue()),
		},
		"domains": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The domains of the provider.",
			Optional:            true,
		},
		"amr_config": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "AMR configuration mapping.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"amr_value": schema.StringAttribute{
						Required: true,
					},
					"ext_amr_value": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
		"userinfo_source": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Source of userinfo. Allowed values are `IDTOKEN` and `USERINFOENDPOINT`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"IDTOKEN", "USERINFOENDPOINT"}...),
			},
		},
		"pkce": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The flag to enable or disable pkce flow. By default, the value is set to `false`",
			Default:             booldefault.StaticBool(false),
		},
		"auth_type": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Type of authentication. Allowed values `APIKEY`, `CIDAAS_OAUTH2` and `TOTP`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"APIKEY", "CIDAAS_OAUTH2", "TOTP"}...),
			},
		},
		"apikey_details": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Configuration for API key-based authentication. It's a **required** parameter when the auth_type is APIKEY.",
			Attributes: map[string]schema.Attribute{
				"placeholder": schema.StringAttribute{
					Required:    true,
					Description: "The attribute is the placeholder for the key which need to be passed as a query parameter or in the request header.",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-z-]+$`),
							"must contain only lowercase alphabets",
						),
					},
				},
				"placement": schema.StringAttribute{
					Required: true,
					Description: "The placement of the API key in the request (e.g., query)." +
						"The allowed value are `header` and `query`.",
					Validators: []validator.String{
						stringvalidator.OneOf(cidaas.AllowedKeyPlacementValue...),
					},
				},
				"key": schema.StringAttribute{
					Required: true,
					Description: "The API key that will be used to authenticate the webhook request." +
						"The key that will be passed in the request header or in query param as configured in the attribute `placement`",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
		"totp_details": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Configuration for TOTP based authentication.  It's a **required** parameter when the auth_type is TOTP.",
			Attributes: map[string]schema.Attribute{
				"placeholder": schema.StringAttribute{
					Required:    true,
					Description: "A placeholder value for the TOTP.",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-z-]+$`),
							"must contain only lowercase alphabets",
						),
					},
				},
				"placement": schema.StringAttribute{
					Required: true,
					Description: "The placement of the TOTP in the request." +
						"The allowed value are `header` and `query`.",
					Validators: []validator.String{
						stringvalidator.OneOf(cidaas.AllowedKeyPlacementValue...),
					},
				},
				"key": schema.StringAttribute{
					Required:    true,
					Description: "The key used for TOTP authentication.",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
		"cidaas_auth_details": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Configuration for Cidaas authentication. It's a **required** parameter when the auth_type is CIDAAS_OAUTH2.",
			Attributes: map[string]schema.Attribute{
				"client_id": schema.StringAttribute{
					Required:    true,
					Description: "The client ID for Cidaas authentication.",
				},
			},
		},
	},
}

func (r *CustomProvider) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("client_secret"),
			path.MatchRoot("client_secret_wo"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("client_secret_wo"),
			path.MatchRoot("client_secret_wo_version"),
		),
		resourcevalidator.PreferWriteOnlyAttribute(
			path.MatchRoot("client_secret"),
			path.MatchRoot("client_secret_wo"),
		),
	}
}

func (r *CustomProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config ProviderConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	cp, d := prepareCpRequestPayload(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to prepare custom provider model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	if !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown() {
		cp.ClientSecret = config.ClientSecretWO.ValueString()
	}
	res, err := r.cidaasClient.CustomProvider.CreateCustomProvider(ctx, cp)
	if err != nil {
		tflog.Error(ctx, "Failed to create custom provider via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create custom provider", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "Successfully created custom provider via API", util.H{
		"provider_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource custom provider created successfully")
}

func (r *CustomProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProviderConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	res, err := r.cidaasClient.CustomProvider.GetCustomProvider(ctx, state.ProviderName.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read custom provider", util.FormatErrorMessage(err))
		return
	}
	state.StandardType = util.StringValueOrNull(&res.Data.StandardType)
	state.AuthorizationEndpoint = util.StringValueOrNull(&res.Data.AuthorizationEndpoint)
	state.TokenEndpoint = util.StringValueOrNull(&res.Data.TokenEndpoint)
	state.ProviderName = util.StringValueOrNull(&res.Data.ProviderName)
	state.DisplayName = util.StringValueOrNull(&res.Data.DisplayName)
	state.LogoURL = util.StringValueOrNull(&res.Data.LogoURL)
	state.UserinfoEndpoint = util.StringValueOrNull(&res.Data.UserinfoEndpoint)
	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.ScopeDisplayLabel = util.StringValueOrNull(&res.Data.Scopes.DisplayLabel)
	// Import is detected by ClientID being null (only provider_name is set by ImportState).
	// Outside import, leave state.ClientSecret null when client_secret_wo is in use.
	isImport := state.ClientID.IsNull()
	state.ClientID = util.StringValueOrNull(&res.Data.ClientID)
	if !state.ClientSecret.IsNull() || isImport {
		state.ClientSecret = util.StringValueOrNull(&res.Data.ClientSecret)
	}
	state.Domains = util.SetValueOrNull(res.Data.Domains)
	state.Pkce = types.BoolValue(res.Data.Pkce)
	state.AuthType = util.StringValueOrNull(&res.Data.AuthType)

	var diag diag.Diagnostics
	var objectValues []attr.Value
	scopeObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"scope_name":  types.StringType,
			"required":    types.BoolType,
			"recommended": types.BoolType,
		},
	}

	for _, sc := range res.Data.Scopes.Scopes {
		scopeName := sc.ScopeName
		required := sc.Required
		recommended := sc.Recommended
		objValue := types.ObjectValueMust(scopeObjectType.AttrTypes, map[string]attr.Value{
			"scope_name":  util.StringValueOrNull(&scopeName),
			"required":    util.BoolValueOrNull(&required),
			"recommended": util.BoolValueOrNull(&recommended),
		})
		objectValues = append(objectValues, objValue)
	}

	state.Scopes, diag = types.ListValueFrom(ctx, scopeObjectType, objectValues)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	standardFields := []string{
		"name", "family_name", "given_name", "middle_name", "nickname",
		"preferred_username", "profile", "picture", "website", "gender",
		"birthdate", "zoneinfo", "locale", "updated_at", "email",
		"phone_number", "mobile_number", "address", "sub", "groups",
	}

	metadataAttributeTypes := make(map[string]attr.Type)
	metadataAttributes := make(map[string]attr.Value)

	standardNestedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ext_field_key": types.StringType,
			"default":       types.StringType,
		},
	}

	emailVerifiedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ext_field_key": types.StringType,
			"default":       types.BoolType,
		},
	}

	for _, field := range standardFields {
		metadataAttributeTypes[field] = standardNestedType
		metadataAttributes[field] = types.ObjectNull(standardNestedType.AttrTypes)
	}

	metadataAttributeTypes["email_verified"] = emailVerifiedType
	metadataAttributes["email_verified"] = types.ObjectNull(emailVerifiedType.AttrTypes)

	customFields := map[string]attr.Value{}
	hasCustomfield := false

	setStandardUserinfoField := func(key string, fieldMap map[string]interface{}) {
		if key == "email_verified" {
			extFieldKey, _ := fieldMap["extFieldKey"].(string)
			defaultValue, _ := fieldMap["default"].(bool)
			metadataAttributes[key] = types.ObjectValueMust(
				metadataAttributeTypes[key].(types.ObjectType).AttrTypes,
				map[string]attr.Value{
					"ext_field_key": types.StringValue(extFieldKey),
					"default":       types.BoolValue(defaultValue),
				},
			)
			return
		}
		extFieldKey, _ := fieldMap["extFieldKey"].(string)
		defaultValue, hasDefault := fieldMap["default"].(string)

		var defaultAttrValue attr.Value
		if hasDefault && defaultValue != "" {
			defaultAttrValue = types.StringValue(defaultValue)
		} else {
			defaultAttrValue = types.StringNull()
		}

		metadataAttributes[key] = types.ObjectValueMust(
			metadataAttributeTypes[key].(types.ObjectType).AttrTypes,
			map[string]attr.Value{
				"ext_field_key": types.StringValue(extFieldKey),
				"default":       defaultAttrValue,
			},
		)
	}

	for key, fieldInterface := range res.Data.UserinfoFields {
		fieldMap, ok := fieldInterface.(map[string]interface{})
		if !ok {
			continue
		}

		if strings.HasPrefix(key, "customFields.") {
			customFieldKey := strings.TrimPrefix(key, "customFields.")
			// API sometimes stores standard fields under customFields.<name>; promote those
			// so state matches the schema default / config shape and avoids perpetual drift.
			if _, isStandard := metadataAttributeTypes[customFieldKey]; isStandard {
				setStandardUserinfoField(customFieldKey, fieldMap)
				continue
			}
			if extFieldKey, exists := fieldMap["extFieldKey"].(string); exists {
				customFields[customFieldKey] = types.StringValue(extFieldKey)
				hasCustomfield = true
			}
			continue
		}

		if _, exists := metadataAttributeTypes[key]; exists {
			setStandardUserinfoField(key, fieldMap)
		}
	}

	metadataAttributeTypes["custom_fields"] = types.MapType{ElemType: types.StringType}
	if hasCustomfield {
		metadataAttributes["custom_fields"] = types.MapValueMust(types.StringType, customFields)
	} else {
		metadataAttributes["custom_fields"] = types.MapNull(types.StringType)
	}

	state.UserinfoFields, diag = types.ObjectValue(metadataAttributeTypes, metadataAttributes)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(res.Data.AmrConfig) > 0 {
		amrObjectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"amr_value":     types.StringType,
				"ext_amr_value": types.StringType,
			},
		}

		amrValues := make([]attr.Value, 0, len(res.Data.AmrConfig))
		for _, amrConfig := range res.Data.AmrConfig {
			amrValues = append(amrValues, types.ObjectValueMust(
				amrObjectType.AttrTypes,
				map[string]attr.Value{
					"amr_value":     types.StringValue(amrConfig.AmrValue),
					"ext_amr_value": types.StringValue(amrConfig.ExtAmrValue),
				},
			))
		}

		state.AmrConfig, diag = types.ListValueFrom(ctx, amrObjectType, amrValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		state.AmrConfig = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"amr_value":     types.StringType,
				"ext_amr_value": types.StringType,
			},
		})
	}

	// API defaults userInfoSource to USERINFOENDPOINT when unset; keep prior null to avoid drift.
	if !state.UserinfoSource.IsNull() || isImport {
		state.UserinfoSource = util.StringValueOrNull(&res.Data.UserInfoSource)
	}

	authConfig := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"placeholder": types.StringType,
			"placement":   types.StringType,
			"key":         types.StringType,
		},
	}

	if res.Data.APIKeyDetails.Apikey != "" {
		apiKeyConfig, diags := types.ObjectValue(authConfig.AttrTypes, map[string]attr.Value{
			"placeholder": util.StringValueOrNull(&res.Data.APIKeyDetails.ApikeyPlaceholder),
			"placement":   util.StringValueOrNull(&res.Data.APIKeyDetails.ApikeyPlacement),
			"key":         util.StringValueOrNull(&res.Data.APIKeyDetails.Apikey),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.APIKeyDetails = apiKeyConfig
	}
	if res.Data.TotpDetails.TotpKey != "" {
		totpConfig, diags := types.ObjectValue(authConfig.AttrTypes, map[string]attr.Value{
			"placeholder": util.StringValueOrNull(&res.Data.TotpDetails.TotpPlaceholder),
			"placement":   util.StringValueOrNull(&res.Data.TotpDetails.TotpPlacement),
			"key":         util.StringValueOrNull(&res.Data.TotpDetails.TotpKey),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.TOTPDetails = totpConfig
	}

	if res.Data.CidaasAuthDetails.ClientID != "" {
		oauthConfig, diags := types.ObjectValue(map[string]attr.Type{
			"client_id": types.StringType,
		}, map[string]attr.Value{
			"client_id": util.StringValueOrNull(&res.Data.CidaasAuthDetails.ClientID),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.CidaasAuthDetails = oauthConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CustomProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config ProviderConfig

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to get plan/state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	cp, d := prepareCpRequestPayload(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to prepare custom provider payload for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	if !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown() {
		cp.ClientSecret = config.ClientSecretWO.ValueString()
	}

	cp.ID = state.ID.ValueString()
	err := r.cidaasClient.CustomProvider.UpdateCustomProvider(ctx, cp)
	if err != nil {
		tflog.Error(ctx, "Failed to update custom provider via API", util.H{
			"provider_id": state.ID.ValueString(),
			"error":       err.Error(),
		})
		resp.Diagnostics.AddError("failed to update custom provider", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "Successfully updated custom provider via API", util.H{
		"provider_id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource custom provider updated successfully")
}

func (r *CustomProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProviderConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.CustomProvider.DeleteCustomProvider(ctx, state.ProviderName.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to delete custom provider via API", util.H{
			"provider_name": state.ProviderName.ValueString(),
			"error":         err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete custom provider", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource custom provider deleted successfully", util.H{
		"provider_name": state.ProviderName.ValueString(),
	})
}

func (r *CustomProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("provider_name"), req, resp)
}

func prepareCpRequestPayload(ctx context.Context, plan ProviderConfig) (*cidaas.CustomProviderModel, diag.Diagnostics) {
	var cp cidaas.CustomProviderModel
	var diags diag.Diagnostics

	cp.StandardType = plan.StandardType.ValueString()
	cp.AuthorizationEndpoint = plan.AuthorizationEndpoint.ValueString()
	cp.TokenEndpoint = plan.TokenEndpoint.ValueString()
	cp.ProviderName = plan.ProviderName.ValueString()
	cp.DisplayName = plan.DisplayName.ValueString()
	cp.LogoURL = plan.LogoURL.ValueString()
	cp.UserinfoEndpoint = plan.UserinfoEndpoint.ValueString()
	cp.ClientID = plan.ClientID.ValueString()
	cp.ClientSecret = plan.ClientSecret.ValueString()
	cp.Pkce = plan.Pkce.ValueBool()
	cp.AuthType = plan.AuthType.ValueString()

	diags = plan.Domains.ElementsAs(ctx, &cp.Domains, false)
	if diags.HasError() {
		return nil, diags
	}
	var childScopes []cidaas.ScopeChild
	for _, v := range plan.scopes {
		childScopes = append(childScopes, cidaas.ScopeChild{
			ScopeName:   v.ScopeName.ValueString(),
			Required:    v.Required.ValueBool(),
			Recommended: v.Recommended.ValueBool(),
		})
	}
	cp.Scopes.Scopes = childScopes
	cp.Scopes.DisplayLabel = plan.ScopeDisplayLabel.ValueString()

	if !plan.UserinfoFields.IsNull() {
		cp.UserinfoFields = make(map[string]interface{})

		addUserInfoField := func(fieldName string, obj types.Object, nestedObj *UfNestedObject) {
			if !obj.IsNull() && nestedObj != nil && !nestedObj.ExtFieldKey.IsNull() {
				cp.UserinfoFields[fieldName] = &cidaas.UserInfoField{
					ExtFieldKey: nestedObj.ExtFieldKey.ValueString(),
					Default:     nestedObj.Default.ValueString(),
				}
			}
		}

		addUserInfoField("name", plan.userinfoFields.Name, plan.userinfoFields.name)
		addUserInfoField("family_name", plan.userinfoFields.FamilyName, plan.userinfoFields.familyName)
		addUserInfoField("given_name", plan.userinfoFields.GivenName, plan.userinfoFields.givenName)
		addUserInfoField("middle_name", plan.userinfoFields.MiddleName, plan.userinfoFields.middleName)
		addUserInfoField("nickname", plan.userinfoFields.Nickname, plan.userinfoFields.nickname)
		addUserInfoField("preferred_username", plan.userinfoFields.PreferredUsername, plan.userinfoFields.preferredUsername)
		addUserInfoField("profile", plan.userinfoFields.Profile, plan.userinfoFields.profile)
		addUserInfoField("picture", plan.userinfoFields.Picture, plan.userinfoFields.picture)
		addUserInfoField("website", plan.userinfoFields.Website, plan.userinfoFields.website)
		addUserInfoField("gender", plan.userinfoFields.Gender, plan.userinfoFields.gender)
		addUserInfoField("birthdate", plan.userinfoFields.Birthdate, plan.userinfoFields.birthdate)
		addUserInfoField("zoneinfo", plan.userinfoFields.Zoneinfo, plan.userinfoFields.zoneinfo)
		addUserInfoField("locale", plan.userinfoFields.Locale, plan.userinfoFields.locale)
		addUserInfoField("updated_at", plan.userinfoFields.UpdatedAt, plan.userinfoFields.updatedAt)
		addUserInfoField("email", plan.userinfoFields.Email, plan.userinfoFields.email)
		addUserInfoField("phone_number", plan.userinfoFields.PhoneNumber, plan.userinfoFields.phoneNumber)
		addUserInfoField("mobile_number", plan.userinfoFields.MobileNumber, plan.userinfoFields.mobileNumber)
		addUserInfoField("address", plan.userinfoFields.Address, plan.userinfoFields.address)
		addUserInfoField("sub", plan.userinfoFields.Sub, plan.userinfoFields.sub)
		addUserInfoField("groups", plan.userinfoFields.Groups, plan.userinfoFields.groups)

		if !plan.userinfoFields.EmailVerified.IsNull() &&
			plan.userinfoFields.emailVerified != nil &&
			!plan.userinfoFields.emailVerified.ExtFieldKey.IsNull() {
			cp.UserinfoFields["email_verified"] = &cidaas.UserInfoFieldBoolean{
				ExtFieldKey: plan.userinfoFields.emailVerified.ExtFieldKey.ValueString(),
				Default:     plan.userinfoFields.emailVerified.Default.ValueBool(),
			}
		}

		if !plan.userinfoFields.CustomFields.IsNull() {
			var cfMap map[string]string
			diags.Append(plan.userinfoFields.CustomFields.ElementsAs(ctx, &cfMap, false)...)
			if diags.HasError() {
				return nil, diags
			}

			for key, value := range cfMap {
				cp.UserinfoFields["customFields."+key] = &cidaas.UserInfoField{
					ExtFieldKey: value,
				}
			}
		}
	}

	if !plan.AmrConfig.IsNull() {
		var amrConfigs []AmrConfig
		diags.Append(plan.AmrConfig.ElementsAs(ctx, &amrConfigs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		cp.AmrConfig = make([]cidaas.AmrConfig, len(amrConfigs))
		for i, config := range amrConfigs {
			cp.AmrConfig[i] = cidaas.AmrConfig{
				AmrValue:    config.AmrValue.ValueString(),
				ExtAmrValue: config.ExtAmrValue.ValueString(),
			}
		}
	}

	if !plan.UserinfoSource.IsNull() {
		cp.UserInfoSource = plan.UserinfoSource.ValueString()
	}

	if !plan.APIKeyDetails.IsNull() {
		cp.APIKeyDetails = cidaas.APIKeyDetails{
			ApikeyPlaceholder: plan.apiKeyDetails.Placeholder.ValueString(),
			ApikeyPlacement:   plan.apiKeyDetails.Placement.ValueString(),
			Apikey:            plan.apiKeyDetails.Key.ValueString(),
		}
	}
	if !plan.TOTPDetails.IsNull() {
		cp.TotpDetails = cidaas.TotpDetails{
			TotpPlaceholder: plan.totpDetails.Placeholder.ValueString(),
			TotpPlacement:   plan.totpDetails.Placement.ValueString(),
			TotpKey:         plan.totpDetails.Key.ValueString(),
		}
	}
	if !plan.CidaasAuthDetails.IsNull() {
		cp.CidaasAuthDetails = cidaas.AuthDetails{
			ClientID: plan.cidaasAuthDetails.ClientID.ValueString(),
		}
	}

	return &cp, diags
}

func userInfoDefaultValue() basetypes.ObjectValue {
	standardFieldType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ext_field_key": types.StringType,
			"default":       types.StringType,
		},
	}

	emailVerifiedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ext_field_key": types.StringType,
			"default":       types.BoolType,
		},
	}

	attrTypes := map[string]attr.Type{
		"name":               standardFieldType,
		"family_name":        standardFieldType,
		"given_name":         standardFieldType,
		"middle_name":        standardFieldType,
		"nickname":           standardFieldType,
		"preferred_username": standardFieldType,
		"profile":            standardFieldType,
		"picture":            standardFieldType,
		"website":            standardFieldType,
		"gender":             standardFieldType,
		"birthdate":          standardFieldType,
		"zoneinfo":           standardFieldType,
		"locale":             standardFieldType,
		"updated_at":         standardFieldType,
		"email":              standardFieldType,
		"email_verified":     emailVerifiedType,
		"phone_number":       standardFieldType,
		"mobile_number":      standardFieldType,
		"address":            standardFieldType,
		"sub":                standardFieldType,
		"groups":             standardFieldType,
		"custom_fields":      types.MapType{ElemType: types.StringType},
	}

	assignedValue := func(value string) basetypes.ObjectValue {
		standardFieldValue := types.ObjectValueMust(
			standardFieldType.AttrTypes,
			map[string]attr.Value{
				"ext_field_key": types.StringValue(value),
				"default":       types.StringNull(),
			},
		)
		return standardFieldValue
	}

	attrValues := map[string]attr.Value{
		"name":               assignedValue("name"),
		"family_name":        assignedValue("family_name"),
		"given_name":         assignedValue("given_name"),
		"middle_name":        assignedValue("middle_name"),
		"nickname":           assignedValue("nickname"),
		"preferred_username": assignedValue("preferred_username"),
		"profile":            assignedValue("profile"),
		"picture":            assignedValue("picture"),
		"website":            assignedValue("website"),
		"gender":             assignedValue("gender"),
		"birthdate":          assignedValue("birthdate"),
		"zoneinfo":           assignedValue("zoneinfo"),
		"locale":             assignedValue("locale"),
		"updated_at":         assignedValue("updated_at"),
		"email":              assignedValue("email"),
		"email_verified": types.ObjectValueMust(
			emailVerifiedType.AttrTypes,
			map[string]attr.Value{
				"ext_field_key": types.StringValue("email_verified"),
				"default":       types.BoolValue(false),
			},
		),
		"phone_number":  assignedValue("phone_number"),
		"mobile_number": assignedValue("mobile_number"),
		"address":       assignedValue("address"),
		"sub":           assignedValue("sub"),
		"groups":        assignedValue("groups"),
		"custom_fields": types.MapNull(types.StringType),
	}

	return types.ObjectValueMust(attrTypes, attrValues)
}

func createStandardNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"ext_field_key": schema.StringAttribute{
				Required: true,
			},
			"default": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
