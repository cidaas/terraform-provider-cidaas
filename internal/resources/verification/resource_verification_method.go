package verification

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &verificationMethodResource{}
	_ resource.ResourceWithConfigure      = &verificationMethodResource{}
	_ resource.ResourceWithImportState    = &verificationMethodResource{}
	_ resource.ResourceWithValidateConfig = &verificationMethodResource{}

	verificationMethodAuthTypes   = []string{"Bearer", "oauth", "oauth2"}
	verificationMethodFactors     = []string{"KNOWLEDGE", "POSSESSION", "INHERENCE", "DELEGATION"}
	verificationMethodEncModes    = []string{"OPTIONAL", "REQUIRED"}
	verificationMethodAbsoluteURL = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^/?#\s]+`)

	// verificationMethodPluggableAttrs can only be written on pluggable (owner CLIENT) methods.
	verificationMethodPluggableAttrs = []string{"url", "auth_type", "cidaas_auth_details", "prerequisite", "factor", "category", "icon", "amr_values", "acr_pin"}

	verificationMethodLocaleType = types.ObjectType{AttrTypes: map[string]attr.Type{
		"locale":      types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	}}
	verificationMethodEncryptionAttrTypes = map[string]attr.Type{
		"enabled":          types.BoolType,
		"mode":             types.StringType,
		"encrypted_fields": types.ListType{ElemType: types.StringType},
	}
	verificationMethodAuthAttrTypes = map[string]attr.Type{"client_id": types.StringType}
	verificationMethodCodeAttrTypes = map[string]attr.Type{"length": types.Int64Type, "validity": types.Int64Type}
)

// verificationMethodResource implements cidaas_verification_method (verification-actions-srv /config).
type verificationMethodResource struct {
	client *client.Client
}

// MethodState is the Terraform model of cidaas_verification_method, shared with the data source.
type MethodState struct {
	ID                       types.String `tfsdk:"id"`
	VerificationType         types.String `tfsdk:"verification_type"`
	Active                   types.Bool   `tfsdk:"active"`
	LocaleText               types.List   `tfsdk:"locale_text"`
	EncryptionConfig         types.Object `tfsdk:"encryption_config"`
	URL                      types.String `tfsdk:"url"`
	AuthType                 types.String `tfsdk:"auth_type"`
	CidaasAuthDetails        types.Object `tfsdk:"cidaas_auth_details"`
	Prerequisite             types.String `tfsdk:"prerequisite"`
	Factor                   types.String `tfsdk:"factor"`
	Category                 types.String `tfsdk:"category"`
	Icon                     types.String `tfsdk:"icon"`
	AMRValues                types.List   `tfsdk:"amr_values"`
	ACRPin                   types.String `tfsdk:"acr_pin"`
	Owner                    types.String `tfsdk:"owner"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	RequiresAuthenticatorApp types.Bool   `tfsdk:"requires_authenticator_app"`
	AlgorithmTypeID          types.String `tfsdk:"algorithm_type_id"`
	Code                     types.Object `tfsdk:"code"`
	CreatedTime              types.String `tfsdk:"created_time"`
	UpdatedTime              types.String `tfsdk:"updated_time"`
}

// NewVerificationMethodResource returns the cidaas_verification_method resource.
func NewVerificationMethodResource() resource.Resource {
	return &verificationMethodResource{}
}

// Metadata sets the resource type name to cidaas_verification_method.
func (r *verificationMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verification_method"
}

// Configure blocks the resource unless the provider targets cidaas 4.x.
func (r *verificationMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.RequireV4("cidaas_verification_method", &resp.Diagnostics) {
		return
	}
	r.client = c
}

// Schema returns the resource schema.
func (r *verificationMethodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MethodResourceSchema()
}

// MethodResourceSchema builds the schema; the data source derives its schema from it.
func MethodResourceSchema() schema.Schema {
	keepString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	optionalComputedString := func(desc string, validators ...validator.String) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: desc, PlanModifiers: keepString, Validators: validators}
	}
	return schema.Schema{
		MarkdownDescription: "Manages an MFA verification method on cidaas v4 via `verification-actions-srv/config` " +
			"(Admin → Multifactor Settings → MFA Settings). Separate from `cidaas_verification_options` and `cidaas_suggest_verification_method`." +
			"\n\n**System methods** (`EMAIL`, `SMS`, `TOTP`, …, owner other than `CLIENT`) already exist: Terraform adopts them and manages " +
			"`active`, `locale_text` and `encryption_config`. Destroying the resource only removes it from state; the method is left unchanged." +
			"\n\n**Pluggable methods** (owner `CLIENT`) are created and deleted by Terraform and need `url`, `auth_type` and `cidaas_auth_details`." +
			"\n\nWhether a method is system or pluggable is decided by its `owner` on the server, like verification-actions-srv does.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Same as `verification_type`.",
				PlanModifiers:       keepString,
			},
			"verification_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Verification method type: a system type (`PASSWORD`, `EMAIL`, `SMS`, `IVR`, `TOTP`, `PUSH`, `TOUCHID`, " +
					"`BACKUPCODE`, `FACE`, `FIDO2`, `PATTERN`, `CHAT`) or the name of a pluggable method. Changing it forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the method is enabled. Kept as is when omitted.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			// A list of objects (not a nested attribute): Terraform proposes a change for omitted optional+computed
			// nested lists on every plan, which would show name/description/updated_time as changing forever.
			"locale_text": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: verificationMethodLocaleType,
				MarkdownDescription: "Display name and description per locale, as a list of `{ locale, name, description }` objects " +
					"(all three keys are required; use `\"\"` for an unused name or description). The API replaces the whole list " +
					"on every update, so when omitted the current texts are sent back unchanged.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"encryption_config": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Encryption settings. System methods only, and only when the method already has an encryption " +
					"configuration on the tenant. Only `enabled` and `mode` can be changed.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{Required: true, MarkdownDescription: "Enable encryption."},
					"mode": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "`OPTIONAL` or `REQUIRED`.",
						Validators:          []validator.String{stringvalidator.OneOf(verificationMethodEncModes...)},
					},
					"encrypted_fields": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Encrypted fields (read-only).",
					},
				},
			},
			"url": optionalComputedString("Pluggable only. Absolute URL of the pluggable method's service (scheme and host).",
				stringvalidator.RegexMatches(verificationMethodAbsoluteURL, "must be an absolute URL (scheme and host)")),
			"auth_type": optionalComputedString("Pluggable only. How cidaas authenticates against `url`: `Bearer`, `oauth` or `oauth2`. Required with `url`.",
				stringvalidator.OneOf(verificationMethodAuthTypes...)),
			"cidaas_auth_details": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Pluggable only. cidaas client used to obtain a token for calling `url`. Required with `url`.",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{Required: true, MarkdownDescription: "Client ID."},
				},
			},
			"prerequisite": optionalComputedString("Pluggable only. Prerequisite of the method."),
			"factor": optionalComputedString("Authentication factor: `KNOWLEDGE`, `POSSESSION`, `INHERENCE` or `DELEGATION`. Writable for pluggable methods only.",
				stringvalidator.OneOf(verificationMethodFactors...)),
			"category": optionalComputedString("Method category, e.g. `code`, `biometric`, `hardware`. Writable for pluggable methods only."),
			"icon":     optionalComputedString("Icon. Writable for pluggable methods only."),
			"amr_values": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				MarkdownDescription: "AMR values. Writable for pluggable methods only; values already used by system methods are " +
					"rejected by the API.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"acr_pin": optionalComputedString("ACR value. Writable for pluggable methods only."),
			"owner": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`CLIENT` for pluggable methods; any other value (e.g. `system`) for system methods.",
				PlanModifiers:       keepString,
			},
			"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Display name resolved by the API from `locale_text`."},
			"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description resolved by the API from `locale_text`."},
			"requires_authenticator_app": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the method needs the authenticator app.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"algorithm_type_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Password algorithm (PASSWORD method).",
				PlanModifiers:       keepString,
			},
			"code": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Code settings of code-based methods (read-only).",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"length":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Number of digits."},
					"validity": schema.Int64Attribute{Computed: true, MarkdownDescription: "Validity in seconds."},
				},
			},
			"created_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp.",
				PlanModifiers:       keepString,
			},
			"updated_time": schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		},
	}
}

// ValidateConfig checks the rules that need only the configuration. Rules that depend on the method's owner
// (system vs pluggable) run on apply, once the method has been looked up.
func (r *verificationMethodResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg MethodState
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if known(cfg.LocaleText) {
		for i, el := range cfg.LocaleText.Elements() {
			obj, ok := el.(types.Object)
			if !ok || !known(obj) {
				continue
			}
			if locale, ok := obj.Attributes()["locale"].(types.String); ok && !locale.IsUnknown() && strings.TrimSpace(locale.ValueString()) == "" {
				resp.Diagnostics.AddAttributeError(path.Root("locale_text").AtListIndex(i), "Missing locale", "Each `locale_text` entry needs a non-empty `locale`, e.g. `en`.")
			}
		}
	}
	if cfg.URL.IsNull() {
		return
	}
	if cfg.AuthType.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("auth_type"), "Missing auth_type", "`auth_type` is required with `url` (pluggable method).")
	}
	if cfg.CidaasAuthDetails.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("cidaas_auth_details"), "Missing cidaas_auth_details", "`cidaas_auth_details` is required with `url` (pluggable method).")
	}
	if !cfg.EncryptionConfig.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("encryption_config"), "encryption_config not allowed",
			"`encryption_config` can only be set on system methods, not on pluggable methods (`url` set).")
	}
}

// Create adopts an existing system method (PUT) or creates a pluggable method (POST).
func (r *verificationMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config MethodState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	verificationType := plan.VerificationType.ValueString()

	existing, err := r.client.VerificationMethod.Get(ctx, verificationType)
	switch {
	case err != nil && !errors.Is(err, client.ErrNotFound):
		resp.Diagnostics.AddError("Failed to look up verification method", err.Error())
		return
	case err == nil && existing.IsPluggable():
		resp.Diagnostics.AddError("Verification method already exists",
			fmt.Sprintf("Pluggable verification method %q already exists. Import it with: terraform import <resource address> %s", verificationType, verificationType))
		return
	case err == nil:
		// System method: it already exists and can only be configured.
		if checkSystemMethodConfig(&config, verificationType, &resp.Diagnostics); resp.Diagnostics.HasError() {
			return
		}
		model, diags := buildSystemMethodUpdate(ctx, &plan, &config, existing)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if _, err := r.client.VerificationMethod.Update(ctx, model); err != nil {
			resp.Diagnostics.AddError("Failed to configure verification method", err.Error())
			return
		}
	default:
		// Not found: create a pluggable method.
		if plan.URL.IsNull() || plan.URL.IsUnknown() {
			resp.Diagnostics.AddError("Verification method not found",
				fmt.Sprintf("Verification method %q does not exist on the tenant. To create a pluggable method, set url, auth_type and cidaas_auth_details.", verificationType))
			return
		}
		model, diags := buildPluggableMethod(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if _, err := r.client.VerificationMethod.Create(ctx, model); err != nil {
			resp.Diagnostics.AddError("Failed to create verification method", err.Error())
			return
		}
	}

	r.readInto(ctx, verificationType, &plan, &resp.State, &resp.Diagnostics)
}

// Read refreshes state from the API and removes the resource from state when the method no longer exists.
func (r *verificationMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MethodState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	api, err := r.client.VerificationMethod.Get(ctx, state.VerificationType.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read verification method", err.Error())
		return
	}
	next, diags := FlattenVerificationMethod(ctx, api, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update sends a system or pluggable PUT depending on the method's owner.
func (r *verificationMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state MethodState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	verificationType := plan.VerificationType.ValueString()

	var model client.VerificationMethodModel
	var diags diag.Diagnostics
	if strings.EqualFold(state.Owner.ValueString(), client.VerificationMethodOwnerClient) {
		if !config.EncryptionConfig.IsNull() {
			resp.Diagnostics.AddAttributeError(path.Root("encryption_config"), "encryption_config not allowed",
				fmt.Sprintf("%q is a pluggable method; encryption_config can only be set on system methods.", verificationType))
			return
		}
		model, diags = buildPluggableMethod(ctx, &plan)
	} else {
		if checkSystemMethodConfig(&config, verificationType, &resp.Diagnostics); resp.Diagnostics.HasError() {
			return
		}
		model, diags = buildSystemMethodUpdate(ctx, &plan, &config, nil)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.VerificationMethod.Update(ctx, model); err != nil {
		resp.Diagnostics.AddError("Failed to update verification method", err.Error())
		return
	}
	r.readInto(ctx, verificationType, &plan, &resp.State, &resp.Diagnostics)
}

// Delete removes a pluggable method. System methods cannot be deleted; they are only removed from state.
func (r *verificationMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MethodState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	verificationType := state.VerificationType.ValueString()
	if !strings.EqualFold(state.Owner.ValueString(), client.VerificationMethodOwnerClient) {
		resp.Diagnostics.AddWarning("System verification method not deleted",
			fmt.Sprintf("%q is a system verification method and cannot be deleted. It was removed from Terraform state; its settings on the tenant are unchanged.", verificationType))
		return
	}
	if err := r.client.VerificationMethod.Delete(ctx, verificationType); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Failed to delete verification method", err.Error())
	}
}

// ImportState imports a method by its verification_type.
func (r *verificationMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verification_type"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// readInto reads the method after a write (the write responses lack name/description) and stores it in state.
func (r *verificationMethodResource) readInto(ctx context.Context, verificationType string, plan *MethodState, state interface {
	Set(context.Context, interface{}) diag.Diagnostics
}, diags *diag.Diagnostics,
) {
	api, err := r.client.VerificationMethod.Get(ctx, verificationType)
	if err != nil {
		diags.AddError("Failed to read verification method after write", err.Error())
		return
	}
	next, d := FlattenVerificationMethod(ctx, api, plan)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	diags.Append(state.Set(ctx, &next)...)
}

// checkSystemMethodConfig rejects pluggable-only attributes on a system method.
func checkSystemMethodConfig(config *MethodState, verificationType string, diags *diag.Diagnostics) {
	values := map[string]attr.Value{
		"url": config.URL, "auth_type": config.AuthType, "cidaas_auth_details": config.CidaasAuthDetails,
		"prerequisite": config.Prerequisite, "factor": config.Factor, "category": config.Category,
		"icon": config.Icon, "amr_values": config.AMRValues, "acr_pin": config.ACRPin,
	}
	for _, name := range verificationMethodPluggableAttrs {
		if !values[name].IsNull() {
			diags.AddAttributeError(path.Root(name), "Attribute not allowed on system method",
				fmt.Sprintf("%q is a system verification method; `%s` can only be set on pluggable methods. "+
					"System methods support active, locale_text and encryption_config.", verificationType, name))
		}
	}
}

// buildSystemMethodUpdate builds the PUT for a system method. Unknown values (first adoption) fall back to the
// current server values, so locale texts are never wiped.
func buildSystemMethodUpdate(ctx context.Context, plan, config *MethodState, existing *client.VerificationMethodModel) (client.VerificationMethodModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.VerificationMethodModel{VerificationType: plan.VerificationType.ValueString()}

	if known(plan.Active) {
		v := plan.Active.ValueBool()
		model.Active = &v
	} else if existing != nil {
		model.Active = existing.Active
	}

	if known(plan.LocaleText) {
		model.LocaleText, diags = expandLocaleText(ctx, plan.LocaleText)
	} else if existing != nil {
		model.LocaleText = existing.LocaleText
	}

	if !config.EncryptionConfig.IsNull() && known(plan.EncryptionConfig) {
		attrs := plan.EncryptionConfig.Attributes()
		model.EncryptionConfig = &client.VerificationMethodEncryption{
			Enabled: attrs["enabled"].(types.Bool).ValueBool(),  //nolint:forcetypeassert
			Mode:    attrs["mode"].(types.String).ValueString(), //nolint:forcetypeassert
		}
	}
	return model, diags
}

// buildPluggableMethod builds the POST/PUT body for a pluggable method. The API replaces category, icon,
// amr_values, acr_pin and factor on every update, so the full plan is sent.
func buildPluggableMethod(ctx context.Context, plan *MethodState) (client.VerificationMethodModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.VerificationMethodModel{
		VerificationType: plan.VerificationType.ValueString(),
		URL:              stringIfKnown(plan.URL),
		AuthType:         stringIfKnown(plan.AuthType),
		Factor:           stringIfKnown(plan.Factor),
		Category:         stringIfKnown(plan.Category),
		Icon:             stringIfKnown(plan.Icon),
		ACR:              stringIfKnown(plan.ACRPin),
	}
	if known(plan.Active) {
		v := plan.Active.ValueBool()
		model.Active = &v
	}
	if known(plan.Prerequisite) {
		v := plan.Prerequisite.ValueString()
		model.Prerequisite = &v
	}
	if known(plan.CidaasAuthDetails) {
		model.CidaasAuthDetails = &client.VerificationMethodAuth{
			ClientID: plan.CidaasAuthDetails.Attributes()["client_id"].(types.String).ValueString(), //nolint:forcetypeassert
		}
	}
	if known(plan.AMRValues) {
		diags.Append(plan.AMRValues.ElementsAs(ctx, &model.AMR, false)...)
	}
	if known(plan.LocaleText) {
		var d diag.Diagnostics
		model.LocaleText, d = expandLocaleText(ctx, plan.LocaleText)
		diags.Append(d...)
	}
	return model, diags
}

type verificationMethodLocaleModel struct {
	Locale      types.String `tfsdk:"locale"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// expandLocaleText converts locale_text to the API list.
func expandLocaleText(ctx context.Context, l types.List) ([]client.VerificationMethodLocale, diag.Diagnostics) {
	var items []verificationMethodLocaleModel
	diags := l.ElementsAs(ctx, &items, false)
	out := make([]client.VerificationMethodLocale, 0, len(items))
	for _, it := range items {
		out = append(out, client.VerificationMethodLocale{
			Locale:      it.Locale.ValueString(),
			Name:        it.Name.ValueString(),
			Description: it.Description.ValueString(),
		})
	}
	return out, diags
}

// FlattenVerificationMethod maps an API method to Terraform. prior (plan or state, may be nil) keeps explicitly
// configured empty lists empty instead of null.
func FlattenVerificationMethod(_ context.Context, api *client.VerificationMethodModel, prior *MethodState) (MethodState, diag.Diagnostics) {
	var diags diag.Diagnostics
	if prior == nil {
		prior = &MethodState{}
	}
	out := MethodState{
		ID:                       types.StringValue(api.VerificationType),
		VerificationType:         types.StringValue(api.VerificationType),
		Active:                   types.BoolValue(api.Active != nil && *api.Active),
		URL:                      stringOrNull(api.URL),
		AuthType:                 stringOrNull(api.AuthType),
		Factor:                   stringOrNull(api.Factor),
		Category:                 stringOrNull(api.Category),
		Icon:                     stringOrNull(api.Icon),
		ACRPin:                   stringOrNull(api.ACR),
		Owner:                    stringOrNull(api.Owner),
		Name:                     stringOrNull(api.Name),
		Description:              stringOrNull(api.Description),
		RequiresAuthenticatorApp: types.BoolValue(api.RequiresAuthenticatorApp != nil && *api.RequiresAuthenticatorApp),
		AlgorithmTypeID:          stringOrNull(api.AlgorithmTypeID),
		CreatedTime:              millisecondTimestamp(api.CreatedTime),
		UpdatedTime:              millisecondTimestamp(api.UpdatedTime),
		Prerequisite:             types.StringNull(),
		AMRValues:                stringListOrPrior(api.AMR, prior.AMRValues),
		EncryptionConfig:         types.ObjectNull(verificationMethodEncryptionAttrTypes),
		CidaasAuthDetails:        types.ObjectNull(verificationMethodAuthAttrTypes),
		Code:                     types.ObjectNull(verificationMethodCodeAttrTypes),
	}
	if api.Prerequisite != nil {
		out.Prerequisite = types.StringValue(*api.Prerequisite)
	}

	locales := make([]attr.Value, 0, len(api.LocaleText))
	for _, l := range api.LocaleText {
		locales = append(locales, types.ObjectValueMust(verificationMethodLocaleType.AttrTypes, map[string]attr.Value{
			"locale":      types.StringValue(l.Locale),
			"name":        types.StringValue(l.Name),
			"description": types.StringValue(l.Description),
		}))
	}
	// Null for an empty API list, unless an explicit empty list was configured.
	configuredEmpty := known(prior.LocaleText) && len(prior.LocaleText.Elements()) == 0
	if len(locales) == 0 && !configuredEmpty {
		out.LocaleText = types.ListNull(verificationMethodLocaleType)
	} else {
		out.LocaleText = types.ListValueMust(verificationMethodLocaleType, locales)
	}

	if e := api.EncryptionConfig; e != nil {
		out.EncryptionConfig = types.ObjectValueMust(verificationMethodEncryptionAttrTypes, map[string]attr.Value{
			"enabled":          types.BoolValue(e.Enabled),
			"mode":             types.StringValue(e.Mode),
			"encrypted_fields": stringListOrPrior(e.EncryptedFields, types.ListNull(types.StringType)),
		})
	}
	if a := api.CidaasAuthDetails; a != nil && a.ClientID != "" {
		out.CidaasAuthDetails = types.ObjectValueMust(verificationMethodAuthAttrTypes, map[string]attr.Value{
			"client_id": types.StringValue(a.ClientID),
		})
	}
	if c := api.Code; c != nil {
		out.Code = types.ObjectValueMust(verificationMethodCodeAttrTypes, map[string]attr.Value{
			"length":   types.Int64Value(c.Length),
			"validity": types.Int64Value(c.Validity),
		})
	}
	return out, diags
}

// known reports whether v is neither null nor unknown.
func known(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// stringIfKnown returns the string value, or "" when null or unknown.
func stringIfKnown(v types.String) string {
	if !known(v) {
		return ""
	}
	return v.ValueString()
}

// stringListOrPrior returns null for an empty API list, unless an explicit empty list was configured.
func stringListOrPrior(values []string, prior basetypes.ListValue) types.List {
	if len(values) == 0 {
		if known(prior) && len(prior.Elements()) == 0 {
			return types.ListValueMust(types.StringType, []attr.Value{})
		}
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}

// millisecondTimestamp normalizes API timestamps to millisecond precision (what MongoDB stores), so values from
// write responses and later reads match.
func millisecondTimestamp(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return types.StringValue(v)
	}
	return types.StringValue(t.UTC().Truncate(time.Millisecond).Format(time.RFC3339Nano))
}
