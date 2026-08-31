// Package app implements cidaas_app_configuration and deprecated cidaas_app resources.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &appConfigurationResource{}
	_ resource.ResourceWithConfigure   = &appConfigurationResource{}
	_ resource.ResourceWithImportState = &appConfigurationResource{}

	clientTypeValues = []string{
		"NON_INTERACTIVE",
		"SINGLE_PAGE",
		"REGULAR_WEB",
		"NATIVE",
		"DEVICE",
		"THIRD_PARTY",
	}
)

type appConfigurationResource struct {
	client *client.Client
}

// NewAppConfigurationResource returns the cidaas_app_configuration resource implementation.
func NewAppConfigurationResource() resource.Resource {
	return &appConfigurationResource{}
}

// Metadata sets the resource type name to cidaas_app_configuration.
func (r *appConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_configuration"
}

// Schema defines the Terraform schema for cidaas_app_configuration.
func (r *appConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages appv3 application configuration on cidaas v4 (Trustdesk) via `app-srv/apps`. " +
			"Extdep content (user setup, hosted pages, verification) is referenced by ID only.\n\n" +
			"Requires OAuth scopes `cidaas:apps_read`, `cidaas:apps_write`, `cidaas:apps_delete`.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OAuth client ID. Auto-generated when omitted on create. Import key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique application name.",
			},
			"client_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(clientTypeValues...),
				},
			},
			"owner": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application owner. Always `client` for Trustdesk customer apps (set automatically on create/update).",
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"grant_types": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(emptyStringList()),
			},
			"response_types": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(emptyStringList()),
			},
			"redirect_uris": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"redirect_uris": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
					"allowed_logout_urls": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
					"post_logout_redirect_uris": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
					"allowed_web_origins": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
				},
			},
			"scopes": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"allowed_scopes": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"default_scopes": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
				},
			},
			"token_lifetimes": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"token_lifetime_in_seconds":         schema.Int64Attribute{Optional: true},
					"refresh_token_lifetime_in_seconds": schema.Int64Attribute{Optional: true},
					"id_token_lifetime_in_seconds":      schema.Int64Attribute{Optional: true},
					"code_lifetime_in_seconds":          schema.Int64Attribute{Optional: true},
					"default_max_age":                   schema.Int64Attribute{Optional: true},
				},
			},
			"authentication_setup": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"verification_options_id":          schema.StringAttribute{Optional: true},
					"group_selection_id":               schema.StringAttribute{Optional: true},
					"group_verification_request_id":    schema.StringAttribute{Optional: true},
					"template_group_id":                schema.StringAttribute{Optional: true},
					"allow_guest_login":                schema.BoolAttribute{Optional: true},
					"is_remember_me_selected":          schema.BoolAttribute{Optional: true},
					"admin_client":                     schema.BoolAttribute{Optional: true},
					"is_login_success_page_enabled":    schema.BoolAttribute{Optional: true},
					"is_register_success_page_enabled": schema.BoolAttribute{Optional: true},
				},
			},
			"hosted_pages_layout_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Reference to `cidaas_hosted_page_layout`.",
			},
			"user_setup_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Reference to `cidaas_user_setup`.",
			},
			"ownership_details": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"company_name":    schema.StringAttribute{Required: true},
					"company_address": schema.StringAttribute{Required: true},
					"company_website": schema.StringAttribute{Required: true},
				},
			},
			"client_auth_config": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"token_endpoint_auth_method": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "OAuth token endpoint auth method. NON_INTERACTIVE defaults to `client_secret_basic` (requires a client secret). Use `none` when secrets are managed outside this resource.",
						Validators: []validator.String{
							stringvalidator.OneOf("none", "client_secret_basic", "client_secret_post", "private_key_jwt", "tls_client_auth"),
						},
					},
				},
			},
			"signing_key_config": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"active_kid": schema.StringAttribute{Computed: true},
					"next_kid":   schema.StringAttribute{Computed: true},
				},
			},
			"created_time": schema.StringAttribute{Computed: true},
			"updated_time": schema.StringAttribute{Computed: true},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *appConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.ValidateResourceVersion("cidaas_app_configuration", &resp.Diagnostics) {
		return
	}
	r.client = c
}

// Create creates an app via POST /app-srv/apps/.
func (r *appConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan appConfigurationConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, diags := plan.toModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.ClientID = ""

	res, err := r.client.AppConfiguration.Create(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Create app configuration failed", err.Error())
		return
	}
	state, diags := flattenAppConfiguration(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state via GET /app-srv/apps/{client_id}. Removes from state on 404.
func (r *appConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state appConfigurationConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.AppConfiguration.Get(ctx, state.ClientID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read app configuration failed", err.Error())
		return
	}
	next, diags := flattenAppConfiguration(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update applies changes via PUT /app-srv/apps/{client_id} (full replace).
func (r *appConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan appConfigurationConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, diags := plan.toModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.ClientID = plan.ClientID.ValueString()

	res, err := r.client.AppConfiguration.Update(ctx, plan.ClientID.ValueString(), model)
	if err != nil {
		resp.Diagnostics.AddError("Update app configuration failed", err.Error())
		return
	}
	state, diags := flattenAppConfiguration(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes the app via DELETE /app-srv/apps/{client_id}.
func (r *appConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state appConfigurationConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AppConfiguration.Delete(ctx, state.ClientID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete app configuration failed", err.Error())
	}
}

// ImportState imports by OAuth client_id.
func (r *appConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("client_id"), req, resp)
}
