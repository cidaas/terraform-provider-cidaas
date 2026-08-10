package resources

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// var allowedClaims = []string{
// 	"name", "given_name", "family_name", "nickname", "preferred_username", "profile",
// 	"picture", "website", "email", "email_verified", "gender", "middle_name", "birthdate", "zoneinfo", "locale",
// 	"phone_number", "phone_number_verified", "formatted", "street_address", "locality", "region", "postal_code", "country",
// }

var allowedProviders = []string{
	"google", "facebook", "linkedin", "amazon", "foursquare", "github",
	"instagram", "yammer", "wordpress", "microsoft", "yahoo", "officeenterprise", "salesforce", "paypal_sandbox", "paypal",
	"apple", "twitter", "netid", "netid_qa",
}

type SocialProvider struct {
	BaseResource
}

func NewSocialProvider() resource.Resource {
	return &SocialProvider{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_SOCIAL_PROVIDER,
				Schema: &socialProviderSchema,
			},
		),
	}
}

type SocialProviderConfig struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ProviderName          types.String `tfsdk:"provider_name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	Claims                types.Object `tfsdk:"claims"`
	EnabledForAdminPortal types.Bool   `tfsdk:"enabled_for_admin_portal"`
	UserInfoFields        types.List   `tfsdk:"userinfo_fields"`
	Scopes                types.Set    `tfsdk:"scopes"`

	claims         *Claims
	userInfoFields []*UserInfoFields
}

type Claims struct {
	RequiredClaims types.Object `tfsdk:"required_claims"`
	OptionalClaims types.Object `tfsdk:"optional_claims"`
	requiredClaims *ClaimConfigs
	optionalClaims *ClaimConfigs
}

type ClaimConfigs struct {
	UserInfo types.Set `tfsdk:"user_info"`
	IDToken  types.Set `tfsdk:"id_token"`
}

type UserInfoFields struct {
	InnerKey      types.String `tfsdk:"inner_key"`
	ExternalKey   types.String `tfsdk:"external_key"`
	IsCustomField types.Bool   `tfsdk:"is_custom_field"`
	IsSystemField types.Bool   `tfsdk:"is_system_field"`
}

func (r *SocialProvider) ConfigValidators(_ context.Context) []resource.ConfigValidator {
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

func (r *SocialProvider) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	var config SocialProviderConfig
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	res.Diagnostics.Append(config.extract(ctx)...)
	if !config.UserInfoFields.IsNull() {
		innerKeysCustom := make(map[string]bool)
		innerKeysSystem := make(map[string]bool)
		for _, v := range config.userInfoFields {
			if v.IsCustomField.Equal(v.IsSystemField) {
				res.Diagnostics.AddError(
					"Unexpected Resource Configure Type",
					fmt.Sprintf("For inner_key \033[1m%s\033[0m the fields is_custom_field and is_system_field cannot have the same value.", v.InnerKey.ValueString()),
				)
			}
			if v.IsCustomField.ValueBool() {
				if innerKeysCustom[v.InnerKey.ValueString()] {
					res.Diagnostics.AddError(
						"Duplicate Custom Field Key",
						fmt.Sprintf("The custom field with the inner_key \033[1m%s\033[0m is repeated. Each key must be unique.", v.InnerKey.ValueString()),
					)
				}
				innerKeysCustom[v.InnerKey.ValueString()] = true
			}
			if v.IsSystemField.ValueBool() {
				if innerKeysSystem[v.InnerKey.ValueString()] {
					res.Diagnostics.AddError(
						"Duplicate System Field Key",
						fmt.Sprintf("The system field with the inner_key \033[1m%s\033[0m  is repeated. Each key must be unique.", v.InnerKey.ValueString()),
					)
				}
				innerKeysSystem[v.InnerKey.ValueString()] = true
			}
		}
	}
}

func claimsNestedAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"user_info": types.SetType{ElemType: types.StringType},
		"id_token":  types.SetType{ElemType: types.StringType},
	}
}

func claimsSchemaAttrTypes() map[string]attr.Type {
	nested := claimsNestedAttrTypes()
	return map[string]attr.Type{
		"required_claims": types.ObjectType{AttrTypes: nested},
		"optional_claims": types.ObjectType{AttrTypes: nested},
	}
}

func emptyClaimsNestedDefault() defaults.Object {
	nested := claimsNestedAttrTypes()
	return objectdefault.StaticValue(
		types.ObjectValueMust(
			nested,
			map[string]attr.Value{
				"user_info": types.SetValueMust(types.StringType, []attr.Value{}),
				"id_token":  types.SetValueMust(types.StringType, []attr.Value{}),
			},
		),
	)
}

var socialProviderSchema = schema.Schema{ //nolint:dupl
	MarkdownDescription: "The `cidaas_social_provider` resource allows you to configure and manage social login providers within Cidaas." +
		"\n Social login providers enable users to authenticate using their existing accounts from popular social platforms such as Google, Facebook, LinkedIn and others." +
		"\n\n Ensure that the below scopes are assigned to the client:" +
		"\n- cidaas:providers_read" +
		"\n- cidaas:providers_write" +
		"\n- cidaas:providers_delete" +
		"\n\n-> **Note:** Write-Only argument `client_secret_wo` is available to use in place of `client_secret`." +
		" Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later." +
		" [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			MarkdownDescription: "The unique identifier of the social provider",
		},
		"name": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
			MarkdownDescription: "The name of the social provider configuration. This should be unique within your Cidaas environment.",
		},
		"provider_name": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
			Validators: []validator.String{
				stringvalidator.OneOf(allowedProviders...),
			},
			MarkdownDescription: "The name of the social provider. Supported values include `google`, `facebook`, `linkedin` etc.",
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "A flag to enable or disable the social provider configuration. Set to `true` to enable and `false` to disable.",
		},
		"client_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The client ID provided by the social provider. This is used to authenticate your application with the social provider.",
		},
		"client_secret": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "The client secret provided by the social provider. Exactly one of `client_secret` or `client_secret_wo` must be set. Note that this will be stored in the state file.",
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
		"scopes": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "A list of scopes of the social provider.",
		},
		"claims": schema.SingleNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "A map defining required and optional claims to be requested from the social provider.",
			Attributes: map[string]schema.Attribute{
				"required_claims": schema.SingleNestedAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Defines the claims that are required from the social provider.",
					Attributes: map[string]schema.Attribute{
						"user_info": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "A list of user information claims that are required.",
						},
						"id_token": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "A list of ID token claims that are required.",
						},
					},
					Default: emptyClaimsNestedDefault(),
				},
				"optional_claims": schema.SingleNestedAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Defines the claims that are optional from the social provider.",
					Attributes: map[string]schema.Attribute{
						"user_info": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "A list of user information claims that are optional.",
						},
						"id_token": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "A list of ID token claims that are optional.",
						},
					},
					Default: emptyClaimsNestedDefault(),
				},
			},
		},
		"userinfo_fields": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "A list of user info fields to be mapped between the social provider and Cidaas.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"inner_key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The internal key used by Cidaas.",
					},
					"external_key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The external key used by the social provider.",
					},
					"is_custom_field": schema.BoolAttribute{
						Required:            true,
						MarkdownDescription: "A flag indicating whether the field is a custom field.",
					},
					"is_system_field": schema.BoolAttribute{
						Required:            true,
						MarkdownDescription: "A flag indicating whether the field is a system field.",
					},
				},
			},
			Default: listdefault.StaticValue(
				types.ListValueMust(
					userInfoFieldsType,
					[]attr.Value{},
				),
			),
		},
		"enabled_for_admin_portal": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "A flag to enable or disable the social provider for the admin portal. Set to `true` to enable and `false` to disable.",
		},
	},
}

func (r *SocialProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config SocialProviderConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	model, diag := prepareSocialProviderModel(ctx, plan)
	if diag.HasError() {
		tflog.Error(ctx, "failed to prepare social provider model", util.H{
			"errors": diag.Errors(),
		})
		resp.Diagnostics.AddError("error preparing social provider payload ", fmt.Sprintf("Error: %+v ", diag.Errors()))
		return
	}
	// write-only: read from config, not plan; never written to state.
	if !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown() {
		model.ClientSecret = config.ClientSecretWO.ValueString()
	}
	res, err := r.cidaasClient.SocialProvider.Upsert(ctx, model)
	if err != nil {
		tflog.Error(ctx, "failed to create social provider via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create social provider", fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "successfully created social provider via API", util.H{
		"provider_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	setClaimsInfo(&plan, res.Data.Claims)
	resp.Diagnostics.Append(setUserInfoFields(&plan, res.Data.UserInfoFields)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set claims, user info fields, or state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource social provider created successfully", util.H{
		"provider_id":   res.Data.ID,
		"provider_name": plan.ProviderName.ValueString(),
	})
}

func (r *SocialProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocialProviderConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "calling Cidaas API to read social provider", util.H{
		"provider_name": state.ProviderName.ValueString(),
		"provider_id":   state.ID.ValueString(),
	})
	res, err := r.cidaasClient.SocialProvider.Get(ctx, state.ProviderName.ValueString(), state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read social provider via API", util.H{
			"provider_name": state.ProviderName.ValueString(),
			"provider_id":   state.ID.ValueString(),
			"error":         err.Error(),
		})
		resp.Diagnostics.AddError("failed to read social provider", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	// Import is detected by ClientID being null (ImportState sets only provider_name and id).
	// Outside import, leave state.ClientSecret null when client_secret_wo is in use.
	isImport := state.ClientID.IsNull()
	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.Name = util.StringValueOrNull(&res.Data.Name)
	state.ProviderName = util.StringValueOrNull(&res.Data.ProviderName)
	state.ClientID = util.StringValueOrNull(&res.Data.ClientID)
	if !state.ClientSecret.IsNull() || isImport {
		state.ClientSecret = util.StringValueOrNull(&res.Data.ClientSecret)
	}
	state.Enabled = util.BoolValueOrNull(&res.Data.Enabled)
	state.EnabledForAdminPortal = util.BoolValueOrNull(&res.Data.EnabledForAdminPortal)
	state.Scopes = util.SetValueOrNull(res.Data.Scopes)

	setClaimsInfo(&state, res.Data.Claims)
	tflog.Debug(ctx, "successfully set claims info")

	if len(res.Data.UserInfoFields) > 0 {
		tflog.Debug(ctx, "processing user info fields")
		userInfoFields, diags := types.ListValueFrom(ctx, userInfoFieldsType, res.Data.UserInfoFields)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "failed to process user info fields", util.H{
				"errors": resp.Diagnostics.Errors(),
			})
			return
		}
		state.UserInfoFields = userInfoFields
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource social provider read successfully", util.H{
		"provider_id":   res.Data.ID,
		"provider_name": res.Data.ProviderName,
	})
}

func (r *SocialProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config SocialProviderConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	model, diag := prepareSocialProviderModel(ctx, plan)
	if diag.HasError() {
		tflog.Error(ctx, "failed to prepare social provider model for update", util.H{
			"errors": diag.Errors(),
		})
		resp.Diagnostics.AddError("error preparing social provider payload ", fmt.Sprintf("Error: %+v ", diag.Errors()))
		return
	}
	if !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown() {
		model.ClientSecret = config.ClientSecretWO.ValueString()
	}

	model.ID = state.ID.ValueString()
	res, err := r.cidaasClient.SocialProvider.Upsert(ctx, model)
	if err != nil {
		tflog.Error(ctx, "failed to update social provider via API", util.H{
			"provider_id": state.ID.ValueString(),
			"error":       err.Error(),
		})
		resp.Diagnostics.AddError("failed to update social provider", fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "successfully updated social provider via API", util.H{
		"provider_id": state.ID.ValueString(),
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	setClaimsInfo(&plan, res.Data.Claims)
	resp.Diagnostics.Append(setUserInfoFields(&plan, res.Data.UserInfoFields)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set claims, user info fields, or state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "successfully completed social provider update", util.H{
		"provider_id":   state.ID.ValueString(),
		"provider_name": plan.ProviderName.ValueString(),
	})
}

func (r *SocialProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocialProviderConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.SocialProvider.Delete(ctx, state.ProviderName.ValueString(), state.ID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete social provider via API", util.H{
			"provider_name": state.ProviderName.ValueString(),
			"provider_id":   state.ID.ValueString(),
			"error":         err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete social provider", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "resource social provider deleted successfully", util.H{
		"provider_name": state.ProviderName.ValueString(),
		"provider_id":   state.ID.ValueString(),
	})
}

func (r *SocialProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'provider_name:provider_id', got: %s", id),
		)
		return
	}
	providerName := parts[0]
	providerID := parts[1]
	if !util.Contains(allowedProviders, providerName) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Invalid provider_name provided in import identifier. Valid provider_names %+v, got: %s", allowedProviders, providerName),
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("provider_name"), providerName)
	resp.State.SetAttribute(ctx, path.Root("id"), providerID)
}

func (sp *SocialProviderConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !sp.Claims.IsNull() && !sp.Claims.IsUnknown() {
		sp.claims = &Claims{}
		diags = sp.Claims.As(ctx, sp.claims, basetypes.ObjectAsOptions{})
		if !sp.claims.OptionalClaims.IsNull() && !sp.claims.OptionalClaims.IsUnknown() {
			sp.claims.optionalClaims = &ClaimConfigs{}
			diags = sp.claims.OptionalClaims.As(ctx, &sp.claims.optionalClaims, basetypes.ObjectAsOptions{})
		}
		if !sp.claims.RequiredClaims.IsNull() && !sp.claims.RequiredClaims.IsUnknown() {
			sp.claims.requiredClaims = &ClaimConfigs{}
			diags = sp.claims.RequiredClaims.As(ctx, &sp.claims.requiredClaims, basetypes.ObjectAsOptions{})
		}
	}
	if !sp.UserInfoFields.IsNull() && !sp.UserInfoFields.IsUnknown() {
		sp.userInfoFields = make([]*UserInfoFields, 0, len(sp.UserInfoFields.Elements()))
		diags = sp.UserInfoFields.ElementsAs(ctx, &sp.userInfoFields, false)
	}
	return diags
}

func prepareSocialProviderModel(ctx context.Context, plan SocialProviderConfig) (*cidaas.SocialProviderModel, diag.Diagnostics) { //nolint:dupl
	var sp cidaas.SocialProviderModel
	var diags diag.Diagnostics

	sp.Name = plan.Name.ValueString()
	sp.ProviderName = plan.ProviderName.ValueString()
	sp.ClientID = plan.ClientID.ValueString()
	sp.ClientSecret = plan.ClientSecret.ValueString()
	sp.Enabled = plan.Enabled.ValueBool()
	sp.EnabledForAdminPortal = plan.EnabledForAdminPortal.ValueBool()

	sp.Claims = &cidaas.ClaimsModel{
		RequiredClaims: cidaas.RequiredClaimsModel{
			UserInfo: []string{},
			IDToken:  []string{},
		},
		OptionalClaims: cidaas.OptionalClaimsModel{
			UserInfo: []string{},
			IDToken:  []string{},
		},
	}

	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var scopes []string
		if diags := plan.Scopes.ElementsAs(ctx, &scopes, false); diags.HasError() {
			return nil, diags
		}
		sp.Scopes = scopes
	} else {
		sp.Scopes = []string{}
	}

	if !plan.Claims.IsNull() && !plan.Claims.IsUnknown() {
		var claims struct {
			RequiredClaims struct {
				UserInfo types.Set `tfsdk:"user_info"`
				IDToken  types.Set `tfsdk:"id_token"`
			} `tfsdk:"required_claims"`
			OptionalClaims struct {
				UserInfo types.Set `tfsdk:"user_info"`
				IDToken  types.Set `tfsdk:"id_token"`
			} `tfsdk:"optional_claims"`
		}

		diags = plan.Claims.As(ctx, &claims, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		if !claims.RequiredClaims.UserInfo.IsNull() && !claims.RequiredClaims.UserInfo.IsUnknown() {
			var userInfo []string
			if diags := claims.RequiredClaims.UserInfo.ElementsAs(ctx, &userInfo, false); diags.HasError() {
				return nil, diags
			}
			sp.Claims.RequiredClaims.UserInfo = userInfo
		}

		if !claims.RequiredClaims.IDToken.IsNull() && !claims.RequiredClaims.IDToken.IsUnknown() {
			var idToken []string
			if diags := claims.RequiredClaims.IDToken.ElementsAs(ctx, &idToken, false); diags.HasError() {
				return nil, diags
			}
			sp.Claims.RequiredClaims.IDToken = idToken
		}

		if !claims.OptionalClaims.UserInfo.IsNull() && !claims.OptionalClaims.UserInfo.IsUnknown() {
			var userInfo []string
			if diags := claims.OptionalClaims.UserInfo.ElementsAs(ctx, &userInfo, false); diags.HasError() {
				return nil, diags
			}
			sp.Claims.OptionalClaims.UserInfo = userInfo
		}

		if !claims.OptionalClaims.IDToken.IsNull() && !claims.OptionalClaims.IDToken.IsUnknown() {
			var idToken []string
			if diags := claims.OptionalClaims.IDToken.ElementsAs(ctx, &idToken, false); diags.HasError() {
				return nil, diags
			}
			sp.Claims.OptionalClaims.IDToken = idToken
		}
	}

	sp.UserInfoFields = make([]cidaas.UserInfoFieldsModel, 0)

	if !plan.UserInfoFields.IsNull() && !plan.UserInfoFields.IsUnknown() {
		var userInfoFields []struct {
			InnerKey      types.String `tfsdk:"inner_key"`
			ExternalKey   types.String `tfsdk:"external_key"`
			IsCustomField types.Bool   `tfsdk:"is_custom_field"`
			IsSystemField types.Bool   `tfsdk:"is_system_field"`
		}

		diags = plan.UserInfoFields.ElementsAs(ctx, &userInfoFields, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, field := range userInfoFields {
			sp.UserInfoFields = append(sp.UserInfoFields, cidaas.UserInfoFieldsModel{
				InnerKey:      field.InnerKey.ValueString(),
				ExternalKey:   field.ExternalKey.ValueString(),
				IsCustomField: field.IsCustomField.ValueBool(),
				IsSystemField: field.IsSystemField.ValueBool(),
			})
		}
	}

	return &sp, diags
}

func setClaimsInfo(state *SocialProviderConfig, claims *cidaas.ClaimsModel) { //nolint:dupl
	attrTypes := claimsSchemaAttrTypes()
	nestedTypes := claimsNestedAttrTypes()
	emptySet := types.SetValueMust(types.StringType, []attr.Value{})

	if claims == nil {
		state.Claims = types.ObjectValueMust(
			attrTypes,
			map[string]attr.Value{
				"required_claims": types.ObjectValueMust(
					nestedTypes,
					map[string]attr.Value{
						"user_info": emptySet,
						"id_token":  emptySet,
					},
				),
				"optional_claims": types.ObjectValueMust(
					nestedTypes,
					map[string]attr.Value{
						"user_info": emptySet,
						"id_token":  emptySet,
					},
				),
			},
		)
		return
	}

	requiredUserInfo := util.SetValueOrEmpty(claims.RequiredClaims.UserInfo)
	requiredIDToken := util.SetValueOrEmpty(claims.RequiredClaims.IDToken)
	optionalUserInfo := util.SetValueOrEmpty(claims.OptionalClaims.UserInfo)
	optionalIDToken := util.SetValueOrEmpty(claims.OptionalClaims.IDToken)

	state.Claims = types.ObjectValueMust(
		attrTypes,
		map[string]attr.Value{
			"required_claims": types.ObjectValueMust(
				nestedTypes,
				map[string]attr.Value{
					"user_info": requiredUserInfo,
					"id_token":  requiredIDToken,
				},
			),
			"optional_claims": types.ObjectValueMust(
				nestedTypes,
				map[string]attr.Value{
					"user_info": optionalUserInfo,
					"id_token":  optionalIDToken,
				},
			),
		},
	)
}

func setUserInfoFields(state *SocialProviderConfig, ufs []cidaas.UserInfoFieldsModel) diag.Diagnostics {
	var diags diag.Diagnostics
	ufObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"inner_key":       types.StringType,
			"external_key":    types.StringType,
			"is_custom_field": types.BoolType,
			"is_system_field": types.BoolType,
		},
	}
	if len(ufs) < 1 {
		state.UserInfoFields = types.ListValueMust(
			ufObjectType,
			[]attr.Value{},
		)
	} else {
		var ufObjectValues []attr.Value
		for _, uf := range ufs {
			innerKey := uf.InnerKey
			externalKey := uf.ExternalKey
			isCustomField := uf.IsCustomField
			isSystemField := uf.IsSystemField
			objValue := types.ObjectValueMust(
				ufObjectType.AttrTypes,
				map[string]attr.Value{
					"inner_key":       util.StringValueOrNull(&innerKey),
					"external_key":    util.StringValueOrNull(&externalKey),
					"is_custom_field": util.BoolValueOrNull(&isCustomField),
					"is_system_field": util.BoolValueOrNull(&isSystemField),
				})
			ufObjectValues = append(ufObjectValues, objValue)
		}
		state.UserInfoFields, diags = types.ListValue(ufObjectType, ufObjectValues)
	}
	return diags
}

var userInfoFieldsType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"inner_key":       types.StringType,
		"external_key":    types.StringType,
		"is_custom_field": types.BoolType,
		"is_system_field": types.BoolType,
	},
}
