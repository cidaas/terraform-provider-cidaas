package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type RoleResource struct {
	BaseResource
}

func NewRoleResource() resource.Resource {
	return &RoleResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_ROLE,
				Schema: &roleSchema,
			},
		),
	}
}

type Role struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Role        types.String `tfsdk:"role"`
}

var roleSchema = schema.Schema{
	MarkdownDescription: "The cidaas_role resource in Terraform facilitates the management of roles in Cidaas system." +
		" This resource allows you to configure and define custom roles to suit your application's specific access control requirements." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:roles_read" +
		"\n- cidaas:roles_write" +
		"\n- cidaas:roles_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the role resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role": schema.StringAttribute{
			Required:    true,
			Description: "The unique identifier of the role. The role name must be unique across the cidaas system and cannot be updated for an existing state.",
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"name": schema.StringAttribute{
			Optional:    true,
			Description: "The name of the role.",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The `description` attribute provides details about the role, explaining its purpose.",
		},
	},
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Role
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	role := cidaas.RoleModel{
		Name:        plan.Name.ValueString(),
		Role:        plan.Role.ValueString(),
		Description: plan.Description.ValueString(),
	}
	response, err := r.cidaasClient.Roles.UpsertRole(ctx, role)
	if err != nil {
		tflog.Error(ctx, "failed to create role via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create role", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created role via API", util.H{
		"role_id": response.Data.Role,
	})

	plan.ID = util.StringValueOrNull(&response.Data.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource role created successfully", util.H{
		"role_id":   response.Data.Role,
		"role_name": plan.Name.ValueString(),
	})
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Role
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	response, err := r.cidaasClient.Roles.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read role via API", util.H{
			"role_id": state.ID.ValueString(),
			"error":   err.Error(),
		})
		resp.Diagnostics.AddError("failed to read role", util.FormatErrorMessage(err))
		return
	}

	state.Role = util.StringValueOrNull(&response.Data.Role)
	state.Description = util.StringValueOrNull(&response.Data.Description)
	state.Name = util.StringValueOrNull(&response.Data.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource role read successfully", util.H{
		"role_id":   state.ID.ValueString(),
		"role_name": response.Data.Name,
	})
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Role
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan or state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	role := cidaas.RoleModel{
		Name:        plan.Name.ValueString(),
		Role:        plan.Role.ValueString(),
		Description: plan.Description.ValueString(),
	}

	response, err := r.cidaasClient.Roles.UpsertRole(ctx, role)
	if err != nil {
		tflog.Error(ctx, "failed to update role via API", util.H{
			"role_id": state.ID.ValueString(),
			"error":   err.Error(),
		})
		resp.Diagnostics.AddError("failed to update role", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated role via API", util.H{
		"role_id": state.ID.ValueString(),
	})

	plan.ID = util.StringValueOrNull(&response.Data.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource role updated successfully", util.H{
		"role_id":   state.ID.ValueString(),
		"role_name": plan.Name.ValueString(),
	})
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state Role
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.Roles.DeleteRole(ctx, state.ID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete role via API", util.H{
			"role_id": state.ID.ValueString(),
			"error":   err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete role", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource role deleted successfully", util.H{
		"role_id": state.ID.ValueString(),
	})
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
