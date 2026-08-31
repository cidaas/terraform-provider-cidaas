package consent

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	SCOPES = "SCOPES"
	URL    = "URL"
)

type ConsentVersionResource struct {
	base.BaseResource
	cidaasClient *cidaas.Client
}

func NewConsentVersionResource() resource.Resource {
	return &ConsentVersionResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_CONSENT_VERSION,
				Schema: &consentversionSchema,
			},
		),
	}
}

type ConsentVersionConfig struct {
	ID             types.String  `tfsdk:"id"`
	Version        types.Float64 `tfsdk:"version"`
	ConsentID      types.String  `tfsdk:"consent_id"`
	ConsentType    types.String  `tfsdk:"consent_type"`
	Scopes         types.Set     `tfsdk:"scopes"`
	RequiredFields types.Set     `tfsdk:"required_fields"`
	ConsentLocales types.Set     `tfsdk:"consent_locales"`

	consentLocale []*ConsentLocale
}

type ConsentLocale struct {
	Content types.String `tfsdk:"content"`
	Locale  types.String `tfsdk:"locale"`
	URL     types.String `tfsdk:"url"`
}

func (t *ConsentVersionConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !t.ConsentLocales.IsNull() && !t.ConsentLocales.IsUnknown() {
		diags.Append(t.ConsentLocales.ElementsAs(ctx, &t.consentLocale, false)...)
	}
	return diags
}

func (r *ConsentVersionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	var config ConsentVersionConfig
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	res.Diagnostics.Append(config.extract(ctx)...)

	localeMap := make(map[string]bool)

	if !config.ConsentLocales.IsNull() && !config.ConsentLocales.IsUnknown() && len(config.consentLocale) > 0 {
		for _, loc := range config.consentLocale {
			locale := loc.Locale.ValueString()
			if _, exists := localeMap[locale]; exists {
				res.Diagnostics.AddError("Duplicate locale not allowed", fmt.Sprintf("Duplicate locale '%s' found in consent_locales", locale))
			}
			if config.ConsentType.ValueString() == SCOPES && !loc.URL.IsNull() && !loc.URL.IsUnknown() {
				res.Diagnostics.AddError("Unsupported attribute", "attribute 'consent_locales.url' not supported when consent_type is 'SCOPES'")
			}
			if config.ConsentType.ValueString() == URL && (loc.URL.IsNull() || loc.URL.ValueString() == "") {
				res.Diagnostics.AddError("Missing required attribute", "attribute 'consent_locales.url' is required or can't be empty when consent_type is 'URL'")
			}
			localeMap[locale] = true
		}
	}
	if config.ConsentType.ValueString() == SCOPES && config.Scopes.IsNull() {
		res.Diagnostics.AddError("Missing required attribute", "attribute 'scopes' is required when consent_type is 'SCOPES'")
	}
	if config.ConsentType.ValueString() == SCOPES && config.RequiredFields.IsNull() {
		res.Diagnostics.AddError("Missing required attribute", "attribute 'required_fields' is required when consent_type is 'SCOPES'")
	}
	if config.ConsentType.ValueString() == URL && !config.Scopes.IsNull() && !config.Scopes.IsUnknown() {
		res.Diagnostics.AddError("Unsupported attribute", "attribute 'scopes' not supported when consent_type is 'URL'")
	}
	if config.ConsentType.ValueString() == URL && !config.RequiredFields.IsNull() && !config.RequiredFields.IsUnknown() {
		res.Diagnostics.AddError("Unsupported attribute", "attribute 'required_fields' not supported when consent_type is 'URL'")
	}
}

var consentversionSchema = schema.Schema{
	MarkdownDescription: "The Consent Version resource in the provider allows you to manage different versions of a specific consent in Cidaas." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:tenant_consent_read" +
		"\n- cidaas:tenant_consent_write",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier of the consent version resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"version": schema.Float64Attribute{
			Required:            true,
			MarkdownDescription: "The version number of the consent. It can not be updated for a specific consent version.",
			PlanModifiers: []planmodifier.Float64{
				validators.ImmutableInt64Identifier{},
			},
		},
		"consent_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The `consent_id` to which the consent version belongs.",
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"consent_type": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The consent_type defines whether consent is URL or SCOPES. Allowed values are `URL`, `SCOPES`.",
			Validators: []validator.String{
				stringvalidator.OneOf(SCOPES, URL),
			},
		},
		"scopes": schema.SetAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Set of scopes associated with the consent version. Required when `consent_type` is `SCOPES`.",
		},
		"required_fields": schema.SetAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Set of required fields associated with the consent version. Required when `consent_type` is `SCOPES`.",
		},
		"consent_locales": schema.SetNestedAttribute{
			Required:            true,
			MarkdownDescription: "Set of locales for the consent version.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"content": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The content of the consent for the specified locale.",
					},
					"locale": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The locale tag (e.g. `en-US`, `de-DE`).",
						Validators: []validator.String{
							stringvalidator.OneOf(util.AllowedBCP47Locales...),
						},
					},
					"url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The URL associated with the consent for the specified locale. Required when `consent_type` is `URL`.",
					},
				},
			},
		},
	},
}

func (r *ConsentVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	if c.Capabilities.TargetVersion == "4.x" {
		resp.Diagnostics.AddError(
			"Consent Resource Coming Soon on v4",
			"Consent resources (`cidaas_consent`, `cidaas_consent_group`, `cidaas_consent_version`) are supported on cidaas v3.x. Support for cidaas v4.x (Trustdesk) is coming soon. Please set cidaas_version = \"3.x\" in your provider block to manage v3 consent resources.",
		)
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
	r.cidaasClient = c.CidaasClient
}

func (r *ConsentVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConsentVersionConfig

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	var scopes, requiredFields []string
	if plan.ConsentType.ValueString() == SCOPES {
		resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
		resp.Diagnostics.Append(plan.RequiredFields.ElementsAs(ctx, &requiredFields, false)...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "failed to get scopes or required fields", util.H{
				"errors": resp.Diagnostics.Errors(),
			})
			return
		}
	}

	consentVersion := cidaas.ConsentVersionModel{
		Version:        plan.Version.ValueFloat64(),
		ConsentID:      plan.ConsentID.ValueString(),
		ConsentType:    plan.ConsentType.ValueString(),
		Scopes:         scopes,
		RequiredFields: requiredFields,
	}

	res, err := r.cidaasClient.ConsentVersion.Upsert(ctx, consentVersion)
	if err != nil {
		tflog.Error(ctx, "failed to create consent version via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create consent version", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "successfully created consent version via API", util.H{
		"consent_version_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	plan.Version = types.Float64Value(res.Data.Version)

	for _, pcl := range plan.consentLocale {
		consentLocal := cidaas.ConsentLocalModel{
			ConsentID:        plan.ConsentID.ValueString(),
			ConsentVersionID: res.Data.ID,
			Content:          pcl.Content.ValueString(),
			Locale:           pcl.Locale.ValueString(),
		}
		if plan.ConsentType.ValueString() == URL {
			consentLocal.URL = pcl.URL.ValueString()
		}
		_, err := r.cidaasClient.ConsentVersion.UpsertLocal(ctx, consentLocal)
		if err != nil {
			tflog.Error(ctx, "failed to create consent locale via API", util.H{
				"locale": pcl.Locale.ValueString(),
				"error":  err.Error(),
			})
			resp.Diagnostics.AddError("failed to create consent locale", fmt.Sprintf("Error: %s", err.Error()))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource consent version created successfully")
}

func (r *ConsentVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConsentVersionConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(state.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.cidaasClient.ConsentVersion.Get(ctx, state.ConsentID.ValueString())
	if err != nil {
		if base.ReadHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read consent version via API", util.H{
			"consent_id": state.ConsentID.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to read consent version", fmt.Sprintf("Error: %s ", err.Error()))
		return
	}

	if res.Success && res.Status == http.StatusOK && len(res.Data) == 0 {
		tflog.Error(ctx, "no consent versions found for consent ID", util.H{
			"consent_id": state.ConsentID.ValueString(),
			"status":     res.Status,
		})
		resp.Diagnostics.AddError("Invalid consent_id", fmt.Sprintf("No consent version found for the provided consent_id %+v", state.ConsentID.String()))
		return
	}

	tflog.Debug(ctx, "processing consent versions")

	isAvailable := false
	if len(res.Data) > 0 {
		for _, version := range res.Data {
			if version.ID == state.ID.ValueString() {
				isAvailable = true
				consentType := version.ConsentType
				state.Version = types.Float64Value(version.Version)
				state.ConsentType = util.StringValueOrNull(&consentType)
				break
			}
		}
	}

	var diag diag.Diagnostics
	var objectValues []attr.Value
	consentLocalObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"content": types.StringType,
			"locale":  types.StringType,
			"url":     types.StringType,
		},
	}

	tflog.Debug(ctx, "processing consent locales")

	for i, cl := range state.consentLocale {
		res, err := r.cidaasClient.ConsentVersion.GetLocal(ctx, state.ID.ValueString(), cl.Locale.ValueString())
		if err != nil {
			tflog.Error(ctx, "failed to read consent version locale via API", util.H{
				"locale_index":       i,
				"locale":             cl.Locale.ValueString(),
				"consent_version_id": state.ID.ValueString(),
				"error":              err.Error(),
			})
			resp.Diagnostics.AddError("Failed to read consent version locale", fmt.Sprintf("Error: %s ", err.Error()))
			return
		}
		if !res.Success && res.Status == http.StatusNoContent {
			tflog.Error(ctx, "consent version locale not found", util.H{
				"locale_index":       i,
				"locale":             cl.Locale.ValueString(),
				"consent_version_id": state.ID.ValueString(),
				"status":             res.Status,
			})
			resp.Diagnostics.AddError(
				"Consent Version Local not found",
				fmt.Sprintf("No consent version locale found for the combination of consent_version_id %s and locale %s.", state.ID.String(), cl.Locale.String()),
			)
			return
		}

		state.ConsentID = util.StringValueOrNull(&res.Data.ConsentID)
		if state.ConsentType.ValueString() == SCOPES {
			state.Scopes = util.SetValueOrNull(res.Data.Scopes)
			state.RequiredFields = util.SetValueOrNull(res.Data.RequiredFields)
		}

		objValue := types.ObjectValueMust(
			consentLocalObjectType.AttrTypes,
			map[string]attr.Value{
				"content": util.StringValueOrNull(&res.Data.Content),
				"locale":  util.StringValueOrNull(&res.Data.Locale),
				"url":     util.StringValueOrNull(&res.Data.URL),
			})
		objectValues = append(objectValues, objValue)
		tflog.Debug(ctx, "successfully processed consent locale")
	}

	state.ConsentLocales, diag = types.SetValueFrom(ctx, consentLocalObjectType, objectValues)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to process consent locales data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully processed all consent locales data")

	if !isAvailable {
		tflog.Error(ctx, "consent version not found", util.H{
			"consent_version_id": state.ID.ValueString(),
			"consent_id":         state.ConsentID.ValueString(),
		})
		resp.Diagnostics.AddError(
			"Consent Version not found",
			fmt.Sprintf("Consent Version with ID %s not found for the provided consent_id %s", state.ID.String(), state.ConsentID.String()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "successfully completed consent version read")
}

func (r *ConsentVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConsentVersionConfig

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	for _, pcl := range plan.consentLocale {
		consentLocal := cidaas.ConsentLocalModel{
			ConsentID:        state.ConsentID.ValueString(),
			ConsentVersionID: state.ID.ValueString(),
			Content:          pcl.Content.ValueString(),
			Locale:           pcl.Locale.ValueString(),
		}
		if plan.ConsentType.ValueString() == URL {
			consentLocal.URL = pcl.URL.ValueString()
		}
		_, err := r.cidaasClient.ConsentVersion.UpsertLocal(ctx, consentLocal)
		if err != nil {
			tflog.Error(ctx, "failed to update consent locale via API", util.H{
				"locale":             pcl.Locale.ValueString(),
				"consent_version_id": state.ID.ValueString(),
				"error":              err.Error(),
			})
			resp.Diagnostics.AddError("Failed to update consent locale", fmt.Sprintf("Error: %s", err.Error()))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "successfully completed consent version update")
}

func (r *ConsentVersionResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *ConsentVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	parts := strings.Split(id, ":")
	if len(parts) < 3 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'consent_id:id:locale', got: %s", id),
		)
		return
	}
	locals := parts[2:]
	var objectValues []attr.Value
	consentLocalObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"content": types.StringType,
			"locale":  types.StringType,
			"url":     types.StringType,
		},
	}

	for _, locale := range locals {
		objValue := types.ObjectValueMust(
			consentLocalObjectType.AttrTypes,
			map[string]attr.Value{
				"content": types.StringNull(),
				"locale":  types.StringValue(locale),
				"url":     types.StringNull(),
			})
		objectValues = append(objectValues, objValue)
	}
	resp.State.SetAttribute(ctx, path.Root("consent_id"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("id"), parts[1])
	resp.State.SetAttribute(ctx, path.Root("consent_locales"), types.SetValueMust(consentLocalObjectType, objectValues))
}
