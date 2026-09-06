// Package verification implements Terraform resources for verification-actions-srv:
// cidaas_suggest_verification_method and cidaas_verification_options.
package verification

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &suggestVerificationMethodResource{}
	_ resource.ResourceWithConfigure      = &suggestVerificationMethodResource{}
	_ resource.ResourceWithImportState    = &suggestVerificationMethodResource{}
	_ resource.ResourceWithValidateConfig = &suggestVerificationMethodResource{}

	entityNameRE = regexp.MustCompile(`^[a-zA-Z0-9:_.-]+$`)
	rangeValues  = []string{"ALLOF", "ONEOF"}
)

// suggestVerificationMethodResource implements cidaas_suggest_verification_method.
type suggestVerificationMethodResource struct {
	client *client.Client
}

// suggestVerificationMethodConfig is the Terraform state/plan model.
type suggestVerificationMethodConfig struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Owner                     types.String `tfsdk:"owner"`
	SuggestVerificationMethod types.Object `tfsdk:"suggest_verification_method"`

	setup *suggestVerificationMethodSetupConfig
}

// suggestVerificationMethodSetupConfig is the nested suggest_verification_method block.
type suggestVerificationMethodSetupConfig struct {
	MandatoryConfig    types.Object `tfsdk:"mandatory_config"`
	OptionalConfig     types.Object `tfsdk:"optional_config"`
	SkipDurationInDays types.Int64  `tfsdk:"skip_duration_in_days"`

	mandatory *mandatoryConfigModel
	optional  *optionalConfigModel
}

// mandatoryConfigModel is mandatory_config (methods + range).
type mandatoryConfigModel struct {
	Methods   types.List   `tfsdk:"methods"`
	Range     types.String `tfsdk:"range"`
	SkipUntil types.String `tfsdk:"skip_until"`
}

// optionalConfigModel is optional_config (methods only).
type optionalConfigModel struct {
	Methods types.List `tfsdk:"methods"`
}

// NewSuggestVerificationMethodResource returns the cidaas_suggest_verification_method resource.
func NewSuggestVerificationMethodResource() resource.Resource {
	return &suggestVerificationMethodResource{}
}

// Metadata sets the resource type name to cidaas_suggest_verification_method.
func (r *suggestVerificationMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_suggest_verification_method"
}

func mandatoryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"methods":    types.ListType{ElemType: types.StringType},
		"range":      types.StringType,
		"skip_until": types.StringType,
	}
}

func optionalAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"methods": types.ListType{ElemType: types.StringType},
	}
}

func setupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mandatory_config":      types.ObjectType{AttrTypes: mandatoryAttrTypes()},
		"optional_config":       types.ObjectType{AttrTypes: optionalAttrTypes()},
		"skip_duration_in_days": types.Int64Type,
	}
}

// Schema defines the Terraform schema for cidaas_suggest_verification_method.
func (r *suggestVerificationMethodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Suggest Verification method via `verification-actions-srv/suggest-verification-configs`. " +
			"Referenced by `cidaas_verification_options.verification_options.suggest_verification_method_id`. " +
			"Requires scopes `cidaas:verification_read`, `cidaas:verification_write`, `cidaas:verification_delete`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique name (3–100 chars; letters, numbers, `:`, `_`, `.`, `-`).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 100),
					entityNameValidator{},
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Description (max 600 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 600),
				},
			},
			"owner": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server-assigned owner (`client` on create).",
			},
			"suggest_verification_method": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Mandatory and/or optional method configuration.",
				Attributes: map[string]schema.Attribute{
					"mandatory_config": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"methods": schema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"range": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf(rangeValues...),
								},
							},
							"skip_until": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Optional RFC3339 timestamp.",
							},
						},
					},
					"optional_config": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"methods": schema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
					"skip_duration_in_days": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(7),
						Validators: []validator.Int64{
							int64validator.Between(1, 365),
						},
					},
				},
			},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *suggestVerificationMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.ValidateResourceVersion("cidaas_suggest_verification_method", &resp.Diagnostics) {
		return
	}
	r.client = c
}

func (c *suggestVerificationMethodConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if c.SuggestVerificationMethod.IsNull() || c.SuggestVerificationMethod.IsUnknown() {
		return diags
	}
	c.setup = &suggestVerificationMethodSetupConfig{}
	diags.Append(c.SuggestVerificationMethod.As(ctx, c.setup, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}
	if !c.setup.MandatoryConfig.IsNull() && !c.setup.MandatoryConfig.IsUnknown() {
		c.setup.mandatory = &mandatoryConfigModel{}
		diags.Append(c.setup.MandatoryConfig.As(ctx, c.setup.mandatory, basetypes.ObjectAsOptions{})...)
	}
	if !c.setup.OptionalConfig.IsNull() && !c.setup.OptionalConfig.IsUnknown() {
		c.setup.optional = &optionalConfigModel{}
		diags.Append(c.setup.OptionalConfig.As(ctx, c.setup.optional, basetypes.ObjectAsOptions{})...)
	}
	return diags
}

// ValidateConfig enforces mandatory/optional presence and no overlapping methods.
func (r *suggestVerificationMethodResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config suggestVerificationMethodConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.extract(ctx)...)
	if resp.Diagnostics.HasError() || config.setup == nil {
		return
	}
	if config.setup.mandatory == nil && config.setup.optional == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("suggest_verification_method"),
			"mandatory_config or optional_config required",
			"At least one of mandatory_config or optional_config must be set.",
		)
		return
	}
	if config.setup.mandatory != nil && config.setup.optional != nil {
		if config.setup.mandatory.Methods.IsUnknown() || config.setup.optional.Methods.IsUnknown() {
			return
		}
		mand, d := listToStrings(ctx, config.setup.mandatory.Methods)
		resp.Diagnostics.Append(d...)
		opt, d := listToStrings(ctx, config.setup.optional.Methods)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if overlap := methodOverlap(mand, opt); len(overlap) > 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("suggest_verification_method").AtName("optional_config").AtName("methods"),
				"Methods cannot be both mandatory and optional",
				fmt.Sprintf("Overlapping methods: %s", strings.Join(overlap, ", ")),
			)
		}
	}
}

func methodOverlap(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	var out []string
	for _, v := range b {
		if _, ok := set[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

func (c *suggestVerificationMethodConfig) toModel(ctx context.Context) (client.SuggestVerificationMethodModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.SuggestVerificationMethodModel{
		ID:          c.ID.ValueString(),
		Name:        c.Name.ValueString(),
		Description: c.Description.ValueString(),
	}
	if c.setup == nil {
		return model, diags
	}
	setup := &client.SuggestVerificationMethodSetup{
		SkipDurationInDays: int(c.setup.SkipDurationInDays.ValueInt64()),
	}
	if c.setup.mandatory != nil {
		methods, d := listToStrings(ctx, c.setup.mandatory.Methods)
		diags.Append(d...)
		setup.MandatoryConfig = &client.SuggestVerificationMandatoryConfig{
			Methods:   methods,
			Range:     c.setup.mandatory.Range.ValueString(),
			SkipUntil: c.setup.mandatory.SkipUntil.ValueString(),
		}
	}
	if c.setup.optional != nil {
		methods, d := listToStrings(ctx, c.setup.optional.Methods)
		diags.Append(d...)
		setup.OptionalConfig = &client.SuggestVerificationOptionalConfig{Methods: methods}
	}
	model.SuggestVerificationMethod = setup
	return model, diags
}

func flattenSuggestVerificationMethod(model client.SuggestVerificationMethodModel) (suggestVerificationMethodConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := suggestVerificationMethodConfig{
		ID:          types.StringValue(model.ID),
		Name:        types.StringValue(model.Name),
		Description: types.StringValue(model.Description),
		Owner:       stringOrNull(model.Owner),
	}
	if model.SuggestVerificationMethod == nil {
		cfg.SuggestVerificationMethod = types.ObjectNull(setupAttrTypes())
		return cfg, diags
	}
	setup := model.SuggestVerificationMethod
	attrs := map[string]attr.Value{
		"skip_duration_in_days": types.Int64Value(int64(setup.SkipDurationInDays)),
	}
	if setup.MandatoryConfig != nil {
		obj, d := types.ObjectValue(mandatoryAttrTypes(), map[string]attr.Value{
			"methods":    stringList(setup.MandatoryConfig.Methods),
			"range":      types.StringValue(setup.MandatoryConfig.Range),
			"skip_until": stringOrNull(setup.MandatoryConfig.SkipUntil),
		})
		diags.Append(d...)
		attrs["mandatory_config"] = obj
	} else {
		attrs["mandatory_config"] = types.ObjectNull(mandatoryAttrTypes())
	}
	if setup.OptionalConfig != nil {
		obj, d := types.ObjectValue(optionalAttrTypes(), map[string]attr.Value{
			"methods": stringList(setup.OptionalConfig.Methods),
		})
		diags.Append(d...)
		attrs["optional_config"] = obj
	} else {
		attrs["optional_config"] = types.ObjectNull(optionalAttrTypes())
	}
	obj, d := types.ObjectValue(setupAttrTypes(), attrs)
	diags.Append(d...)
	cfg.SuggestVerificationMethod = obj
	return cfg, diags
}

// Create creates via POST /verification-actions-srv/suggest-verification-configs/.
func (r *suggestVerificationMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { //nolint:dupl // mirrors verification_options CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan suggestVerificationMethodConfig
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
	res, err := r.client.SuggestVerificationMethod.Create(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Create suggest verification method failed", err.Error())
		return
	}
	state, diags := flattenSuggestVerificationMethod(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state via GET .../suggest-verification-configs/{id}. Removes from state on 404.
func (r *suggestVerificationMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) { //nolint:dupl // mirrors verification_options CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state suggestVerificationMethodConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.SuggestVerificationMethod.Get(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read suggest verification method failed", err.Error())
		return
	}
	next, diags := flattenSuggestVerificationMethod(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update applies changes via PUT .../suggest-verification-configs/{id} (srv verb).
func (r *suggestVerificationMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl // mirrors verification_options CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan suggestVerificationMethodConfig
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
	res, err := r.client.SuggestVerificationMethod.Update(ctx, plan.ID.ValueString(), model)
	if err != nil {
		resp.Diagnostics.AddError("Update suggest verification method failed", err.Error())
		return
	}
	state, diags := flattenSuggestVerificationMethod(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes via DELETE .../suggest-verification-configs/{id}.
func (r *suggestVerificationMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl // mirrors verification_options CRUD
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state suggestVerificationMethodConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.SuggestVerificationMethod.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete suggest verification method failed", err.Error())
	}
}

// ImportState imports by server-assigned UUID (id).
func (r *suggestVerificationMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func listToStrings(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	var out []string
	diags := l.ElementsAs(ctx, &out, false)
	return out, diags
}

func stringList(values []string) types.List {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

type entityNameValidator struct{}

func (v entityNameValidator) Description(_ context.Context) string {
	return "name must match [a-zA-Z0-9:_.-]+ and be 3–100 characters"
}

func (v entityNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v entityNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	name := req.ConfigValue.ValueString()
	if !entityNameRE.MatchString(name) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid name", "name may include letters, numbers, colon (:), underscore (_), dot (.), and hyphen (-) only")
	}
}
