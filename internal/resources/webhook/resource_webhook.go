// Package webhook manages cIDAAS webhook resources.
package webhook

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

//nolint:revive
type WebhookResource struct {
	base.BaseResource
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_WEBHOOK,
				Schema: &webhookSchema,
			},
		),
	}
}

//nolint:revive
type WebhookConfig struct {
	ID               types.String `tfsdk:"id"`
	AuthType         types.String `tfsdk:"auth_type"`
	URL              types.String `tfsdk:"url"`
	Events           types.Set    `tfsdk:"events"`
	APIKeyConfig     types.Object `tfsdk:"apikey_config"`
	TOTPConfig       types.Object `tfsdk:"totp_config"`
	CidaasAuthConfig types.Object `tfsdk:"cidaas_auth_config"`
	Disable          types.Bool   `tfsdk:"disable"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	apiKeyConfig     *AuthConfig
	totpConfig       *AuthConfig
	cidaasAuthConfig *CidaasAuthConfig
}

type AuthConfig struct {
	Placeholder types.String `tfsdk:"placeholder"`
	Placement   types.String `tfsdk:"placement"`
	Key         types.String `tfsdk:"key"`
}

type CidaasAuthConfig struct {
	ClientID types.String `tfsdk:"client_id"`
}

func (w *WebhookConfig) extractAuthConfigs(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !w.APIKeyConfig.IsNull() {
		w.apiKeyConfig = &AuthConfig{}
		diags = w.APIKeyConfig.As(ctx, w.apiKeyConfig, basetypes.ObjectAsOptions{})
	}
	if !w.TOTPConfig.IsNull() {
		w.totpConfig = &AuthConfig{}
		diags = w.TOTPConfig.As(ctx, w.totpConfig, basetypes.ObjectAsOptions{})
	}
	if !w.CidaasAuthConfig.IsNull() {
		w.cidaasAuthConfig = &CidaasAuthConfig{}
		diags = w.CidaasAuthConfig.As(ctx, w.cidaasAuthConfig, basetypes.ObjectAsOptions{})
	}
	return diags
}

func authConfigObject(placeholder, placement, key *string) (types.Object, diag.Diagnostics) {
	authConfig := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"placeholder": types.StringType,
			"placement":   types.StringType,
			"key":         types.StringType,
		},
	}
	return types.ObjectValue(authConfig.AttrTypes, map[string]attr.Value{
		"placeholder": util.StringValueOrNull(placeholder),
		"placement":   util.StringValueOrNull(placement),
		"key":         util.StringValueOrNull(key),
	})
}

var webhookSchema = schema.Schema{
	MarkdownDescription: "The `cidaas_webhook` resource manages Webhook integration settings in Cidaas.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the webhook.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"auth_type": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Determines the authentication method. Allowed values are `APIKEY`, `TOTP`, `CIDAAS_OAUTH2`, and `NO_AUTH`.",
			Validators: []validator.String{
				stringvalidator.OneOf(cidaas.AllowedAuthType...),
			},
			PlanModifiers: []planmodifier.String{
				&configVerifier{},
			},
		},
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Target URL to receive webhook event payloads.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^https://`),
					"must be a valid https URL",
				),
			},
		},
		"events": schema.SetAttribute{
			ElementType:         types.StringType,
			Required:            true,
			MarkdownDescription: "Set of event IDs subscribed by this webhook. Discover values via `GET /webhook-srv/eventdescriptions?category=webhook`.",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"apikey_config": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "API Key authentication details.",
			Attributes: map[string]schema.Attribute{
				"placeholder": schema.StringAttribute{
					Required: true,
				},
				"placement": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf(cidaas.AllowedPlacement...),
					},
				},
				"key": schema.StringAttribute{
					Required:  true,
					Sensitive: true,
				},
			},
		},
		"totp_config": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "TOTP authentication details.",
			Attributes: map[string]schema.Attribute{
				"placeholder": schema.StringAttribute{
					Required: true,
				},
				"placement": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf(cidaas.AllowedPlacement...),
					},
				},
				"key": schema.StringAttribute{
					Required:  true,
					Sensitive: true,
				},
			},
		},
		"cidaas_auth_config": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "OAuth2 client details.",
			Attributes: map[string]schema.Attribute{
				"client_id": schema.StringAttribute{
					Required: true,
				},
			},
		},
		"disable": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Flag indicating if the webhook is disabled.",
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: "The timestamp when the resource was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: "The timestamp when the resource was last updated.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Starting webhook creation")

	var plan WebhookConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extractAuthConfigs(ctx)...)
	wbModel, diags := prepareWebhookModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to prepare webhook model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	resp.Diagnostics.Append(validateWebhookEvents(ctx, r.CidaasClient, wbModel.Events)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Calling Cidaas API to create webhook")
	res, err := r.CidaasClient.Webhook.Upsert(ctx, *wbModel)
	if err != nil {
		tflog.Error(ctx, "Failed to create webhook via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create webhook", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "Successfully created webhook via API", util.H{
		"webhook_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	plan.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	plan.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state after creation", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.CidaasClient.Webhook.Get(ctx, state.ID.ValueString())
	if err != nil {
		if base.ReadHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read webhook via API", util.H{
			"webhook_id": state.ID.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to read webhook", util.FormatErrorMessage(err))
		return
	}

	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.AuthType = util.StringValueOrNull(&res.Data.AuthType)
	state.URL = util.StringValueOrNull(&res.Data.URL)
	state.Disable = util.BoolValueOrNull(&res.Data.Disable)
	state.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	state.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)
	state.Events = util.SetValueOrNull(res.Data.Events)

	if res.Data.APIKeyDetails.Apikey != "" {
		apiKeyConfig, diags := authConfigObject(
			&res.Data.APIKeyDetails.ApikeyPlaceholder,
			&res.Data.APIKeyDetails.ApikeyPlacement,
			&res.Data.APIKeyDetails.Apikey,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.APIKeyConfig = apiKeyConfig
	}

	if res.Data.TotpDetails.TotpKey != "" {
		totpConfig, diags := authConfigObject(
			&res.Data.TotpDetails.TotpPlaceholder,
			&res.Data.TotpDetails.TotpPlacement,
			&res.Data.TotpDetails.TotpKey,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.TOTPConfig = totpConfig
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
		state.CidaasAuthConfig = oauthConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Starting webhook update")

	var plan, state WebhookConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(plan.extractAuthConfigs(ctx)...)
	wbModel, diags := prepareWebhookModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateWebhookEvents(ctx, r.CidaasClient, wbModel.Events)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.CidaasClient.Webhook.Upsert(ctx, *wbModel)
	if err != nil {
		resp.Diagnostics.AddError("failed to update webhook", util.FormatErrorMessage(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.Webhook.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete webhook", util.FormatErrorMessage(err))
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ planmodifier.String = configVerifier{}

type configVerifier struct{}

func (v configVerifier) Description(_ context.Context) string {
	return "Verifies the availability of config details for the provided auth_type."
}

func (v configVerifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v configVerifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsUnknown() ||
		req.ConfigValue.IsNull() ||
		!util.Contains(cidaas.AllowedAuthType, req.ConfigValue.ValueString()) {
		return
	}

	var tempConfig types.Object
	configAttr := "apikey_config"
	authType := "APIKEY"

	if req.ConfigValue.ValueString() == "TOTP" {
		configAttr = "totp_config"
		authType = "TOTP"
	}
	if req.ConfigValue.ValueString() == "CIDAAS_OAUTH2" {
		configAttr = "cidaas_auth_config"
		authType = "CIDAAS_OAUTH2"
	}
	diags := req.Config.GetAttribute(ctx, path.Root(configAttr), &tempConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if tempConfig.IsNull() {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("The attribute %s cannot be empty when the auth_type is %s", configAttr, authType))
	}
}

func prepareWebhookModel(ctx context.Context, plan WebhookConfig) (*cidaas.WebhookModel, diag.Diagnostics) {
	wb := cidaas.WebhookModel{
		ID:       plan.ID.ValueString(),
		AuthType: plan.AuthType.ValueString(),
		URL:      plan.URL.ValueString(),
		Disable:  plan.Disable.ValueBool(),
	}
	if !plan.APIKeyConfig.IsNull() {
		wb.APIKeyDetails = cidaas.APIKeyDetails{
			ApikeyPlaceholder: plan.apiKeyConfig.Placeholder.ValueString(),
			ApikeyPlacement:   plan.apiKeyConfig.Placement.ValueString(),
			Apikey:            plan.apiKeyConfig.Key.ValueString(),
		}
	}
	if !plan.TOTPConfig.IsNull() {
		wb.TotpDetails = cidaas.TotpDetails{
			TotpPlaceholder: plan.totpConfig.Placeholder.ValueString(),
			TotpPlacement:   plan.totpConfig.Placement.ValueString(),
			TotpKey:         plan.totpConfig.Key.ValueString(),
		}
	}
	if !plan.CidaasAuthConfig.IsNull() {
		wb.CidaasAuthDetails = cidaas.AuthDetails{
			ClientID: plan.cidaasAuthConfig.ClientID.ValueString(),
		}
	}
	diags := plan.Events.ElementsAs(ctx, &wb.Events, false)
	if diags.HasError() {
		return nil, diags
	}
	return &wb, nil
}

func validateWebhookEvents(ctx context.Context, client *cidaas.Client, events []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if client == nil || client.Webhook == nil {
		diags.AddError("Webhook client not configured", "Cannot validate webhook events without a configured Cidaas client.")
		return diags
	}
	allowed, err := client.Webhook.ListWebhookEventIDs(ctx)
	if err != nil {
		diags.AddError("failed to list webhook-capable events", util.FormatErrorMessage(err))
		return diags
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	for _, event := range events {
		if _, ok := allowedSet[event]; !ok {
			diags.AddError(
				"Invalid webhook event",
				fmt.Sprintf("event %q is not a webhook-capable event in this tenant; use GET /webhook-srv/eventdescriptions?category=webhook to list valid values", event),
			)
		}
	}
	return diags
}
