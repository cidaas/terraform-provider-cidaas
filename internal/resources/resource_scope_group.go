package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ScopeGroupResource struct {
	BaseResource
}

func NewScopeGroupResource() resource.Resource {
	return &ScopeGroupResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_SCOPE_GROUP,
				Schema: &scopeGroupSchema,
			},
		),
	}
}

type ScopeGroupConfig struct {
	ID          types.String `tfsdk:"id"`
	GroupName   types.String `tfsdk:"group_name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

var scopeGroupSchema = schema.Schema{
	MarkdownDescription: "The cidaas_scope_group resource in the provider allows to manage Scope Groups in Cidaas system." +
		" Scope Groups help organize and group related scopes for better categorization and access control." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:scopes_read" +
		"\n- cidaas:scopes_write" +
		"\n- cidaas:scopes_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of th resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_name": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
			MarkdownDescription: "The name of the group. The group name must be unique across the cidaas system and cannot be updated for an existing state.",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The `description` attribute provides details about the scope of the group, explaining its purpose.",
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

func (r *ScopeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScopeGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	scopeGroup := cidaas.ScopeGroupConfig{
		GroupName:   plan.GroupName.ValueString(),
		Description: plan.Description.ValueString(),
	}
	res, err := r.cidaasClient.ScopeGroup.Upsert(ctx, scopeGroup)
	if err != nil {
		tflog.Error(ctx, "failed to create scope group via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create scope group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created scope group via API", util.H{
		"group_id": res.Data.ID,
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

	tflog.Info(ctx, "resource scope group created successfully", util.H{
		"group_id":   res.Data.ID,
		"group_name": plan.GroupName.ValueString(),
	})
}

func (r *ScopeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) { //nolint:dupl
	var state ScopeGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.ScopeGroup.Get(ctx, state.GroupName.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read scope group via API", util.H{
			"group_name": state.GroupName.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to read scope group", util.FormatErrorMessage(err))
		return
	}

	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	state.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)
	state.GroupName = util.StringValueOrNull(&res.Data.GroupName)
	state.Description = util.StringValueOrNull(&res.Data.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource scope group read successfully", util.H{
		"group_id":   res.Data.ID,
		"group_name": res.Data.GroupName,
	})
}

func (r *ScopeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ScopeGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan or state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	scopeGroup := cidaas.ScopeGroupConfig{
		GroupName:   plan.GroupName.ValueString(),
		Description: plan.Description.ValueString(),
	}
	_, err := r.cidaasClient.ScopeGroup.Upsert(ctx, scopeGroup)
	if err != nil {
		tflog.Error(ctx, "failed to update scope group via API", util.H{
			"group_name": plan.GroupName.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to update scope group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated scope group via API", util.H{
		"group_name": plan.GroupName.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource scope group updated successfully", util.H{
		"group_name": plan.GroupName.ValueString(),
	})
}

func (r *ScopeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state ScopeGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.ScopeGroup.Delete(ctx, state.GroupName.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete scope group via API", util.H{
			"group_name": state.GroupName.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete scope group", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource scope group deleted successfully", util.H{
		"group_name": state.GroupName.ValueString(),
	})
}

func (r *ScopeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_name"), req, resp)
}
