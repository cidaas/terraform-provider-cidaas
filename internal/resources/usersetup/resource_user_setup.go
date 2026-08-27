// Package usersetup implements the cidaas_user_setup Terraform resource
// against user-srv (/user-srv/usersetup).
package usersetup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &userSetupResource{}
	_ resource.ResourceWithConfigure      = &userSetupResource{}
	_ resource.ResourceWithImportState    = &userSetupResource{}
	_ resource.ResourceWithValidateConfig = &userSetupResource{}

	communicationMediumVerificationValues = []string{
		"none",
		"mobile_and_email_verification_required",
		"email_verification_required",
		"mobile_verification_required",
		"email_verification_required_on_usage",
		"mobile_verification_required_on_usage",
		"verification_required_on_usage",
	}
	communicationMethodValues = []string{"email", "mobile_number", "phone_number"}
	invalidNameRunes          = []string{"/", `\`, "<", ">", "$", "+", "{", "}"}
)

// userSetupResource implements the Terraform resource for cidaas_user_setup.
type userSetupResource struct {
	client *client.Client
}

// setupConfig is the Terraform state/plan model for cidaas_user_setup.
type setupConfig struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ConsentRefs types.List   `tfsdk:"consent_refs"`
	UserSetup   types.Object `tfsdk:"user_setup"`

	userSetup *userSetupDetailConfig
}

// userSetupDetailConfig is the nested user_setup block in Terraform state.
type userSetupDetailConfig struct {
	AllowedFields                   types.List   `tfsdk:"allowed_fields"`
	RequiredFields                  types.List   `tfsdk:"required_fields"`
	AllowLoginWith                  types.List   `tfsdk:"allow_login_with"`
	EnableDeduplication             types.Bool   `tfsdk:"enable_deduplication"`
	ValidatePhoneNumber             types.Bool   `tfsdk:"validate_phone_number"`
	ValidateEmail                   types.Bool   `tfsdk:"validate_email"`
	AutoActivateUser                types.Bool   `tfsdk:"auto_activate_user"`
	AllowDisposableEmail            types.Bool   `tfsdk:"allow_disposable_email"`
	AcceptRolesInRegistration       types.Bool   `tfsdk:"accept_roles_in_the_registration"`
	SendWelcomeNotification         types.Bool   `tfsdk:"send_welcome_notification"`
	BirthdateAsDate                 types.Bool   `tfsdk:"birthdate_as_date"`
	CommunicationMediumVerification types.String `tfsdk:"communication_medium_verification"`
	AutoConfirmCommunicationMethod  types.List   `tfsdk:"auto_confirm_communication_method"`
	VerificationForMedium           types.List   `tfsdk:"verification_for_medium"`
	OperationsAllowedGroups         types.Set    `tfsdk:"operations_allowed_groups"`

	operationsAllowedGroups []*allowedGroupConfig
}

// allowedGroupConfig is one operations_allowed_groups element.
type allowedGroupConfig struct {
	GroupID      types.String `tfsdk:"group_id"`
	Roles        types.List   `tfsdk:"roles"`
	DefaultRoles types.List   `tfsdk:"default_roles"`
}

// NewUserSetupResource returns the cidaas_user_setup resource implementation.
func NewUserSetupResource() resource.Resource {
	return &userSetupResource{}
}

// Metadata sets the resource type name to cidaas_user_setup.
func (r *userSetupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_setup"
}

func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}

func allowedGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"group_id":      types.StringType,
		"roles":         types.ListType{ElemType: types.StringType},
		"default_roles": types.ListType{ElemType: types.StringType},
	}
}

func userSetupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_fields":                    types.ListType{ElemType: types.StringType},
		"required_fields":                   types.ListType{ElemType: types.StringType},
		"allow_login_with":                  types.ListType{ElemType: types.StringType},
		"enable_deduplication":              types.BoolType,
		"validate_phone_number":             types.BoolType,
		"validate_email":                    types.BoolType,
		"auto_activate_user":                types.BoolType,
		"allow_disposable_email":            types.BoolType,
		"accept_roles_in_the_registration":  types.BoolType,
		"send_welcome_notification":         types.BoolType,
		"birthdate_as_date":                 types.BoolType,
		"communication_medium_verification": types.StringType,
		"auto_confirm_communication_method": types.ListType{ElemType: types.StringType},
		"verification_for_medium":           types.ListType{ElemType: types.StringType},
		"operations_allowed_groups":         types.SetType{ElemType: types.ObjectType{AttrTypes: allowedGroupAttrTypes()}},
	}
}

// Schema defines the Terraform schema for cidaas_user_setup.
func (r *userSetupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages reusable User Setup profiles on cidaas v4 (Trustdesk) via `user-srv/usersetup`. " +
			"Applications reference a profile via `user_setup_id`.\n\n" +
			"Writes (create/update/delete) require admin or developer roles " +
			"(`USERSETUP_MANAGER` / `APP_MANAGER` / `ADMIN` / `SECONDARY_ADMIN` / `SUPER_ADMIN`) — there is no write OAuth scope. " +
			"Read-by-ID may use roles or scope `cidaas:usersetup_read`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server-assigned UUID of the user setup profile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique name of the profile (max 100 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					nameCharsValidator{},
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description of the profile.",
			},
			"consent_refs": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Consent template IDs presented during registration. Must be UUIDs. Stored in the API as `user_setup.consent_refs`.",
				Default:             listdefault.StaticValue(emptyStringList()),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(uuidStringValidator{}),
				},
			},
			"user_setup": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Registration and login configuration for this profile.",
				Attributes: map[string]schema.Attribute{
					"allowed_fields": schema.ListAttribute{
						Required:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Fields permitted during registration. Each key must exist and be enabled in Field Setup.",
					},
					"required_fields": schema.ListAttribute{
						Required:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Mandatory registration fields. Must be a subset of `allowed_fields`, and each key must exist and be enabled in Field Setup.",
					},
					"allow_login_with": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						MarkdownDescription: "Allowed login identifiers. Built-ins: `EMAIL`, `MOBILE`, `USER_NAME` (not checked against Field Setup). " +
							"Other values must exist and be enabled in Field Setup. Defaults to `[\"EMAIL\"]`.",
						Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{
							types.StringValue("EMAIL"),
						})),
					},
					"enable_deduplication": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable field deduplication checks.",
					},
					"validate_phone_number": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable phone number format validation.",
					},
					"validate_email": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable email domain validation.",
					},
					"auto_activate_user": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Auto-activate the user on registration.",
					},
					"allow_disposable_email": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Allow registrations with disposable email addresses.",
					},
					"accept_roles_in_the_registration": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Accept roles provided in the registration payload.",
					},
					"send_welcome_notification": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Send a welcome notification after registration.",
					},
					"birthdate_as_date": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Treat birthdate as a date type instead of a string.",
					},
					"communication_medium_verification": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Verification behavior for communication mediums.",
						Validators: []validator.String{
							stringvalidator.OneOf(communicationMediumVerificationValues...),
						},
					},
					"auto_confirm_communication_method": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Methods auto-confirmed on registration: `email`, `mobile_number`, `phone_number`.",
						Default:             listdefault.StaticValue(emptyStringList()),
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.OneOf(communicationMethodValues...)),
						},
					},
					"verification_for_medium": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Methods verified on registration: `email`, `mobile_number`, `phone_number`.",
						Default:             listdefault.StaticValue(emptyStringList()),
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.OneOf(communicationMethodValues...)),
						},
					},
					"operations_allowed_groups": schema.SetNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Groups allowed to perform setup operations.",
						Default: setdefault.StaticValue(types.SetValueMust(
							types.ObjectType{AttrTypes: allowedGroupAttrTypes()},
							[]attr.Value{},
						)),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"group_id": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "ID of the allowed group.",
								},
								"roles": schema.ListAttribute{
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									MarkdownDescription: "Roles allowed to perform operations.",
									Default:             listdefault.StaticValue(emptyStringList()),
								},
								"default_roles": schema.ListAttribute{
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									MarkdownDescription: "Default roles assigned to group members.",
									Default:             listdefault.StaticValue(emptyStringList()),
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
func (r *userSetupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (c *setupConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if c.UserSetup.IsNull() || c.UserSetup.IsUnknown() {
		return diags
	}
	c.userSetup = &userSetupDetailConfig{}
	diags.Append(c.UserSetup.As(ctx, c.userSetup, basetypes.ObjectAsOptions{})...)
	if diags.HasError() || c.userSetup.OperationsAllowedGroups.IsNull() || c.userSetup.OperationsAllowedGroups.IsUnknown() {
		return diags
	}
	diags.Append(c.userSetup.OperationsAllowedGroups.ElementsAs(ctx, &c.userSetup.operationsAllowedGroups, false)...)
	return diags
}

// ValidateConfig ensures required_fields is a subset of allowed_fields (plan-time, no API call).
func (r *userSetupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config setupConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.extract(ctx)...)
	if resp.Diagnostics.HasError() || config.userSetup == nil {
		return
	}
	if config.userSetup.AllowedFields.IsUnknown() || config.userSetup.RequiredFields.IsUnknown() {
		return
	}
	allowed, diags := listToStrings(ctx, config.userSetup.AllowedFields)
	resp.Diagnostics.Append(diags...)
	required, diags := listToStrings(ctx, config.userSetup.RequiredFields)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if missing := missingRequiredFields(allowed, required); len(missing) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("user_setup").AtName("required_fields"),
			"required_fields must be a subset of allowed_fields",
			fmt.Sprintf("These required_fields are not in allowed_fields: %s", strings.Join(missing, ", ")),
		)
	}
}

func missingRequiredFields(allowed, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(allowed))
	for _, f := range allowed {
		set[f] = struct{}{}
	}
	var missing []string
	for _, f := range required {
		if _, ok := set[f]; !ok {
			missing = append(missing, f)
		}
	}
	return missing
}

// isBuiltinAllowLogin mirrors user-srv: EMAIL / MOBILE / USER_NAME (case- and dash-insensitive).
func isBuiltinAllowLogin(key string) bool {
	switch strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(key), "-", "_")) {
	case "EMAIL", "MOBILE", "USER_NAME":
		return true
	default:
		return false
	}
}

func enabledFieldKeySet(fields []client.FieldSetupEntry) map[string]struct{} {
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if f.Enabled && f.FieldKey != "" {
			set[f.FieldKey] = struct{}{}
		}
	}
	return set
}

func missingFromEnabledSetup(values []string, enabledKeys map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}
	var missing []string
	for _, v := range values {
		if _, ok := enabledKeys[v]; !ok {
			missing = append(missing, v)
		}
	}
	return missing
}

func fieldSetupMissingMessage(fieldName string, missing []string) string {
	quoted := make([]string, 0, len(missing))
	for _, v := range missing {
		quoted = append(quoted, fmt.Sprintf(`"%q"`, v))
	}
	return fmt.Sprintf("%s[%s] are not available/enabled in the field setups", fieldName, strings.Join(quoted, ", "))
}

// fieldKeysMissingFromSetup mirrors user-srv validateFieldsExistence against enabled Field Setup keys.
func fieldKeysMissingFromSetup(allowed, required, allowLoginWith []string, fields []client.FieldSetupEntry) (string, string) {
	enabled := enabledFieldKeySet(fields)

	var loginCustom []string
	for _, v := range allowLoginWith {
		if !isBuiltinAllowLogin(v) {
			loginCustom = append(loginCustom, v)
		}
	}

	checks := []struct {
		name   string
		values []string
	}{
		{"allowed_fields", allowed},
		{"required_fields", required},
		{"allow_login_with", loginCustom},
	}
	for _, c := range checks {
		if missing := missingFromEnabledSetup(c.values, enabled); len(missing) > 0 {
			return c.name, fieldSetupMissingMessage(c.name, missing)
		}
	}
	return "", ""
}

func (c *setupConfig) toModel(ctx context.Context) (client.UserAppSetupModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.UserAppSetupModel{
		ID:          c.ID.ValueString(),
		Name:        c.Name.ValueString(),
		Description: c.Description.ValueString(),
	}
	if c.userSetup == nil {
		return model, diags
	}
	detail := client.UserSetupDetail{
		EnableDeduplication:       boolPtr(c.userSetup.EnableDeduplication),
		ValidatePhoneNumber:       boolPtr(c.userSetup.ValidatePhoneNumber),
		ValidateEmail:             boolPtr(c.userSetup.ValidateEmail),
		AutoActivateUser:          boolPtr(c.userSetup.AutoActivateUser),
		AllowDisposableEmail:      boolPtr(c.userSetup.AllowDisposableEmail),
		AcceptRolesInRegistration: boolPtr(c.userSetup.AcceptRolesInRegistration),
		SendWelcomeNotification:   boolPtr(c.userSetup.SendWelcomeNotification),
		BirthdateAsDate:           boolPtr(c.userSetup.BirthdateAsDate),
	}
	if !c.userSetup.CommunicationMediumVerification.IsNull() && !c.userSetup.CommunicationMediumVerification.IsUnknown() {
		detail.CommunicationMediumVerification = c.userSetup.CommunicationMediumVerification.ValueString()
	}
	var d diag.Diagnostics
	detail.AllowedFields, d = listToStrings(ctx, c.userSetup.AllowedFields)
	diags.Append(d...)
	detail.RequiredFields, d = listToStrings(ctx, c.userSetup.RequiredFields)
	diags.Append(d...)
	detail.AllowLoginWith, d = listToStrings(ctx, c.userSetup.AllowLoginWith)
	diags.Append(d...)
	detail.ConsentRefs, d = listToStrings(ctx, c.ConsentRefs)
	diags.Append(d...)
	detail.AutoConfirmCommunicationMethod, d = listToStrings(ctx, c.userSetup.AutoConfirmCommunicationMethod)
	diags.Append(d...)
	detail.VerificationForMedium, d = listToStrings(ctx, c.userSetup.VerificationForMedium)
	diags.Append(d...)

	for _, g := range c.userSetup.operationsAllowedGroups {
		if g == nil {
			continue
		}
		ag := client.AllowedGroup{GroupID: g.GroupID.ValueString()}
		ag.Roles, d = listToStrings(ctx, g.Roles)
		diags.Append(d...)
		ag.DefaultRoles, d = listToStrings(ctx, g.DefaultRoles)
		diags.Append(d...)
		detail.OperationsAllowedGroups = append(detail.OperationsAllowedGroups, ag)
	}
	model.UserSetup = detail
	return model, diags
}

func flattenUserSetup(model client.UserAppSetupModel) (setupConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := setupConfig{
		ID:          types.StringValue(model.ID),
		Name:        types.StringValue(model.Name),
		Description: stringValueOrNull(model.Description),
		ConsentRefs: stringList(model.UserSetup.ConsentRefs),
	}

	groups := make([]attr.Value, 0, len(model.UserSetup.OperationsAllowedGroups))
	for _, g := range model.UserSetup.OperationsAllowedGroups {
		obj, d := types.ObjectValue(allowedGroupAttrTypes(), map[string]attr.Value{
			"group_id":      types.StringValue(g.GroupID),
			"roles":         stringList(g.Roles),
			"default_roles": stringList(g.DefaultRoles),
		})
		diags.Append(d...)
		groups = append(groups, obj)
	}
	groupSet, d := types.SetValue(types.ObjectType{AttrTypes: allowedGroupAttrTypes()}, groups)
	diags.Append(d...)

	obj, d := types.ObjectValue(userSetupAttrTypes(), map[string]attr.Value{
		"allowed_fields":                    stringList(model.UserSetup.AllowedFields),
		"required_fields":                   stringList(model.UserSetup.RequiredFields),
		"allow_login_with":                  stringList(model.UserSetup.AllowLoginWith),
		"enable_deduplication":              boolValueOrNull(model.UserSetup.EnableDeduplication),
		"validate_phone_number":             boolValueOrNull(model.UserSetup.ValidatePhoneNumber),
		"validate_email":                    boolValueOrNull(model.UserSetup.ValidateEmail),
		"auto_activate_user":                boolValueOrNull(model.UserSetup.AutoActivateUser),
		"allow_disposable_email":            boolValueOrNull(model.UserSetup.AllowDisposableEmail),
		"accept_roles_in_the_registration":  boolValueOrNull(model.UserSetup.AcceptRolesInRegistration),
		"send_welcome_notification":         boolValueOrNull(model.UserSetup.SendWelcomeNotification),
		"birthdate_as_date":                 boolValueOrNull(model.UserSetup.BirthdateAsDate),
		"communication_medium_verification": stringValueOrNull(model.UserSetup.CommunicationMediumVerification),
		"auto_confirm_communication_method": stringList(model.UserSetup.AutoConfirmCommunicationMethod),
		"verification_for_medium":           stringList(model.UserSetup.VerificationForMedium),
		"operations_allowed_groups":         groupSet,
	})
	diags.Append(d...)
	cfg.UserSetup = obj
	return cfg, diags
}

// Create creates a user setup profile via POST /user-srv/usersetup.
func (r *userSetupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan setupConfig
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
	if !r.validateFieldSetupKeys(ctx, model, &resp.Diagnostics) {
		return
	}
	model.ID = ""

	res, err := r.client.UserSetup.Create(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Create user setup failed", err.Error())
		return
	}
	state, diags := flattenUserSetup(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state via GET /user-srv/usersetup/{id}. Removes from state on 404.
func (r *userSetupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state setupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.UserSetup.Get(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read user setup failed", err.Error())
		return
	}
	next, diags := flattenUserSetup(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update applies changes via PATCH /user-srv/usersetup/{id} (user-srv verb).
func (r *userSetupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan setupConfig
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
	if !r.validateFieldSetupKeys(ctx, model, &resp.Diagnostics) {
		return
	}

	res, err := r.client.UserSetup.Update(ctx, plan.ID.ValueString(), model)
	if err != nil {
		resp.Diagnostics.AddError("Update user setup failed", err.Error())
		return
	}
	state, diags := flattenUserSetup(res.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes the profile via DELETE /user-srv/usersetup/{id}.
func (r *userSetupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state setupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UserSetup.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete user setup failed", err.Error())
	}
}

// ImportState imports by server-assigned UUID (id).
func (r *userSetupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateFieldSetupKeys ensures allowed/required/custom login keys exist and are enabled in Field Setup.
func (r *userSetupResource) validateFieldSetupKeys(ctx context.Context, model client.UserAppSetupModel, diags *diag.Diagnostics) bool {
	fields, err := r.client.FieldSetup.List(ctx)
	if err != nil {
		diags.AddError("Failed to list field setups", err.Error())
		return false
	}
	attrName, msg := fieldKeysMissingFromSetup(
		model.UserSetup.AllowedFields,
		model.UserSetup.RequiredFields,
		model.UserSetup.AllowLoginWith,
		fields,
	)
	if attrName == "" {
		return true
	}
	diags.AddAttributeError(
		path.Root("user_setup").AtName(attrName),
		"Field keys must exist and be enabled in Field Setup",
		msg,
	)
	return false
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

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func boolValueOrNull(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

type uuidStringValidator struct{}

func (v uuidStringValidator) Description(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v uuidStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uuidStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	if _, err := uuid.Parse(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID", fmt.Sprintf("%q is not a valid UUID", req.ConfigValue.ValueString()))
	}
}

type nameCharsValidator struct{}

func (v nameCharsValidator) Description(_ context.Context) string {
	return "name must not contain /, \\, <, >, $, +, {, or }"
}

func (v nameCharsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nameCharsValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	name := req.ConfigValue.ValueString()
	for _, ch := range invalidNameRunes {
		if strings.Contains(name, ch) {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid name", fmt.Sprintf("name contains invalid character %q", ch))
			return
		}
	}
}
