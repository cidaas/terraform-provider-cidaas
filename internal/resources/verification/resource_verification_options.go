// Package verification implements cidaas verification method and option resources.
package verification

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &verificationOptionsResource{}
	_ resource.ResourceWithConfigure      = &verificationOptionsResource{}
	_ resource.ResourceWithImportState    = &verificationOptionsResource{}
	_ resource.ResourceWithValidateConfig = &verificationOptionsResource{}

	settingValues = []string{"OFF", "ALWAYS", "SMART", "TIME_BASED", "SMART_PLUS_TIME_BASED"}
)

// verificationOptionsResource implements cidaas_verification_options.
type verificationOptionsResource struct {
	client *client.Client
}

// verificationOptionsConfig is the Terraform state/plan model.
type verificationOptionsConfig struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Owner               types.String `tfsdk:"owner"`
	VerificationOptions types.Object `tfsdk:"verification_options"`

	options *verificationOptionsDetailConfig
}

// verificationOptionsDetailConfig is the nested verification_options block.
type verificationOptionsDetailConfig struct {
	Setting                     types.String `tfsdk:"setting"`
	TimeIntervalInSeconds       types.Int64  `tfsdk:"time_interval_in_seconds"`
	AllowedMethods              types.List   `tfsdk:"allowed_methods"`
	PasswordPolicyRef           types.String `tfsdk:"password_policy_ref"`
	UseDefaultPasswordPolicy    types.Bool   `tfsdk:"use_default_password_policy"`
	SuggestVerificationMethodID types.String `tfsdk:"suggest_verification_method_id"`
	AppAttest                   types.Object `tfsdk:"app_attest"`

	appAttest *appAttestConfig
}

// appAttestConfig is optional nested app_attest.
type appAttestConfig struct {
	Android types.Object `tfsdk:"android"`
	IOS     types.Object `tfsdk:"ios"`

	android *appAttestAndroidConfig
	ios     *appAttestIOSConfig
}

// appAttestAndroidConfig is app_attest.android.
type appAttestAndroidConfig struct {
	Provider            types.String `tfsdk:"provider"`
	RelaxAppRecognition types.Bool   `tfsdk:"relax_app_recognition"`
	CertDigests         types.List   `tfsdk:"cert_digests"`
	GCPServiceAccount   types.String `tfsdk:"gcp_service_account"`
	ProjectID           types.String `tfsdk:"project_id"`
	IOSAppID            types.String `tfsdk:"ios_app_id"`
	AndroidAppID        types.String `tfsdk:"android_app_id"`
}

// appAttestIOSConfig is app_attest.ios.
type appAttestIOSConfig struct {
	Provider          types.String `tfsdk:"provider"`
	AppleRootCert     types.String `tfsdk:"apple_root_cert"`
	GCPServiceAccount types.String `tfsdk:"gcp_service_account"`
	ProjectID         types.String `tfsdk:"project_id"`
	IOSAppID          types.String `tfsdk:"ios_app_id"`
}

// NewVerificationOptionsResource returns the cidaas_verification_options resource.
func NewVerificationOptionsResource() resource.Resource {
	return &verificationOptionsResource{}
}

// Metadata sets the resource type name to cidaas_verification_options.
func (r *verificationOptionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verification_options"
}

func androidAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider":              types.StringType,
		"relax_app_recognition": types.BoolType,
		"cert_digests":          types.ListType{ElemType: types.StringType},
		"gcp_service_account":   types.StringType,
		"project_id":            types.StringType,
		"ios_app_id":            types.StringType,
		"android_app_id":        types.StringType,
	}
}

func iosAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider":            types.StringType,
		"apple_root_cert":     types.StringType,
		"gcp_service_account": types.StringType,
		"project_id":          types.StringType,
		"ios_app_id":          types.StringType,
	}
}

func appAttestAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"android": types.ObjectType{AttrTypes: androidAttrTypes()},
		"ios":     types.ObjectType{AttrTypes: iosAttrTypes()},
	}
}

func verificationOptionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"setting":                        types.StringType,
		"time_interval_in_seconds":       types.Int64Type,
		"allowed_methods":                types.ListType{ElemType: types.StringType},
		"password_policy_ref":            types.StringType,
		"use_default_password_policy":    types.BoolType,
		"suggest_verification_method_id": types.StringType,
		"app_attest":                     types.ObjectType{AttrTypes: appAttestAttrTypes()},
	}
}

func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}

// Schema defines the Terraform schema for cidaas_verification_options.
func (r *verificationOptionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages verification options via `verification-actions-srv/verification-options`. " +
			"Apps reference this via `authentication_setup.verification_options_id`. " +
			"Requires scopes `cidaas:verification_read`, `cidaas:verification_write`, `cidaas:verification_delete`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 100),
					entityNameValidator{},
				},
			},
			"description": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 600),
				},
			},
			"owner": schema.StringAttribute{
				Computed: true,
			},
			"verification_options": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"setting": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("OFF"),
						Validators: []validator.String{
							stringvalidator.OneOf(settingValues...),
						},
					},
					"time_interval_in_seconds": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"allowed_methods": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     listdefault.StaticValue(emptyStringList()),
					},
					"password_policy_ref": schema.StringAttribute{
						Optional: true,
					},
					"use_default_password_policy": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"suggest_verification_method_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "ID of a `cidaas_suggest_verification_method` resource.",
					},
					"app_attest": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"android": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"provider": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("google", "firebase"),
										},
									},
									"relax_app_recognition": schema.BoolAttribute{Optional: true},
									"cert_digests": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
									"gcp_service_account": schema.StringAttribute{Optional: true, Sensitive: true},
									"project_id":          schema.StringAttribute{Optional: true},
									"ios_app_id":          schema.StringAttribute{Optional: true},
									"android_app_id":      schema.StringAttribute{Optional: true},
								},
							},
							"ios": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"provider": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("apple", "firebase"),
										},
									},
									"apple_root_cert":     schema.StringAttribute{Optional: true, Sensitive: true},
									"gcp_service_account": schema.StringAttribute{Optional: true, Sensitive: true},
									"project_id":          schema.StringAttribute{Optional: true},
									"ios_app_id":          schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *verificationOptionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.ValidateResourceVersion("cidaas_verification_options", &resp.Diagnostics) {
		return
	}
	r.client = c
}

func (c *verificationOptionsConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if c.VerificationOptions.IsNull() || c.VerificationOptions.IsUnknown() {
		return diags
	}
	c.options = &verificationOptionsDetailConfig{}
	diags.Append(c.VerificationOptions.As(ctx, c.options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}
	if !c.options.AppAttest.IsNull() && !c.options.AppAttest.IsUnknown() {
		c.options.appAttest = &appAttestConfig{}
		diags.Append(c.options.AppAttest.As(ctx, c.options.appAttest, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return diags
		}
		if !c.options.appAttest.Android.IsNull() && !c.options.appAttest.Android.IsUnknown() {
			c.options.appAttest.android = &appAttestAndroidConfig{}
			diags.Append(c.options.appAttest.Android.As(ctx, c.options.appAttest.android, basetypes.ObjectAsOptions{})...)
		}
		if !c.options.appAttest.IOS.IsNull() && !c.options.appAttest.IOS.IsUnknown() {
			c.options.appAttest.ios = &appAttestIOSConfig{}
			diags.Append(c.options.appAttest.IOS.As(ctx, c.options.appAttest.ios, basetypes.ObjectAsOptions{})...)
		}
	}
	return diags
}

// ValidateConfig enforces setting-dependent rules (time_interval, allowed_methods).
func (r *verificationOptionsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config verificationOptionsConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.extract(ctx)...)
	if resp.Diagnostics.HasError() || config.options == nil {
		return
	}
	setting := config.options.Setting.ValueString()
	if config.options.Setting.IsUnknown() {
		return
	}
	if setting == "" {
		setting = "OFF"
	}
	if setting == "TIME_BASED" || setting == "SMART_PLUS_TIME_BASED" {
		if config.options.TimeIntervalInSeconds.IsNull() || config.options.TimeIntervalInSeconds.IsUnknown() || config.options.TimeIntervalInSeconds.ValueInt64() <= 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("verification_options").AtName("time_interval_in_seconds"),
				"time_interval_in_seconds required",
				"Required and must be greater than 0 when setting is TIME_BASED or SMART_PLUS_TIME_BASED.",
			)
		}
	}
	if config.options.AllowedMethods.IsUnknown() {
		return
	}
	methods, d := listToStrings(ctx, config.options.AllowedMethods)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if setting != "OFF" && len(methods) == 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("verification_options").AtName("allowed_methods"),
			"Invalid allowed_methods",
			"When setting is not OFF, allowed_methods must be empty or contain at least 2 methods.",
		)
	}
}

func (c *verificationOptionsConfig) toModel(ctx context.Context) (client.VerificationOptionsModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.VerificationOptionsModel{
		ID:          c.ID.ValueString(),
		Name:        c.Name.ValueString(),
		Description: c.Description.ValueString(),
	}
	if c.options == nil {
		return model, diags
	}
	detail := &client.VerificationOptions{
		Setting:                     c.options.Setting.ValueString(),
		PasswordPolicyRef:           c.options.PasswordPolicyRef.ValueString(),
		UseDefaultPasswordPolicy:    !c.options.UseDefaultPasswordPolicy.IsNull() && c.options.UseDefaultPasswordPolicy.ValueBool(),
		SuggestVerificationMethodID: c.options.SuggestVerificationMethodID.ValueString(),
	}
	methods, d := listToStrings(ctx, c.options.AllowedMethods)
	diags.Append(d...)
	if methods == nil {
		methods = []string{}
	}
	detail.AllowedMethods = methods
	if !c.options.TimeIntervalInSeconds.IsNull() && !c.options.TimeIntervalInSeconds.IsUnknown() {
		v := int(c.options.TimeIntervalInSeconds.ValueInt64())
		detail.TimeIntervalInSeconds = &v
	}
	if c.options.appAttest != nil {
		detail.AppAttest = &client.AppAttestEntry{}
		if c.options.appAttest.android != nil {
			a := c.options.appAttest.android
			certs, d := listToStrings(ctx, a.CertDigests)
			diags.Append(d...)
			detail.AppAttest.Android = &client.AppAttestAndroidEntry{
				Provider:            a.Provider.ValueString(),
				RelaxAppRecognition: !a.RelaxAppRecognition.IsNull() && a.RelaxAppRecognition.ValueBool(),
				CertDigests:         certs,
				GCPServiceAccount:   a.GCPServiceAccount.ValueString(),
				ProjectID:           a.ProjectID.ValueString(),
				IOSAppID:            a.IOSAppID.ValueString(),
				AndroidAppID:        a.AndroidAppID.ValueString(),
			}
		}
		if c.options.appAttest.ios != nil {
			i := c.options.appAttest.ios
			detail.AppAttest.IOS = &client.AppAttestIOSEntry{
				Provider:          i.Provider.ValueString(),
				AppleRootCert:     i.AppleRootCert.ValueString(),
				GCPServiceAccount: i.GCPServiceAccount.ValueString(),
				ProjectID:         i.ProjectID.ValueString(),
				IOSAppID:          i.IOSAppID.ValueString(),
			}
		}
	}
	model.VerificationOptions = detail
	return model, diags
}

func flattenVerificationOptions(model client.VerificationOptionsModel) (verificationOptionsConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := verificationOptionsConfig{
		ID:          types.StringValue(model.ID),
		Name:        types.StringValue(model.Name),
		Description: types.StringValue(model.Description),
		Owner:       stringOrNull(model.Owner),
	}
	if model.VerificationOptions == nil {
		cfg.VerificationOptions = types.ObjectNull(verificationOptionsAttrTypes())
		return cfg, diags
	}
	o := model.VerificationOptions
	attrs := map[string]attr.Value{
		"setting":                        types.StringValue(o.Setting),
		"allowed_methods":                stringList(o.AllowedMethods),
		"password_policy_ref":            stringOrNull(o.PasswordPolicyRef),
		"use_default_password_policy":    types.BoolValue(o.UseDefaultPasswordPolicy),
		"suggest_verification_method_id": stringOrNull(o.SuggestVerificationMethodID),
	}
	if o.TimeIntervalInSeconds != nil {
		attrs["time_interval_in_seconds"] = types.Int64Value(int64(*o.TimeIntervalInSeconds))
	} else {
		attrs["time_interval_in_seconds"] = types.Int64Null()
	}
	if o.AppAttest != nil {
		appAttrs := map[string]attr.Value{}
		if o.AppAttest.Android != nil {
			a := o.AppAttest.Android
			obj, d := types.ObjectValue(androidAttrTypes(), map[string]attr.Value{
				"provider":              types.StringValue(a.Provider),
				"relax_app_recognition": types.BoolValue(a.RelaxAppRecognition),
				"cert_digests":          stringList(a.CertDigests),
				"gcp_service_account":   stringOrNull(a.GCPServiceAccount),
				"project_id":            stringOrNull(a.ProjectID),
				"ios_app_id":            stringOrNull(a.IOSAppID),
				"android_app_id":        stringOrNull(a.AndroidAppID),
			})
			diags.Append(d...)
			appAttrs["android"] = obj
		} else {
			appAttrs["android"] = types.ObjectNull(androidAttrTypes())
		}
		if o.AppAttest.IOS != nil {
			i := o.AppAttest.IOS
			obj, d := types.ObjectValue(iosAttrTypes(), map[string]attr.Value{
				"provider":            types.StringValue(i.Provider),
				"apple_root_cert":     stringOrNull(i.AppleRootCert),
				"gcp_service_account": stringOrNull(i.GCPServiceAccount),
				"project_id":          stringOrNull(i.ProjectID),
				"ios_app_id":          stringOrNull(i.IOSAppID),
			})
			diags.Append(d...)
			appAttrs["ios"] = obj
		} else {
			appAttrs["ios"] = types.ObjectNull(iosAttrTypes())
		}
		obj, d := types.ObjectValue(appAttestAttrTypes(), appAttrs)
		diags.Append(d...)
		attrs["app_attest"] = obj
	} else {
		attrs["app_attest"] = types.ObjectNull(appAttestAttrTypes())
	}
	obj, d := types.ObjectValue(verificationOptionsAttrTypes(), attrs)
	diags.Append(d...)
	cfg.VerificationOptions = obj
	return cfg, diags
}

// Create creates via POST /verification-actions-srv/verification-options/.
func (r *verificationOptionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { //nolint:dupl // mirrors suggest_verification_method CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan verificationOptionsConfig
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
	model.ID = ""
	res, err := r.client.VerificationOptions.Create(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Create verification options failed", err.Error())
		return
	}
	state, diags := flattenVerificationOptions(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state via GET .../verification-options/{id}. Removes from state on 404.
func (r *verificationOptionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) { //nolint:dupl // mirrors suggest_verification_method CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state verificationOptionsConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.VerificationOptions.Get(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read verification options failed", err.Error())
		return
	}
	next, diags := flattenVerificationOptions(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update applies changes via PUT .../verification-options/{id} (srv verb).
func (r *verificationOptionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl // mirrors suggest_verification_method CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan verificationOptionsConfig
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
	res, err := r.client.VerificationOptions.Update(ctx, plan.ID.ValueString(), model)
	if err != nil {
		resp.Diagnostics.AddError("Update verification options failed", err.Error())
		return
	}
	state, diags := flattenVerificationOptions(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes via DELETE .../verification-options/{id}.
func (r *verificationOptionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl // mirrors suggest_verification_method CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state verificationOptionsConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.VerificationOptions.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete verification options failed", err.Error())
	}
}

// ImportState imports by server-assigned UUID (id).
func (r *verificationOptionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
