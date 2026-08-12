// previously user_group_category
package resources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

type GroupTypeResource struct {
	BaseResource
}

func NewGroupTypeResource() resource.Resource {
	return &GroupTypeResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_GROUP_TYPE,
				Schema: &groupTypeSchema,
			},
		),
	}
}

type GroupTypeConfig struct {
	ID           types.String `tfsdk:"id"`
	RoleMode     types.String `tfsdk:"role_mode"`
	Description  types.String `tfsdk:"description"`
	GroupType    types.String `tfsdk:"group_type"`
	AllowedRoles types.Set    `tfsdk:"allowed_roles"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

var groupTypeSchema = schema.Schema{
	MarkdownDescription: "The Group Type, managed through the `cidaas_group_type` resource in the provider defines and configures categories for user groups within the Cidaas system." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:group_type_read" +
		"\n- cidaas:group_type_write" +
		"\n- cidaas:group_type_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role_mode": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Determines the role mode for the user group type. Allowed values are `any_roles`, `no_roles`, `roles_required` and `allowed_roles`",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"any_roles", "no_roles", "roles_required", "allowed_roles"}...),
			},
		},
		// TODO: description is a required parameter in admin ui
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The `description` attribute provides details about the group type, explaining its purpose.",
		},
		"group_type": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The unique identifier of the group type. This cannot be updated for an existing state.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"allowed_roles": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "List of allowed roles in this group type.",
			Validators: []validator.Set{
				&allowedRolesValidator{},
			},
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

func (r *GroupTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupTypeConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	groupType := cidaas.GroupTypeData{
		RoleMode:    plan.RoleMode.ValueString(),
		GroupType:   plan.GroupType.ValueString(),
		Description: plan.Description.ValueString(),
	}

	diag := plan.AllowedRoles.ElementsAs(ctx, &groupType.AllowedRoles, false)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to extract allowed roles", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.cidaasClient.GroupType.Create(ctx, groupType)
	if err != nil {
		tflog.Error(ctx, "failed to create group type via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create group type", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created group type via API", util.H{
		"group_type_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	plan.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	plan.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "resource group type created successfully")
}

func (r *GroupTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupTypeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.cidaasClient.GroupType.Get(ctx, state.GroupType.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read group type via API", util.H{
			"group_type": state.GroupType.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to read group type", util.FormatErrorMessage(err))
		return
	}

	// Update state with API response
	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	state.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)
	state.RoleMode = util.StringValueOrNull(&res.Data.RoleMode)
	state.GroupType = util.StringValueOrNull(&res.Data.GroupType)
	state.Description = util.StringValueOrNull(&res.Data.Description)
	state.AllowedRoles = util.SetValueOrNull(res.Data.AllowedRoles)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource group_type read successfully", util.H{
		"group_type_id": res.Data.ID,
		"group_type":    res.Data.GroupType,
	})
}

func (r *GroupTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GroupTypeConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan or state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	groupType := cidaas.GroupTypeData{
		RoleMode:    plan.RoleMode.ValueString(),
		GroupType:   plan.GroupType.ValueString(),
		Description: plan.Description.ValueString(),
	}

	diag := plan.AllowedRoles.ElementsAs(ctx, &groupType.AllowedRoles, false)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to extract allowed roles for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.GroupType.Update(ctx, groupType)
	if err != nil {
		tflog.Error(ctx, "failed to update group_type via API", util.H{
			"group_type_id": state.ID.ValueString(),
			"error":         err.Error(),
		})
		resp.Diagnostics.AddError("failed to update group_type", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated group_type via API", util.H{
		"group_type_id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource group_type updated successfully", util.H{
		"group_type_id": state.ID.ValueString(),
		"group_type":    plan.GroupType.ValueString(),
	})
}

func (r *GroupTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupTypeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.GroupType.Delete(ctx, state.GroupType.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete group type via API", util.H{
			"group_type": state.GroupType.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete group type", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource group_type deleted successfully", util.H{
		"group_type": state.GroupType.ValueString(),
	})
}

func (r *GroupTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_type"), req, resp)
}

type allowedRolesValidator struct{}

func (v allowedRolesValidator) Description(_ context.Context) string {
	return "Validates allowed_roles if required"
}

func (v allowedRolesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v allowedRolesValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.Config.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var (
		roleMode     string
		allowedRoles []types.String
		diags        diag.Diagnostics
	)

	diags = req.ConfigValue.ElementsAs(ctx, &allowedRoles, true)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.GetAttribute(ctx, path.Root("role_mode"), &roleMode)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	allowedRolesRequired := util.Contains([]string{"roles_required", "allowed_roles"}, roleMode)

	// Filter out null elements, but keep unknown elements (resource references) as valid
	var hasAllowedRoles bool
	for _, r := range allowedRoles {
		if !r.IsNull() {
			hasAllowedRoles = true
			break
		}
	}

	if allowedRolesRequired && !hasAllowedRoles {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("The attribute allowed_roles cannot be empty when role_mode is set to %s", roleMode))
	} else if !allowedRolesRequired && hasAllowedRoles {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("The attribute allowed_roles must be empty or omitted when role_mode is set to %s", roleMode))
	}
}
