// Package usergroup manages cidaas user group resources.
package usergroup

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

//nolint:revive
type UserGroupResource struct {
	base.BaseResource
}

func NewUserGroupResource() resource.Resource {
	return &UserGroupResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_USER_GROUP,
				Schema: &userGroupSchema,
			},
		),
	}
}

//nolint:revive
type UserGroupConfig struct {
	ID                          types.String `tfsdk:"id"`
	GroupType                   types.String `tfsdk:"group_type"`
	GroupID                     types.String `tfsdk:"group_id"`
	GroupName                   types.String `tfsdk:"group_name"`
	ParentID                    types.String `tfsdk:"parent_id"`
	LogoURL                     types.String `tfsdk:"logo_url"`
	Description                 types.String `tfsdk:"description"`
	MakeFirstUserAdmin          types.Bool   `tfsdk:"make_first_user_admin"`
	MemberProfileVisibility     types.String `tfsdk:"member_profile_visibility"`
	NoneMemberProfileVisibility types.String `tfsdk:"none_member_profile_visibility"`
	CustomFields                types.Map    `tfsdk:"custom_fields"`
	CreatedAt                   types.String `tfsdk:"created_at"`
	UpdatedAt                   types.String `tfsdk:"updated_at"`
}

var userGroupSchema = schema.Schema{
	MarkdownDescription: "The cidaas_user_groups resource enables the creation of user groups in the cidaas system." +
		" These groups allow users to be organized and assigned group-specific roles." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:groups_write" +
		"\n- cidaas:groups_read" +
		"\n- cidaas:groups_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the user group resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_type": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Type of the user group.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"group_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Identifier for the user group.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			// ValidateFunc: validation.StringDoesNotContainAny(" "),
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"group_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the user group.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"parent_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Identifier of the parent user group.",
			Default:             stringdefault.StaticString("root"),
		},
		"logo_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "URL for the user group's logo",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Description of the user group.",
			Validators: []validator.String{
				stringvalidator.LengthAtMost(256),
			},
		},
		"make_first_user_admin": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Indicates whether the first user should be made an admin.",
			Default:             booldefault.StaticBool(false),
		},
		"member_profile_visibility": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Visibility of member profiles. Allowed values `public` or `full`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"public", "full"}...),
			},
			Default: stringdefault.StaticString("public"),
		},
		"none_member_profile_visibility": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Visibility of non-member profiles. Allowed values `none` or `public`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"none", "public"}...),
			},
			Default: stringdefault.StaticString("none"),
		},
		"custom_fields": schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Custom fields for the user group.",
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

func (r *UserGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	userGroup, diags := prepareUserGroupPayload(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare user_group payload", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.CidaasClient.UserGroup.Create(ctx, *userGroup)
	if err != nil {
		tflog.Error(ctx, "failed to create user group via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create user_group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created user group via API", util.H{
		"group_id": res.Data.ID,
	})

	plan.ID = types.StringValue(res.Data.ID)
	plan.GroupType = types.StringValue(res.Data.GroupType)
	plan.CreatedAt = types.StringValue(res.Data.CreatedTime)
	plan.UpdatedAt = types.StringValue(res.Data.UpdatedTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource user_group created successfully", util.H{
		"group_id": res.Data.ID,
	})
}

func (r *UserGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.CidaasClient.UserGroup.Get(ctx, state.GroupID.ValueString())
	if err != nil {
		if base.ReadHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read user_group via API", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to read user_group", util.FormatErrorMessage(err))
		return
	}

	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.GroupType = util.StringValueOrNull(&res.Data.GroupType)
	state.GroupID = util.StringValueOrNull(&res.Data.GroupID)
	state.GroupName = util.StringValueOrNull(&res.Data.GroupName)
	state.ParentID = util.StringValueOrNull(&res.Data.ParentID)
	state.LogoURL = util.StringValueOrNull(&res.Data.LogoURL)
	state.Description = util.StringValueOrNull(&res.Data.Description)
	state.MakeFirstUserAdmin = util.BoolValueOrNull(&res.Data.MakeFirstUserAdmin)
	state.MemberProfileVisibility = util.StringValueOrNull(&res.Data.MemberProfileVisibility)
	state.NoneMemberProfileVisibility = util.StringValueOrNull(&res.Data.NoneMemberProfileVisibility)
	state.CreatedAt = types.StringValue(res.Data.CreatedTime)
	state.UpdatedAt = types.StringValue(res.Data.UpdatedTime)

	cf, diags := util.MapValueOrNull(&res.Data.CustomFields)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to process custom fields", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	state.CustomFields = cf

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource user_group read successfully", util.H{
		"group_id":   res.Data.GroupID,
		"group_name": res.Data.GroupName,
	})
}

func (r *UserGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan or state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	userGroup, diags := prepareUserGroupPayload(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare user_group payload for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.CidaasClient.UserGroup.Update(ctx, *userGroup)
	if err != nil {
		tflog.Error(ctx, "failed to update user_group via API", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to update user_group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated user_group via API", util.H{
		"group_id": state.GroupID.ValueString(),
	})

	plan.GroupType = types.StringValue(res.Data.GroupType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "successfully completed user_group update", util.H{
		"group_id": state.GroupID.ValueString(),
	})
}

func (r *UserGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	subGroups, err := r.CidaasClient.UserGroup.GetSubGroups(ctx, state.GroupID.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to check sub user groups", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to check sub user groups", util.FormatErrorMessage(err))
		return
	}

	if len(subGroups) > 0 {
		tflog.Error(ctx, "Cannot delete user group with sub groups", util.H{
			"group_id":         state.GroupID.ValueString(),
			"sub_groups_count": len(subGroups),
		})
		resp.Diagnostics.AddError("Invalid Request", "The group contains sub user groups and cannot be deleted.")
		return
	}
	err = r.CidaasClient.UserGroup.Delete(ctx, state.GroupID.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to delete user group via API", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete user group", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource user_group deleted successfully", util.H{
		"group_id": state.GroupID.ValueString(),
	})
}

func (r *UserGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_id"), req, resp)
}

func prepareUserGroupPayload(ctx context.Context, plan UserGroupConfig) (*cidaas.UserGroupData, diag.Diagnostics) {
	userGroup := cidaas.UserGroupData{
		GroupType:                   plan.GroupType.ValueString(),
		GroupID:                     plan.GroupID.ValueString(),
		GroupName:                   plan.GroupName.ValueString(),
		ParentID:                    plan.ParentID.ValueString(),
		LogoURL:                     plan.LogoURL.ValueString(),
		Description:                 plan.Description.ValueString(),
		MakeFirstUserAdmin:          plan.MakeFirstUserAdmin.ValueBool(),
		MemberProfileVisibility:     plan.MemberProfileVisibility.ValueString(),
		NoneMemberProfileVisibility: plan.NoneMemberProfileVisibility.ValueString(),
	}
	diags := plan.CustomFields.ElementsAs(ctx, &userGroup.CustomFields, false)
	if diags.HasError() {
		return nil, diags
	}
	return &userGroup, nil
}
