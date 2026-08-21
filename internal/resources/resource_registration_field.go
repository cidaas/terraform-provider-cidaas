package resources

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const layout = "2006-01-02T15:04:05Z"

var allowedDataTypes = []string{
	"TEXT", "NUMBER", "SELECT", "MULTISELECT", "RADIO", "CHECKBOX", "PASSWORD", "DATE", "URL", "EMAIL",
	"TEXTAREA", "MOBILE", "CONSENT", "JSON_STRING", "USERNAME", "ARRAY", "GROUPING", "DAYDATE",
}

var regFieldOrderMutex sync.Mutex

type RegFieldResource struct {
	BaseResource
}

func NewRegFieldResource() resource.Resource {
	return &RegFieldResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_REGISTRATION_FIELD,
				Schema: &regFieldSchema,
			},
		),
	}
}

var _ resource.ResourceWithModifyPlan = (*RegFieldResource)(nil)

// ModifyPlan sets field_definition.regex to the RE2 shape-merged composition of
// regexes so plan matches apply (avoids UseStateForUnknown keeping a stale regex).
func (r *RegFieldResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan RegFieldConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(plan.ExtractConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !fieldDefinitionHasRegexes(plan.fieldDefinition) {
		unknown := plan.fieldDefinition != nil && plan.fieldDefinition.Regex.IsUnknown()
		resp.Diagnostics.Append(ensureFieldDefinitionRegexKnown(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if unknown {
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
		return
	}

	var patterns []string
	resp.Diagnostics.Append(plan.fieldDefinition.Regexes.ElementsAs(ctx, &patterns, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	composed, err := composeANDRegexes(patterns)
	if err != nil {
		resp.Diagnostics.AddError("Invalid field_definition.regexes", err.Error())
		return
	}
	resp.Diagnostics.Append(syncComposedRegexIntoPlan(ctx, &plan, composed)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

type RegFieldConfig struct {
	ID                                  types.String `tfsdk:"id"`
	BaseDataType                        types.String `tfsdk:"base_data_type"`
	ParentGroupID                       types.String `tfsdk:"parent_group_id"`
	FieldType                           types.String `tfsdk:"field_type"`
	DataType                            types.String `tfsdk:"data_type"`
	FieldKey                            types.String `tfsdk:"field_key"`
	Required                            types.Bool   `tfsdk:"required"`
	Internal                            types.Bool   `tfsdk:"internal"`
	Claimable                           types.Bool   `tfsdk:"claimable"`
	IsSearchable                        types.Bool   `tfsdk:"is_searchable"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	Unique                              types.Bool   `tfsdk:"unique"`
	OverwriteWithNullFromSocialProvider types.Bool   `tfsdk:"overwrite_with_null_value_from_social_provider"`
	ReadOnly                            types.Bool   `tfsdk:"read_only"`
	IsList                              types.Bool   `tfsdk:"is_list"`
	Order                               types.Int64  `tfsdk:"order"`
	Scopes                              types.Set    `tfsdk:"scopes"`
	ConsentRefs                         types.Set    `tfsdk:"consent_refs"`
	LocalTexts                          types.List   `tfsdk:"local_texts"`
	FieldDefinition                     types.Object `tfsdk:"field_definition"`
	RemoteFieldSettings                 types.Object `tfsdk:"remote_field_settings"`

	localTexts          []*LocalTexts
	fieldDefinition     *FieldDefinition
	remoteFieldSettings *RemoteFieldSettingsConfig
}

// RemoteFieldSettingsConfig holds Terraform config for remote_field_settings (GROUPING only).
type RemoteFieldSettingsConfig struct {
	CallOnce       types.Bool   `tfsdk:"call_once"`
	ApiClientSetup types.Object `tfsdk:"api_client_setup"` //nolint:revive // API/json field names
	apiClientSetup *ApiClientSetupConfig
}

// ApiClientSetupConfig holds Terraform config for api_client_setup.
type ApiClientSetupConfig struct { //nolint:revive // API/json field names
	CommunicationEP    types.String `tfsdk:"communication_ep"`
	HTTPMethod         types.String `tfsdk:"http_method"`
	APIAccessType      types.String `tfsdk:"api_access_type"`
	ApikeyConfig       types.Object `tfsdk:"apikey_config"`
	TotpConfig         types.Object `tfsdk:"totp_config"`
	BasicAuthConfig    types.Object `tfsdk:"basic_auth_config"`
	CidaasOAuth2Config types.Object `tfsdk:"cidaas_oauth2_config"`
	GenOAuth2Config    types.Object `tfsdk:"gen_oauth2_config"`
}

type ApikeyConfig struct {
	Apikey            types.String `tfsdk:"apikey"`
	ApikeyPlaceholder types.String `tfsdk:"apikey_placeholder"`
	ApikeyPlacement   types.String `tfsdk:"apikey_placement"`
}

type TotpConfig struct {
	TotpKey         types.String `tfsdk:"totpkey"`
	TotpPlaceholder types.String `tfsdk:"totp_placeholder"`
	TotpPlacement   types.String `tfsdk:"totp_placement"`
}

type BasicAuthConfig struct {
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type CidaasOAuth2Config struct {
	ClientID  types.String `tfsdk:"client_id"`
	ReqScopes types.String `tfsdk:"req_scopes"`
}

type GenOAuth2Config struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	WellknownURL types.String `tfsdk:"wellknown_url"`
	ReqScopes    types.String `tfsdk:"req_scopes"`
}

type LocalTexts struct {
	Locale       types.String `tfsdk:"locale"`
	Name         types.String `tfsdk:"name"`
	MaxLengthMsg types.String `tfsdk:"max_length_msg"`
	MinLengthMsg types.String `tfsdk:"min_length_msg"`
	RequiredMsg  types.String `tfsdk:"required_msg"`
	MatchWithMsg types.String `tfsdk:"match_with_msg"`
	Attributes   types.List   `tfsdk:"attributes"`
	ConsentLabel types.Object `tfsdk:"consent_label"`

	attributes []*Attributes
	consent    *Consent
}

type Attributes struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type Consent struct {
	Label     types.String `tfsdk:"label"`
	LabelText types.String `tfsdk:"label_text"`
}

type FieldDefinition struct {
	MaxLength       types.Int64  `tfsdk:"max_length"`
	MinLength       types.Int64  `tfsdk:"min_length"`
	MinDate         types.String `tfsdk:"min_date"`
	MaxDate         types.String `tfsdk:"max_date"`
	InitialDateView types.String `tfsdk:"initial_date_view"`
	InitialDate     types.String `tfsdk:"initial_date"`
	Regex           types.String `tfsdk:"regex"`
	Regexes         types.List   `tfsdk:"regexes"`
	MatchWith       types.String `tfsdk:"match_with"`
}

func fieldDefinitionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_length":        types.Int64Type,
		"min_length":        types.Int64Type,
		"min_date":          types.StringType,
		"max_date":          types.StringType,
		"initial_date_view": types.StringType,
		"initial_date":      types.StringType,
		"regex":             types.StringType,
		"regexes":           types.ListType{ElemType: types.StringType},
		"match_with":        types.StringType,
	}
}

var regFieldSchema = schema.Schema{
	MarkdownDescription: "The `cidaas_registration_field` in the provider allows management of registration fields in the Cidaas system." +
		" This resource enables you to configure and customize the fields displayed during user registration." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:field_setup_read" +
		"\n- cidaas:field_setup_write" +
		"\n- cidaas:field_setup_delete\n",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The ID of the resource",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		// "string", "double", "datetime", "bool", "array"
		"base_data_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The base data type of the field. This is computed property.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"parent_group_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The ID of the parent registration group. Defaults to `DEFAULT` if not provided.",
			Default:             stringdefault.StaticString("DEFAULT"),
		},
		"field_type": schema.StringAttribute{
			Optional: true,
			Computed: true,
			MarkdownDescription: "Specifies whether the field type is `SYSTEM` or `CUSTOM`. Defaults to `CUSTOM`." +
				" This cannot be modified for an existing resource. `SYSTEM` fields cannot be created but can be modified. To modify an existing field import it first and then update.",
			Default: stringdefault.StaticString("CUSTOM"),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"CUSTOM", "SYSTEM"}...),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
				&fieldTypeModifier{},
			},
		},
		"data_type": schema.StringAttribute{
			Required: true,
			MarkdownDescription: "The data type of the field. This cannot be modified for an existing resource." +
				fmt.Sprintf(" Allowed values are %s", func() string {
					var temp string
					for _, v := range allowedDataTypes {
						temp += fmt.Sprintf("`%s`,", v)
					}
					return temp
				}()),
			Validators: []validator.String{
				stringvalidator.OneOf(allowedDataTypes...),
				&dataTypeValidator{},
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"field_key": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The unique identifier of the registration field. This cannot be modified for an existing resource.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"required": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is required in registration. Defaults set to `false`",
			Default:             booldefault.StaticBool(false),
			Validators: []validator.Bool{
				&validateIsRequiredMsgAvailable{},
			},
		},
		"internal": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is internal. Defaults set to `false`",
			Default:             booldefault.StaticBool(false),
		},
		"claimable": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is claimable. Defaults set to `true`",
			Default:             booldefault.StaticBool(true),
		},
		"is_searchable": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is searchable. Defaults set to `true`",
			Default:             booldefault.StaticBool(true),
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is enabled. Defaults set to `true`",
			Default:             booldefault.StaticBool(true),
		},
		"unique": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is unique. Defaults set to `false`",
			Default:             booldefault.StaticBool(false),
		},
		// set to true if you want the value should be reset by identity provider
		"overwrite_with_null_value_from_social_provider": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Set to true if you want the value should be reset by identity provider. Defaults set to `false`",
			Default:             booldefault.StaticBool(false),
		},
		"read_only": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Flag to mark if a field is read only. Defaults set to `false`",
			Default:             booldefault.StaticBool(false),
		},
		"is_list": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		// optional: Order of the Field in the UI
		"order": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The display order of the field in the registration UI. When omitted, the API assigns an order.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"scopes": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "The scopes of the registration field.",
		},
		"consent_refs": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "List of consents(the ids of the consent in cidaas must be passed) in registration. The data type must be `CONSENT` in this case",
		},
		"local_texts": schema.ListNestedAttribute{
			Required:            true,
			MarkdownDescription: "The localized detail of the registration field.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"locale": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The locale of the field. example: de-DE.",
						Default:             stringdefault.StaticString("en"),
						Validators: []validator.String{
							stringvalidator.OneOf(
								func() []string {
									validLocals := make([]string, len(util.Locales))
									for i, locale := range util.Locales {
										validLocals[i] = locale.LocaleString
									}
									return validLocals
								}()...),
						},
					},
					"name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The name of the field in the local configured. for example: in **en-US** the name is `Sample Field` in de-DE `Beispielfeld`.",
					},
					"max_length_msg": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "warning/error msg to show to the user when user exceeds the maximum character configured. This is applicable only for the attributes of base_data_type string.",
					},
					"min_length_msg": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "warning/error msg to show to the user when user don't provide the minimum character required. This is applicable only for the attributes of base_data_type string.",
					},
					"required_msg": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Message shown when the field is required but empty. Must be provided when `required` is true. May also be set when `required` is false to pre-define translations for fields marked required at the application level.",
					},
					"match_with_msg": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Localized error message when field values do not match the referenced field. Only allowed when `field_key` is `password_echo`. Required in every locale when `field_definition.match_with` is set.",
						Validators: []validator.String{
							&validateMatchWithMsg{},
						},
					},
					// optional: in case of datatype is RADIO, SELECT, MULTISELECT, etc. the localized attribute values are specified here
					"attributes": schema.ListNestedAttribute{
						Optional:            true,
						MarkdownDescription: "The field attributes must be provided for the data_type SELECT, MULTISELECT and RADIO. it's an array of key value pairs. Example provided in the example section.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Required: true,
								},
								"value": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
					"consent_label": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "required when data_type is CONSENT. Example provided in the example section.",
						Attributes: map[string]schema.Attribute{
							"label": schema.StringAttribute{
								Required: true,
							},
							"label_text": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
		"field_definition": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"max_length": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "The maximum length of a string type attribute.",
					Validators: []validator.Int64{
						&validateIsMaxMinMsgAvailable{},
					},
				},
				"min_length": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "The minimum length of a string type attribute",
					Validators: []validator.Int64{
						&validateIsMaxMinMsgAvailable{},
					},
				},
				"min_date": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The earliest date a user can select. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.",
					Validators: []validator.String{
						&dateTypeValidator{},
						&dateValidator{},
					},
				},
				"max_date": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The maximum date a user can select. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.",
					Validators: []validator.String{
						&dateTypeValidator{},
						&dateValidator{},
					},
				},
				"initial_date_view": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The view of the calender. Applicable only for DATE attributes. Allowed values: `month`, `year` and `multi-year`",
					Validators: []validator.String{
						&dateTypeValidator{},
						stringvalidator.OneOf("month", "year", "multi-year"),
					},
				},
				"initial_date": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The initial date. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.",
					Validators: []validator.String{
						&dateTypeValidator{},
						&dateValidator{},
					},
				},
				"regex": schema.StringAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "A single regular expression stored as `fieldDefinition.regex` in the API. " +
						"Must be valid Go `regexp` (RE2) syntax — cidaas evaluates it in the backend. " +
						"Only allowed for data types TEXT and URL. Mutually exclusive with `regexes`. " +
						"When `regexes` is set, this attribute is the RE2 shape-merged result in plan and state.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
					Validators: []validator.String{
						&validateIsMaxMinMsgAvailableForRegex{},
						&validateRegexExclusive{},
					},
				},
				"regexes": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					MarkdownDescription: "List of Go `regexp` (RE2) full-string patterns merged with AND into one `fieldDefinition.regex`. " +
						"Supported shapes: length (`^.{m,n}$`), charset (`^[…]*$` / `^[…]+$`), contains (`^.*[…].*$`), no_leading (`^[^x].*$`). " +
						"Not string concatenation and not JavaScript lookaheads — unknown or unmergable shapes fail closed. " +
						"Equivalence holds for supported shapes (same accept/reject as matching every entry). " +
						"Mutually exclusive with `regex`. Not a 1:1 for Zod ErrorKeys. " +
						"Requires `min_length_msg` and `max_length_msg` in every `local_texts` entry, same as `regex`.",
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						&validateIsMaxMinMsgAvailableForRegexes{},
						&validateRegexesExclusive{},
					},
				},
				"match_with": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The `field_key` of another field whose value this field must match (e.g. password confirmation). Only allowed when `field_key` is `password_echo`.",
					Validators: []validator.String{
						&validateMatchWith{},
					},
				},
			},
			Default: objectdefault.StaticValue(types.ObjectValueMust(
				fieldDefinitionAttrTypes(),
				map[string]attr.Value{
					"max_length":        types.Int64Null(),
					"min_length":        types.Int64Null(),
					"min_date":          types.StringNull(),
					"max_date":          types.StringNull(),
					"initial_date_view": types.StringNull(),
					"initial_date":      types.StringNull(),
					"regex":             types.StringNull(),
					"regexes":           types.ListNull(types.StringType),
					"match_with":        types.StringNull(),
				})),
		},
	},
	Blocks: map[string]schema.Block{
		"remote_field_settings": schema.SingleNestedBlock{
			MarkdownDescription: "Remote field settings for GROUPING fields that fetch data from an external API. Only valid when data_type is GROUPING.",
			Attributes: map[string]schema.Attribute{
				"call_once": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "When true, the remote API is called once per session.",
				},
			},
			Blocks: map[string]schema.Block{
				"api_client_setup": schema.SingleNestedBlock{
					MarkdownDescription: "API client configuration for the remote endpoint.",
					Attributes: map[string]schema.Attribute{
						"communication_ep": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The remote API endpoint URL. Supports {{sub}} placeholder.",
						},
						"http_method": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "HTTP method (e.g. GET, POST).",
						},
						"api_access_type": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Authentication type: APIKEY, TOTP, BASIC_AUTH, CIDAAS_OAUTH2, or GEN_OAUTH2.",
						},
					},
					Blocks: map[string]schema.Block{
						"apikey_config": schema.SingleNestedBlock{
							MarkdownDescription: "Required when api_access_type is APIKEY.",
							Attributes: map[string]schema.Attribute{
								"apikey":             schema.StringAttribute{Optional: true},
								"apikey_placeholder": schema.StringAttribute{Optional: true},
								"apikey_placement":   schema.StringAttribute{Optional: true},
							},
						},
						"totp_config": schema.SingleNestedBlock{
							MarkdownDescription: "Required when api_access_type is TOTP.",
							Attributes: map[string]schema.Attribute{
								"totpkey":          schema.StringAttribute{Optional: true},
								"totp_placeholder": schema.StringAttribute{Optional: true},
								"totp_placement":   schema.StringAttribute{Optional: true},
							},
						},
						"basic_auth_config": schema.SingleNestedBlock{
							MarkdownDescription: "Required when api_access_type is BASIC_AUTH.",
							Attributes: map[string]schema.Attribute{
								"user":     schema.StringAttribute{Optional: true},
								"password": schema.StringAttribute{Optional: true},
							},
						},
						"cidaas_oauth2_config": schema.SingleNestedBlock{
							MarkdownDescription: "Required when api_access_type is CIDAAS_OAUTH2.",
							Attributes: map[string]schema.Attribute{
								"client_id":  schema.StringAttribute{Optional: true},
								"req_scopes": schema.StringAttribute{Optional: true},
							},
						},
						"gen_oauth2_config": schema.SingleNestedBlock{
							MarkdownDescription: "Required when api_access_type is GEN_OAUTH2.",
							Attributes: map[string]schema.Attribute{
								"client_id":     schema.StringAttribute{Optional: true},
								"client_secret": schema.StringAttribute{Optional: true},
								"wellknown_url": schema.StringAttribute{Optional: true},
								"req_scopes":    schema.StringAttribute{Optional: true},
							},
						},
					},
				},
			},
		},
	},
}

func (rfc *RegFieldConfig) ExtractConfigs(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !rfc.FieldDefinition.IsNull() && !rfc.FieldDefinition.IsUnknown() {
		rfc.fieldDefinition = &FieldDefinition{}
		diags.Append(rfc.FieldDefinition.As(ctx, rfc.fieldDefinition, basetypes.ObjectAsOptions{})...)
	}
	if !rfc.LocalTexts.IsNull() && !rfc.LocalTexts.IsUnknown() {
		rfc.localTexts = make([]*LocalTexts, 0, len(rfc.LocalTexts.Elements()))
		diags.Append(rfc.LocalTexts.ElementsAs(ctx, &rfc.localTexts, false)...)
		for _, localText := range rfc.localTexts {
			if !localText.Attributes.IsNull() && !localText.Attributes.IsUnknown() {
				localText.attributes = make([]*Attributes, 0, len(localText.Attributes.Elements()))
				diags.Append(localText.Attributes.ElementsAs(ctx, &localText.attributes, false)...)
			}

			if !localText.ConsentLabel.IsNull() && !localText.ConsentLabel.IsUnknown() {
				localText.consent = &Consent{}
				diags.Append(localText.ConsentLabel.As(ctx, localText.consent, basetypes.ObjectAsOptions{})...)
			}
		}
	}
	if !rfc.RemoteFieldSettings.IsNull() && !rfc.RemoteFieldSettings.IsUnknown() {
		rfc.remoteFieldSettings = &RemoteFieldSettingsConfig{}
		diags.Append(rfc.RemoteFieldSettings.As(ctx, rfc.remoteFieldSettings, basetypes.ObjectAsOptions{})...)
		if rfc.remoteFieldSettings != nil && !rfc.remoteFieldSettings.ApiClientSetup.IsNull() && !rfc.remoteFieldSettings.ApiClientSetup.IsUnknown() {
			rfc.remoteFieldSettings.apiClientSetup = &ApiClientSetupConfig{}
			diags.Append(rfc.remoteFieldSettings.ApiClientSetup.As(ctx, rfc.remoteFieldSettings.apiClientSetup, basetypes.ObjectAsOptions{})...)
		}
	}
	return diags
}

const (
	registrationFieldDefaultParentGroupID = "DEFAULT"
	registrationFieldPasswordEchoKey      = "password_echo"
)

func registrationFieldParentGroupID(plan RegFieldConfig) string {
	if !plan.ParentGroupID.IsNull() && !plan.ParentGroupID.IsUnknown() {
		if v := plan.ParentGroupID.ValueString(); v != "" {
			return v
		}
	}
	return registrationFieldDefaultParentGroupID
}

func registrationFieldOrderChangeRequested(plan, state RegFieldConfig) (int64, int64, bool) {
	if plan.Order.IsNull() || plan.Order.IsUnknown() {
		return 0, 0, false
	}
	if state.Order.IsNull() || state.Order.IsUnknown() {
		return 0, 0, false
	}
	current := plan.Order.ValueInt64()
	previous := state.Order.ValueInt64()
	if current == previous {
		return 0, 0, false
	}
	return current, previous, true
}

func registrationFieldOrderMatchesPlan(plan RegFieldConfig, actual int64) bool {
	if plan.Order.IsNull() || plan.Order.IsUnknown() {
		return true
	}
	return plan.Order.ValueInt64() == actual
}

func (r *RegFieldResource) applyRegistrationFieldOrderChange(ctx context.Context, plan RegFieldConfig, previousOrder int64) error {
	if plan.Order.IsNull() || plan.Order.IsUnknown() {
		return fmt.Errorf("order is required when reordering registration field %q", plan.FieldKey.ValueString())
	}
	currentOrder := plan.Order.ValueInt64()
	if previousOrder <= 0 || currentOrder <= 0 {
		return fmt.Errorf("registration field %q order must be greater than 0 (previous=%d, current=%d)",
			plan.FieldKey.ValueString(), previousOrder, currentOrder)
	}
	return r.cidaasClient.RegFields.UpdateOrder(ctx, cidaas.RegistrationFieldOrder{
		ParentGroupID: registrationFieldParentGroupID(plan),
		FieldKey:      plan.FieldKey.ValueString(),
		CurrentOrder:  currentOrder,
		PreviousOrder: previousOrder,
	})
}

func (r *RegFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RegFieldConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.ExtractConfigs(ctx)...)
	rfModel, diags := prepareRegFieldModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to prepare registration field model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.RegFields.Upsert(ctx, *rfModel)
	if err != nil {
		tflog.Error(ctx, "Failed to create registration field via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create registration field", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "Successfully created registration field via API", util.H{
		"field_id": res.Data.ID,
	})

	plan.ID = types.StringValue(res.Data.ID)
	if plan.Order.IsNull() || plan.Order.IsUnknown() {
		plan.Order = types.Int64Value(res.Data.Order)
	}
	plan.BaseDataType = types.StringValue(res.Data.BaseDataType)

	if rfModel.FieldDefinition != nil && rfModel.FieldDefinition.Regex != "" {
		resp.Diagnostics.Append(syncComposedRegexIntoPlan(ctx, &plan, rfModel.FieldDefinition.Regex)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(ensureFieldDefinitionRegexKnown(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !registrationFieldOrderMatchesPlan(plan, res.Data.Order) {
		regFieldOrderMutex.Lock()
		actualField, err := r.cidaasClient.RegFields.Get(ctx, plan.FieldKey.ValueString())
		if err != nil {
			regFieldOrderMutex.Unlock()
			resp.Diagnostics.AddError("failed to get actual registration field order before update", util.FormatErrorMessage(err))
			return
		}

		actualOrder := actualField.Data.Order
		if actualOrder != plan.Order.ValueInt64() {
			if err := r.applyRegistrationFieldOrderChange(ctx, plan, actualOrder); err != nil {
				regFieldOrderMutex.Unlock()
				resp.Diagnostics.AddError("failed to update registration field order", util.FormatErrorMessage(err))
				return
			}
		}

		getRes, err := r.cidaasClient.RegFields.Get(ctx, plan.FieldKey.ValueString())
		if err != nil {
			regFieldOrderMutex.Unlock()
			resp.Diagnostics.AddError("failed to read registration field after order update", util.FormatErrorMessage(err))
			return
		}
		plan.BaseDataType = types.StringValue(getRes.Data.BaseDataType)
		regFieldOrderMutex.Unlock()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "Resource registration field created successfully", util.H{
		"field_id":  res.Data.ID,
		"field_key": plan.FieldKey.ValueString(),
	})
}

func (r *RegFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RegFieldConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	res, err := r.cidaasClient.RegFields.Get(ctx, state.FieldKey.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read registration field", util.FormatErrorMessage(err))
		return
	}
	state.ID = util.StringValueOrNull(&res.Data.ID)
	// GROUPING is the only data_type with empty base_data_type from the API; others always have a value.
	if res.Data.BaseDataType == "" {
		state.BaseDataType = types.StringValue("")
	} else {
		state.BaseDataType = util.StringValueOrNull(&res.Data.BaseDataType)
	}
	state.ParentGroupID = util.StringValueOrNull(&res.Data.ParentGroupID)
	state.FieldType = util.StringValueOrNull(&res.Data.FieldType)
	state.DataType = util.StringValueOrNull(&res.Data.DataType)
	state.FieldKey = util.StringValueOrNull(&res.Data.FieldKey)
	state.Required = util.BoolValueOrNull(&res.Data.Required)
	state.Internal = util.BoolValueOrNull(&res.Data.Internal)
	state.Claimable = util.BoolValueOrNull(&res.Data.Claimable)
	state.IsSearchable = util.BoolValueOrNull(&res.Data.IsSearchable)
	state.Enabled = util.BoolValueOrNull(&res.Data.Enabled)
	state.Unique = util.BoolValueOrNull(&res.Data.Unique)
	state.OverwriteWithNullFromSocialProvider = util.BoolValueOrNull(&res.Data.OverwriteWithNullValueFromSocialProvider)
	state.ReadOnly = util.BoolValueOrNull(&res.Data.ReadOnly)
	state.IsList = util.BoolValueOrNull(&res.Data.IsList)
	state.Scopes = util.SetValueOrNull(res.Data.Scopes)
	state.ConsentRefs = util.SetValueOrNull(res.Data.ConsentRefs)
	state.Order = util.Int64ValueOrNull(&res.Data.Order)

	var localTextsObjectValues []attr.Value
	typesOfAttribute := map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}

	typesOfConsentLabel := map[string]attr.Type{
		"label":      types.StringType,
		"label_text": types.StringType,
	}

	localTextObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"locale":         types.StringType,
			"name":           types.StringType,
			"max_length_msg": types.StringType,
			"min_length_msg": types.StringType,
			"required_msg":   types.StringType,
			"match_with_msg": types.StringType,
			"attributes":     types.ListType{ElemType: types.ObjectType{AttrTypes: typesOfAttribute}},
			"consent_label":  types.ObjectType{AttrTypes: typesOfConsentLabel},
		},
	}
	for _, lt := range res.Data.LocaleTexts {
		var attributeValues []attr.Value
		for _, v := range lt.Attributes {
			attributeValue := types.ObjectValueMust(
				typesOfAttribute,
				map[string]attr.Value{
					"key":   util.StringValueOrNull(&v.Key),
					"value": util.StringValueOrNull(&v.Value),
				})
			attributeValues = append(attributeValues, attributeValue)
		}
		objValue := types.ObjectValueMust(
			localTextObjectType.AttrTypes,
			map[string]attr.Value{
				"locale":         util.StringValueOrNull(&lt.Locale),
				"name":           util.StringValueOrNull(&lt.Name),
				"max_length_msg": util.StringValueOrNull(&lt.MaxLengthErrorMsg),
				"min_length_msg": util.StringValueOrNull(&lt.MinLengthErrorMsg),
				"required_msg":   util.StringValueOrNull(&lt.RequiredMsg),
				"match_with_msg": util.StringValueOrNull(&lt.MatchWithMsg),
				"attributes": func() types.List {
					if len(lt.Attributes) == 0 {
						return types.ListNull(types.ObjectType{AttrTypes: typesOfAttribute})
					}
					return types.ListValueMust(
						types.ObjectType{AttrTypes: typesOfAttribute},
						attributeValues,
					)
				}(),
				"consent_label": func() types.Object {
					if lt.ConsentLabel == nil {
						return types.ObjectNull(typesOfConsentLabel)
					}
					return types.ObjectValueMust(
						typesOfConsentLabel,
						map[string]attr.Value{
							"label":      util.StringValueOrNull(&lt.ConsentLabel.Label),
							"label_text": util.StringValueOrNull(&lt.ConsentLabel.LabelText),
						},
					)
				}(),
			})
		localTextsObjectValues = append(localTextsObjectValues, objValue)
	}
	var diags diag.Diagnostics
	state.LocalTexts, diags = types.ListValueFrom(ctx, localTextObjectType, localTextsObjectValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if res.Data.FieldDefinition != nil {
		priorRegexes := types.ListNull(types.StringType)
		if !state.FieldDefinition.IsNull() && !state.FieldDefinition.IsUnknown() {
			var priorFD FieldDefinition
			diags := state.FieldDefinition.As(ctx, &priorFD, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !priorFD.Regexes.IsNull() && !priorFD.Regexes.IsUnknown() {
				priorRegexes = priorFD.Regexes
			}
		}
		fd, diags := types.ObjectValue(
			fieldDefinitionAttrTypes(),
			map[string]attr.Value{
				"max_length": func() basetypes.Int64Value {
					if util.Contains([]string{"TEXT", "URL"}, res.Data.DataType) {
						return types.Int64Null()
					}
					return util.Int64ValueOrNull(res.Data.FieldDefinition.MaxLength)
				}(),
				"min_length": func() basetypes.Int64Value {
					if util.Contains([]string{"TEXT", "URL"}, res.Data.DataType) {
						return types.Int64Null()
					}
					return util.Int64ValueOrNull(res.Data.FieldDefinition.MinLength)
				}(),
				"min_date":          util.TimeValueOrNull(res.Data.FieldDefinition.MinDate),
				"max_date":          util.TimeValueOrNull(res.Data.FieldDefinition.MaxDate),
				"initial_date_view": util.StringValueOrNull(&res.Data.FieldDefinition.InitialDateView),
				"initial_date":      util.TimeValueOrNull(res.Data.FieldDefinition.InitialDate),
				"regex":             util.StringValueOrNull(&res.Data.FieldDefinition.Regex),
				"regexes":           priorRegexes,
				"match_with":        util.StringValueOrNull(&res.Data.FieldDefinition.MatchWith),
			})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.FieldDefinition = fd
	}
	if res.Data.RemoteSettings != nil {
		remoteObj, diagRemote := remoteFieldSettingsToState(ctx, res.Data.RemoteSettings)
		resp.Diagnostics.Append(diagRemote...)
		if !resp.Diagnostics.HasError() {
			state.RemoteFieldSettings = remoteObj
		}
	} else {
		state.RemoteFieldSettings = types.ObjectNull(remoteFieldSettingsAttrTypes())
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// remoteFieldSettingsAttrTypes returns attribute types for remote_field_settings state object.
func remoteFieldSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"call_once": types.BoolType,
		"api_client_setup": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"communication_ep":     types.StringType,
				"http_method":          types.StringType,
				"api_access_type":      types.StringType,
				"apikey_config":        types.ObjectType{AttrTypes: map[string]attr.Type{"apikey": types.StringType, "apikey_placeholder": types.StringType, "apikey_placement": types.StringType}},
				"totp_config":          types.ObjectType{AttrTypes: map[string]attr.Type{"totpkey": types.StringType, "totp_placeholder": types.StringType, "totp_placement": types.StringType}},
				"basic_auth_config":    types.ObjectType{AttrTypes: map[string]attr.Type{"user": types.StringType, "password": types.StringType}},
				"cidaas_oauth2_config": types.ObjectType{AttrTypes: map[string]attr.Type{"client_id": types.StringType, "req_scopes": types.StringType}},
				"gen_oauth2_config":    types.ObjectType{AttrTypes: map[string]attr.Type{"client_id": types.StringType, "client_secret": types.StringType, "wellknown_url": types.StringType, "req_scopes": types.StringType}},
			},
		},
	}
}

// remoteFieldSettingsToState converts API RemoteFieldSettings to state object (sensitive values may be masked by API).
func remoteFieldSettingsToState(_ context.Context, rs *cidaas.RemoteFieldSettings) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := remoteFieldSettingsAttrTypes()
	callOnce := types.BoolNull()
	if rs.CallOnce != nil {
		callOnce = types.BoolValue(*rs.CallOnce)
	}
	apiSetupObj := types.ObjectNull(attrTypes["api_client_setup"].(types.ObjectType).AttrTypes)
	if rs.APIClientSetup != nil {
		acc := rs.APIClientSetup.ApiAccess
		apiSetupAttrTypes := attrTypes["api_client_setup"].(types.ObjectType).AttrTypes
		apiSetupAttrs := map[string]attr.Value{
			"communication_ep": types.StringValue(rs.APIClientSetup.CommunicationEP),
			"http_method":      types.StringValue(rs.APIClientSetup.HttpMethod),
			"api_access_type":  types.StringValue(acc.ApiAccessType),
			"apikey_config": objectNullOrValue(apiSetupAttrTypes["apikey_config"].(types.ObjectType).AttrTypes, acc.APIKeyDetails, func(d *cidaas.APIKeySetup) map[string]attr.Value {
				return map[string]attr.Value{"apikey": types.StringValue(d.APIKey), "apikey_placeholder": types.StringValue(d.APIKeyPlaceholder), "apikey_placement": types.StringValue(d.APIKeyPlacement)}
			}),
			"totp_config": objectNullOrValue(apiSetupAttrTypes["totp_config"].(types.ObjectType).AttrTypes, acc.TotpDetails, func(d *cidaas.TotpSetup) map[string]attr.Value {
				return map[string]attr.Value{"totpkey": types.StringValue(d.TotpKey), "totp_placeholder": types.StringValue(d.TotpPlaceholder), "totp_placement": types.StringValue(d.TotpPlacement)}
			}),
			"basic_auth_config": objectNullOrValue(apiSetupAttrTypes["basic_auth_config"].(types.ObjectType).AttrTypes, acc.BasicAuthDetails, func(d *cidaas.BasicAuthSetup) map[string]attr.Value {
				return map[string]attr.Value{"user": types.StringValue(d.User), "password": types.StringValue(d.Password)}
			}),
			"cidaas_oauth2_config": types.ObjectNull(apiSetupAttrTypes["cidaas_oauth2_config"].(types.ObjectType).AttrTypes),
			"gen_oauth2_config":    types.ObjectNull(apiSetupAttrTypes["gen_oauth2_config"].(types.ObjectType).AttrTypes),
		}
		if acc.OAuth2Details != nil {
			if acc.OAuth2Details.WellknownURL != "" {
				apiSetupAttrs["gen_oauth2_config"] = types.ObjectValueMust(apiSetupAttrTypes["gen_oauth2_config"].(types.ObjectType).AttrTypes, map[string]attr.Value{
					"client_id": types.StringValue(acc.OAuth2Details.ClientID), "client_secret": types.StringValue(acc.OAuth2Details.ClientSecret),
					"wellknown_url": types.StringValue(acc.OAuth2Details.WellknownURL), "req_scopes": types.StringValue(acc.OAuth2Details.ReqScopes),
				})
			} else {
				apiSetupAttrs["cidaas_oauth2_config"] = types.ObjectValueMust(apiSetupAttrTypes["cidaas_oauth2_config"].(types.ObjectType).AttrTypes, map[string]attr.Value{
					"client_id": types.StringValue(acc.OAuth2Details.ClientID), "req_scopes": types.StringValue(acc.OAuth2Details.ReqScopes),
				})
			}
		}
		var d diag.Diagnostics
		apiSetupObj, d = types.ObjectValue(apiSetupAttrTypes, apiSetupAttrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(attrTypes), diags
		}
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"call_once":        callOnce,
		"api_client_setup": apiSetupObj,
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), diags
	}
	return obj, diags
}

func objectNullOrValue[T any](attrTypes map[string]attr.Type, ptr *T, fn func(*T) map[string]attr.Value) types.Object {
	if ptr == nil {
		return types.ObjectNull(attrTypes)
	}
	return types.ObjectValueMust(attrTypes, fn(ptr))
}

func (r *RegFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl
	var plan, state RegFieldConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.ExtractConfigs(ctx)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	fieldModel, diags := prepareRegFieldModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare registration field model for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	fieldModel.ID = state.ID.ValueString()

	if _, _, ok := registrationFieldOrderChangeRequested(plan, state); ok {
		regFieldOrderMutex.Lock()
		// Fetch the latest registration field details from the server to get the actual current order.
		// The stored state order can be stale if other fields were reordered in the same apply run.
		actualField, err := r.cidaasClient.RegFields.Get(ctx, plan.FieldKey.ValueString())
		if err != nil {
			regFieldOrderMutex.Unlock()
			resp.Diagnostics.AddError("failed to get actual registration field order before update", util.FormatErrorMessage(err))
			return
		}

		actualOrder := actualField.Data.Order
		if actualOrder != plan.Order.ValueInt64() {
			if err := r.applyRegistrationFieldOrderChange(ctx, plan, actualOrder); err != nil {
				regFieldOrderMutex.Unlock()
				resp.Diagnostics.AddError("failed to update registration field order", util.FormatErrorMessage(err))
				return
			}
		}
		regFieldOrderMutex.Unlock()
	}

	res, err := r.cidaasClient.RegFields.Upsert(ctx, *fieldModel)
	if err != nil {
		tflog.Error(ctx, "failed to update registration field via API", util.H{
			"field_id": state.ID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to update registration field", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated registration field via API", util.H{
		"field_id": state.ID.ValueString(),
	})

	if plan.Order.IsNull() || plan.Order.IsUnknown() {
		plan.Order = types.Int64Value(res.Data.Order)
	}

	if fieldModel.FieldDefinition != nil && fieldModel.FieldDefinition.Regex != "" {
		resp.Diagnostics.Append(syncComposedRegexIntoPlan(ctx, &plan, fieldModel.FieldDefinition.Regex)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(ensureFieldDefinitionRegexKnown(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Computed base_data_type must be known after apply (e.g. GROUPING has empty). Fallback to state or "".
	if plan.BaseDataType.IsUnknown() {
		if !state.BaseDataType.IsUnknown() {
			plan.BaseDataType = state.BaseDataType
		} else {
			plan.BaseDataType = types.StringValue("")
		}
	}
	if plan.BaseDataType.IsUnknown() {
		plan.BaseDataType = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource registration field updated successfully", util.H{
		"field_id":  state.ID.ValueString(),
		"field_key": plan.FieldKey.ValueString(),
	})
}

func (r *RegFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state RegFieldConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.RegFields.Delete(ctx, state.FieldKey.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete registration field via API", util.H{
			"field_key": state.FieldKey.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete registration field", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource registration field deleted successfully", util.H{
		"field_key": state.FieldKey.ValueString(),
	})
}

func (r *RegFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("field_key"), req, resp)
}

func prepareRegFieldModel(ctx context.Context, plan RegFieldConfig) (*cidaas.RegistrationFieldConfig, diag.Diagnostics) { //nolint:gocognit
	var regConfig cidaas.RegistrationFieldConfig
	regConfig.Internal = plan.Internal.ValueBool()
	regConfig.ReadOnly = plan.ReadOnly.ValueBool()
	regConfig.Claimable = plan.Claimable.ValueBool()
	regConfig.Required = plan.Required.ValueBool()
	regConfig.Unique = plan.Unique.ValueBool()
	regConfig.IsSearchable = plan.IsSearchable.ValueBool()
	regConfig.OverwriteWithNullValueFromSocialProvider = plan.OverwriteWithNullFromSocialProvider.ValueBool()
	regConfig.Enabled = plan.Enabled.ValueBool()
	regConfig.IsList = plan.IsList.ValueBool()
	regConfig.ParentGroupID = plan.ParentGroupID.ValueString()
	regConfig.FieldType = plan.FieldType.ValueString()
	regConfig.FieldKey = plan.FieldKey.ValueString()
	regConfig.DataType = plan.DataType.ValueString()
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		regConfig.Order = plan.Order.ValueInt64()
	}

	className := "FieldSetup"
	if regConfig.FieldType == "SYSTEM" {
		className = "de.cidaas.core.db.RegistrationFieldSetup"
	}
	regConfig.ClassName = className

	diag := plan.Scopes.ElementsAs(ctx, &regConfig.Scopes, false)
	if diag.HasError() {
		return nil, diag
	}

	if plan.DataType.ValueString() == "CONSENT" {
		diag = plan.ConsentRefs.ElementsAs(ctx, &regConfig.ConsentRefs, false)
		if diag.HasError() {
			return nil, diag
		}
	}

	var attrKeys []string
	setLocalTexts := func(source []*LocalTexts, target *[]*cidaas.LocaleText) {
		for _, s := range source {
			tempLocalText := &cidaas.LocaleText{
				Locale:            s.Locale.ValueString(),
				Name:              s.Name.ValueString(),
				MaxLengthErrorMsg: s.MaxLengthMsg.ValueString(),
				MinLengthErrorMsg: s.MinLengthMsg.ValueString(),
				RequiredMsg:       s.RequiredMsg.ValueString(),
				MatchWithMsg:      s.MatchWithMsg.ValueString(),
			}
			cidaasAttribues := []*cidaas.Attribute{}

			for _, v := range s.attributes {
				cidaasAttribues = append(cidaasAttribues, &cidaas.Attribute{
					Key:   v.Key.ValueString(),
					Value: v.Value.ValueString(),
				})
				attrKeys = append(attrKeys, v.Key.ValueString())
			}
			if len(s.attributes) > 0 {
				tempLocalText.Attributes = cidaasAttribues
			}
			if !s.ConsentLabel.IsNull() && !s.ConsentLabel.IsUnknown() {
				tempLocalText.ConsentLabel = &cidaas.ConsentLabel{
					Label:     s.consent.Label.ValueString(),
					LabelText: s.consent.LabelText.ValueString(),
				}
			}
			*target = append(*target, tempLocalText)
		}
	}
	if !plan.LocalTexts.IsNull() && !plan.LocalTexts.IsUnknown() && len(plan.localTexts) > 0 {
		setLocalTexts(plan.localTexts, &regConfig.LocaleTexts)
	}

	if !plan.FieldDefinition.IsNull() {
		regexValue := ""
		if plan.fieldDefinition != nil && !plan.fieldDefinition.Regex.IsNull() && !plan.fieldDefinition.Regex.IsUnknown() {
			regexValue = plan.fieldDefinition.Regex.ValueString()
		}
		if plan.fieldDefinition != nil && !plan.fieldDefinition.Regexes.IsNull() && !plan.fieldDefinition.Regexes.IsUnknown() {
			var patterns []string
			diags := plan.fieldDefinition.Regexes.ElementsAs(ctx, &patterns, false)
			diag.Append(diags...)
			if diag.HasError() {
				return nil, diag
			}
			if len(patterns) > 0 {
				composed, err := composeANDRegexes(patterns)
				if err != nil {
					diag.AddError("Validation Error", fmt.Sprintf("field_definition.regexes: %s", err.Error()))
					return nil, diag
				}
				regexValue = composed
			}
		}
		regConfig.FieldDefinition = &cidaas.FieldDefinition{
			MinLength:       plan.fieldDefinition.MinLength.ValueInt64Pointer(),
			MaxLength:       plan.fieldDefinition.MaxLength.ValueInt64Pointer(),
			InitialDateView: plan.fieldDefinition.InitialDateView.ValueString(),
			Regex:           regexValue,
			MatchWith:       plan.fieldDefinition.MatchWith.ValueString(),
		}
		if len(attrKeys) > 0 {
			regConfig.FieldDefinition.AttributesKeys = attrKeys
		}
		if !plan.fieldDefinition.MinDate.IsNull() {
			minDate, err := time.Parse(layout, plan.fieldDefinition.MinDate.ValueString())
			if err != nil {
				diag.AddError("Parse Error", "failed to parse min_date configured")
			}
			regConfig.FieldDefinition.MinDate = &minDate
		}
		if !plan.fieldDefinition.MaxDate.IsNull() {
			maxDate, err := time.Parse(layout, plan.fieldDefinition.MaxDate.ValueString())
			if err != nil {
				diag.AddError("Parse Error", "failed to parse max_date configured")
			}
			regConfig.FieldDefinition.MaxDate = &maxDate
		}
		if !plan.fieldDefinition.InitialDate.IsNull() {
			initialDate, err := time.Parse(layout, plan.fieldDefinition.InitialDate.ValueString())
			if err != nil {
				diag.AddError("Parse Error", "failed to parse initial_date configured")
			}
			regConfig.FieldDefinition.InitialDate = &initialDate
		}
	}
	if plan.remoteFieldSettings != nil && plan.remoteFieldSettings.apiClientSetup != nil {
		remote, diagRemote := buildRemoteFieldSettings(ctx, plan.remoteFieldSettings)
		if diagRemote.HasError() {
			return nil, diagRemote
		}
		regConfig.RemoteSettings = remote
	}
	return &regConfig, nil
}

// buildRemoteFieldSettings converts Terraform remote_field_settings config to API type.
func buildRemoteFieldSettings(ctx context.Context, cfg *RemoteFieldSettingsConfig) (*cidaas.RemoteFieldSettings, diag.Diagnostics) {
	var diags diag.Diagnostics
	setup := cfg.apiClientSetup
	apiSetup := &cidaas.ApiClientSetup{
		CommunicationEP: setup.CommunicationEP.ValueString(),
		HttpMethod:      setup.HTTPMethod.ValueString(),
		ApiAccess: cidaas.APIAccessSetup{
			ApiAccessType: setup.APIAccessType.ValueString(),
		},
	}
	// Map the appropriate auth config block to ApiAccess.
	if !setup.ApikeyConfig.IsNull() && !setup.ApikeyConfig.IsUnknown() {
		var ac ApikeyConfig
		diags.Append(setup.ApikeyConfig.As(ctx, &ac, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			apiSetup.ApiAccess.APIKeyDetails = &cidaas.APIKeySetup{
				APIKey:            ac.Apikey.ValueString(),
				APIKeyPlaceholder: ac.ApikeyPlaceholder.ValueString(),
				APIKeyPlacement:   ac.ApikeyPlacement.ValueString(),
			}
		}
	}
	if !setup.TotpConfig.IsNull() && !setup.TotpConfig.IsUnknown() {
		var tc TotpConfig
		diags.Append(setup.TotpConfig.As(ctx, &tc, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			apiSetup.ApiAccess.TotpDetails = &cidaas.TotpSetup{
				TotpKey:         tc.TotpKey.ValueString(),
				TotpPlaceholder: tc.TotpPlaceholder.ValueString(),
				TotpPlacement:   tc.TotpPlacement.ValueString(),
			}
		}
	}
	if !setup.BasicAuthConfig.IsNull() && !setup.BasicAuthConfig.IsUnknown() {
		var bc BasicAuthConfig
		diags.Append(setup.BasicAuthConfig.As(ctx, &bc, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			apiSetup.ApiAccess.BasicAuthDetails = &cidaas.BasicAuthSetup{
				User:     bc.User.ValueString(),
				Password: bc.Password.ValueString(),
			}
		}
	}
	if !setup.CidaasOAuth2Config.IsNull() && !setup.CidaasOAuth2Config.IsUnknown() {
		var oc CidaasOAuth2Config
		diags.Append(setup.CidaasOAuth2Config.As(ctx, &oc, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			apiSetup.ApiAccess.OAuth2Details = &cidaas.OAuth2Setup{
				ClientID:  oc.ClientID.ValueString(),
				ReqScopes: oc.ReqScopes.ValueString(),
			}
		}
	}
	if !setup.GenOAuth2Config.IsNull() && !setup.GenOAuth2Config.IsUnknown() {
		var gc GenOAuth2Config
		diags.Append(setup.GenOAuth2Config.As(ctx, &gc, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			apiSetup.ApiAccess.OAuth2Details = &cidaas.OAuth2Setup{
				ClientID:     gc.ClientID.ValueString(),
				ClientSecret: gc.ClientSecret.ValueString(),
				WellknownURL: gc.WellknownURL.ValueString(),
				ReqScopes:    gc.ReqScopes.ValueString(),
			}
		}
	}
	out := &cidaas.RemoteFieldSettings{
		APIClientSetup: apiSetup,
	}
	if !cfg.CallOnce.IsNull() && !cfg.CallOnce.IsUnknown() {
		callOnce := cfg.CallOnce.ValueBool()
		out.CallOnce = &callOnce
	}
	return out, diags
}

// custom validations
var (
	_ validator.Bool      = validateIsRequiredMsgAvailable{}
	_ validator.Int64     = validateIsMaxMinMsgAvailable{}
	_ planmodifier.String = fieldTypeModifier{}
	_ validator.String    = dateTypeValidator{}
	_ validator.String    = dateValidator{}
	_ validator.String    = dataTypeValidator{}
	_ validator.String    = validateIsMaxMinMsgAvailableForRegex{}
	_ validator.List      = validateIsMaxMinMsgAvailableForRegexes{}
	_ validator.String    = validateRegexExclusive{}
	_ validator.List      = validateRegexesExclusive{}
	_ validator.String    = validateMatchWith{}
	_ validator.String    = validateMatchWithMsg{}
)

type (
	validateIsRequiredMsgAvailable         struct{}
	validateIsMaxMinMsgAvailable           struct{}
	fieldTypeModifier                      struct{}
	dateTypeValidator                      struct{}
	dateValidator                          struct{}
	dataTypeValidator                      struct{}
	validateIsMaxMinMsgAvailableForRegex   struct{}
	validateIsMaxMinMsgAvailableForRegexes struct{}
	validateRegexExclusive                 struct{}
	validateRegexesExclusive               struct{}
	validateMatchWith                      struct{}
	validateMatchWithMsg                   struct{}
)

func ensureFieldDefinitionRegexKnown(ctx context.Context, plan *RegFieldConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if plan.fieldDefinition == nil {
		return diags
	}
	if fieldDefinitionHasRegexes(plan.fieldDefinition) {
		return diags
	}
	if !plan.fieldDefinition.Regex.IsUnknown() {
		return diags
	}
	plan.fieldDefinition.Regex = types.StringNull()
	obj, d := types.ObjectValueFrom(ctx, fieldDefinitionAttrTypes(), plan.fieldDefinition)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.FieldDefinition = obj
	return diags
}

func syncComposedRegexIntoPlan(ctx context.Context, plan *RegFieldConfig, regex string) diag.Diagnostics {
	var diags diag.Diagnostics
	if plan.fieldDefinition == nil || regex == "" {
		return diags
	}
	if plan.fieldDefinition.Regexes.IsNull() || plan.fieldDefinition.Regexes.IsUnknown() || len(plan.fieldDefinition.Regexes.Elements()) == 0 {
		return diags
	}
	plan.fieldDefinition.Regex = types.StringValue(regex)
	obj, d := types.ObjectValueFrom(ctx, fieldDefinitionAttrTypes(), plan.fieldDefinition)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.FieldDefinition = obj
	return diags
}

func fieldDefinitionHasRegex(fd *FieldDefinition) bool {
	return fd != nil && !fd.Regex.IsNull() && !fd.Regex.IsUnknown() && fd.Regex.ValueString() != ""
}

func fieldDefinitionHasRegexes(fd *FieldDefinition) bool {
	return fd != nil && !fd.Regexes.IsNull() && !fd.Regexes.IsUnknown() && len(fd.Regexes.Elements()) > 0
}

func validateRegexDataTypeAndMessages(_ context.Context, config RegFieldConfig, attrPath string, respDiags *diag.Diagnostics) {
	if !util.Contains([]string{"TEXT", "URL"}, config.DataType.ValueString()) {
		respDiags.AddError(
			"Validation Error",
			fmt.Sprintf("The attribute %s is only allowed when data_type is TEXT or URL", attrPath),
		)
		return
	}
	for _, v := range config.localTexts {
		if v.MinLengthMsg.IsNull() || v.MinLengthMsg.ValueString() == "" {
			respDiags.AddError(
				"Validation Error",
				fmt.Sprintf("The attribute local_texts.min_length_msg can not be empty when %s is set", attrPath),
			)
			return
		}
		if v.MaxLengthMsg.IsNull() || v.MaxLengthMsg.ValueString() == "" {
			respDiags.AddError(
				"Validation Error",
				fmt.Sprintf("The attribute local_texts.max_length_msg can not be empty when %s is set", attrPath),
			)
			return
		}
	}
}

func (v validateIsMaxMinMsgAvailableForRegex) Description(_ context.Context) string {
	return "Checks min_length_msg and max_length_msg when field_definition.regex is set"
}

func (v validateIsMaxMinMsgAvailableForRegex) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateIsMaxMinMsgAvailableForRegex) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateRegexDataTypeAndMessages(ctx, config, req.Path.String(), &resp.Diagnostics)
}

func (v validateIsMaxMinMsgAvailableForRegexes) Description(_ context.Context) string {
	return "Checks min_length_msg and max_length_msg when field_definition.regexes is set"
}

func (v validateIsMaxMinMsgAvailableForRegexes) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateIsMaxMinMsgAvailableForRegexes) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateRegexDataTypeAndMessages(ctx, config, req.Path.String(), &resp.Diagnostics)
}

func (v validateRegexExclusive) Description(_ context.Context) string {
	return "regex and regexes are mutually exclusive"
}

func (v validateRegexExclusive) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateRegexExclusive) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if fieldDefinitionHasRegexes(config.fieldDefinition) {
		resp.Diagnostics.AddError(
			"Validation Error",
			"field_definition.regex and field_definition.regexes cannot be set together; use one or the other",
		)
	}
}

func (v validateRegexesExclusive) Description(_ context.Context) string {
	return "regex and regexes are mutually exclusive"
}

func (v validateRegexesExclusive) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateRegexesExclusive) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if fieldDefinitionHasRegex(config.fieldDefinition) {
		resp.Diagnostics.AddError(
			"Validation Error",
			"field_definition.regex and field_definition.regexes cannot be set together; use one or the other",
		)
	}
}

func (v validateIsRequiredMsgAvailable) Description(_ context.Context) string {
	return "required_msg must be set when required is true"
}

func (v validateIsRequiredMsgAvailable) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateIsRequiredMsgAvailable) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueBool() && len(config.localTexts) > 0 {
		for _, v := range config.localTexts {
			if v.RequiredMsg.IsNull() || v.RequiredMsg.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Validation Error",
					fmt.Sprintf("The attribute local_texts.required_msg is required when %s is set to true", req.Path.String()),
				)
				return
			}
		}
	}
}

func (v validateIsMaxMinMsgAvailable) Description(_ context.Context) string {
	return "max_length & min_length validation"
}

func (v validateIsMaxMinMsgAvailable) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateIsMaxMinMsgAvailable) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) { //nolint:gocognit
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)

	if !req.ConfigValue.IsNull() {
		if req.ConfigValue.ValueInt64() < 1 {
			resp.Diagnostics.AddError(
				"Validation Error",
				fmt.Sprintf("The attribute %s must be greater than 0", req.Path.String()),
			)
			return
		}
		if req.Path.String() == "field_definition.min_length" {
			for _, v := range config.localTexts {
				if v.MinLengthMsg.IsNull() || v.MinLengthMsg.ValueString() == "" {
					resp.Diagnostics.AddError(
						"Validation Error",
						fmt.Sprintf("The attribute local_texts.min_length_msg can not be empty when %s is set", req.Path.String()),
					)
					return
				}
			}

			if config.fieldDefinition.MaxLength.IsNull() || config.fieldDefinition.MaxLength.ValueInt64() <= 0 {
				resp.Diagnostics.AddError(
					"Validation Error",
					fmt.Sprintf("The attribute field_definition.max_length can not be empty when %s is set", req.Path.String()),
				)
				return
			}

			if config.fieldDefinition.MaxLength.ValueInt64() < config.fieldDefinition.MinLength.ValueInt64() {
				resp.Diagnostics.AddError(
					"Validation Error",
					fmt.Sprintf("The attribute field_definition.max_length can not be less than %s", req.Path.String()),
				)
				return
			}
		}
		if req.Path.String() == "field_definition.max_length" {
			for _, v := range config.localTexts {
				if v.MaxLengthMsg.IsNull() || v.MaxLengthMsg.ValueString() == "" {
					resp.Diagnostics.AddError(
						"Validation Error",
						fmt.Sprintf("The attribute local_texts.max_length_msg can not be empty when %s is set", req.Path.String()),
					)
					return
				}
			}
		}
	}
}

func (v fieldTypeModifier) Description(_ context.Context) string {
	return "Checks if field_type is SYSTEM while creating a field"
}

func (v fieldTypeModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v fieldTypeModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() && req.ConfigValue.Equal(types.StringValue("SYSTEM")) {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			"field with SYSYTEM field_type cannot be created. SYSTEM fields can only be updated. To update an existing field please import first",
		)
	}
}

func (v dateTypeValidator) Description(_ context.Context) string {
	return "Checks min_date, max_date, initial_date_view and initiate_date"
}

func (v dateTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dateTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !req.ConfigValue.IsNull() {
		var config RegFieldConfig
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if config.DataType.ValueString() != "DATE" {
			resp.Diagnostics.AddError(
				"Validation Error",
				fmt.Sprintf("The attribute %s is only allowed when data_type is DATE", req.Path.String()),
			)
			return
		}
	}
}

func (v dateValidator) Description(_ context.Context) string {
	return "Validates that the value is a valid date."
}

func (v dateValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v dateValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	_, err := time.Parse(layout, req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Date Format",
			fmt.Sprintf("Attribute %s expected to be a valid ISO 8601 date in the format %s.", req.Path.String(), layout),
		)
	}
}

func (v dataTypeValidator) Description(_ context.Context) string {
	return "Validates that the value is a valid date."
}

func (v dataTypeValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v dataTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)

	if req.ConfigValue.ValueString() == "DATE" &&
		(config.FieldDefinition.IsNull() ||
			config.fieldDefinition.MinDate.IsNull() ||
			config.fieldDefinition.MaxDate.IsNull() ||
			config.fieldDefinition.InitialDate.IsNull()) {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			"Attributes min_date, max_date, initial_date and initial_date_view can not be empty when data_type is DATE.",
		)
	}

	if req.ConfigValue.ValueString() != "CONSENT" &&
		!config.ConsentRefs.IsNull() {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			"Attribute consent_refs is only allowed when data_type is set to CONSENT.",
		)
	}

	attrKeysRequiredDataTypes := []string{"SELECT", "RADIO", "MULTISELECT"}
	if util.Contains(attrKeysRequiredDataTypes, req.ConfigValue.ValueString()) {
		for _, s := range config.localTexts {
			if !s.Attributes.IsNull() && !s.Attributes.IsUnknown() && len(s.attributes) == 0 {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attributes local_texts[i].attributes can not be empty when data_type is %s.", req.ConfigValue.ValueString()),
				)
			}
		}
	}

	noMaxMinLengthDataTypes := []string{"CHECKBOX", "CONSENT", "JSON_STRING", "ARRAY", "NUMBER", "SELECT", "RADIO", "MULTISELECT", "MOBILE", "JSON_STRING", "TEXT", "URL"}
	if util.Contains(noMaxMinLengthDataTypes, req.ConfigValue.ValueString()) {
		if config.FieldDefinition.IsNull() {
			return
		}
		if !config.fieldDefinition.MinLength.IsNull() || !config.fieldDefinition.MaxLength.IsNull() {
			resp.Diagnostics.AddError(
				"Unexpected Resource Configuration",
				fmt.Sprintf("Attributes min_length, max_length are not allowed in config when the data_type is %s.", req.ConfigValue.ValueString()),
			)
		}
	}

	noAttributesDataTypes := []string{
		"TEXT", "NUMBER", "CHECKBOX", "PASSWORD", "DATE", "URL", "EMAIL",
		"TEXTAREA", "MOBILE", "CONSENT", "JSON_STRING", "USERNAME", "ARRAY", "GROUPING", "DAYDATE",
	}
	if util.Contains(noAttributesDataTypes, req.ConfigValue.ValueString()) {
		for _, v := range config.localTexts {
			if len(v.attributes) > 0 {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("param local_texts[i].attributes not allowed in config when the data_type is %s.", req.ConfigValue.ValueString()),
				)
			}
		}
	}

	if config.FieldKey.ValueString() != registrationFieldPasswordEchoKey {
		if !config.FieldDefinition.IsNull() && !config.FieldDefinition.IsUnknown() &&
			!config.fieldDefinition.MatchWith.IsNull() && config.fieldDefinition.MatchWith.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Unexpected Resource Configuration",
				"Attribute field_definition.match_with is only allowed when field_key is password_echo.",
			)
		}
		for _, lt := range config.localTexts {
			if !lt.MatchWithMsg.IsNull() && lt.MatchWithMsg.ValueString() != "" {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					"Attribute local_texts.match_with_msg is only allowed when field_key is password_echo.",
				)
				return
			}
		}
	}
}

func (v validateMatchWith) Description(_ context.Context) string {
	return "match_with is only allowed when field_key is password_echo"
}

func (v validateMatchWith) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateMatchWith) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)
	if config.FieldKey.ValueString() != registrationFieldPasswordEchoKey {
		resp.Diagnostics.AddError(
			"Validation Error",
			fmt.Sprintf("The attribute %s is only allowed when field_key is password_echo", req.Path.String()),
		)
		return
	}

	for _, lt := range config.localTexts {
		if lt.MatchWithMsg.IsNull() || lt.MatchWithMsg.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Validation Error",
				"The attribute local_texts.match_with_msg can not be empty when field_definition.match_with is set",
			)
			return
		}
	}
}

func (v validateMatchWithMsg) Description(_ context.Context) string {
	return "match_with_msg is only allowed when field_key is password_echo and required when match_with is set"
}

func (v validateMatchWithMsg) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validateMatchWithMsg) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var config RegFieldConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractConfigs(ctx)...)

	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" &&
		config.FieldKey.ValueString() != registrationFieldPasswordEchoKey {
		resp.Diagnostics.AddError(
			"Validation Error",
			fmt.Sprintf("The attribute %s is only allowed when field_key is password_echo", req.Path.String()),
		)
		return
	}

	if config.FieldDefinition.IsNull() || config.FieldDefinition.IsUnknown() || config.fieldDefinition == nil {
		return
	}
	if config.fieldDefinition.MatchWith.IsNull() || config.fieldDefinition.MatchWith.ValueString() == "" {
		return
	}

	if req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Validation Error",
			"The attribute local_texts.match_with_msg can not be empty when field_definition.match_with is set",
		)
	}
}
