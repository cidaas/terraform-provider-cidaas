package resources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type PasswordPolicy struct {
	BaseResource
}

func NewPasswordPolicy() resource.Resource {
	return &PasswordPolicy{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_PASSWORD_POLICY,
				Schema: &passwordPolicySchema,
			},
		),
	}
}

type PasswordPolicyConfig struct {
	ID             types.String `tfsdk:"id"`
	PolicyName     types.String `tfsdk:"policy_name"`
	PasswordPolicy types.Object `tfsdk:"password_policy"`

	passwordPolicy *PasswordPolicyMap
}

type PasswordPolicyMap struct {
	BlockCompromised  types.Bool   `tfsdk:"block_compromised"`
	DenyUsageCount    types.Int64  `tfsdk:"deny_usage_count"`
	StrengthRegexes   types.Set    `tfsdk:"strength_regexes"`
	ChangeEnforcement types.Object `tfsdk:"change_enforcement"`

	changeEnforcement *ChangeEnforcement
}

type ChangeEnforcement struct {
	ExpirationInDays       types.Int64 `tfsdk:"expiration_in_days"`
	NotifyUserBeforeInDays types.Int64 `tfsdk:"notify_user_before_in_days"`
}

func (r *PasswordPolicyConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !r.PasswordPolicy.IsNull() && !r.PasswordPolicy.IsUnknown() {
		r.passwordPolicy = &PasswordPolicyMap{}
		diags = r.PasswordPolicy.As(ctx, r.passwordPolicy, basetypes.ObjectAsOptions{})
		if !r.passwordPolicy.ChangeEnforcement.IsNull() && !r.passwordPolicy.ChangeEnforcement.IsUnknown() {
			r.passwordPolicy.changeEnforcement = &ChangeEnforcement{}
			diags = r.passwordPolicy.ChangeEnforcement.As(ctx, &r.passwordPolicy.changeEnforcement, basetypes.ObjectAsOptions{})
		}
	}
	return diags
}

var passwordPolicySchema = schema.Schema{
	MarkdownDescription: "The Password Policy resource in the provider allows you to manage the password policy within the Cidaas." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:password_policy_read" +
		"\n- cidaas:password_policy_write" +
		"\n- cidaas:password_policy_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Unique identifier of the password policy.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"policy_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the password policy.",
		},
		"password_policy": schema.SingleNestedAttribute{
			Required:            true,
			MarkdownDescription: "The password policy configuration. All attributes are optional except strength_regexes. If not provided, default values will be applied.",
			Attributes: map[string]schema.Attribute{
				"block_compromised": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "Flag to block passwords that have been compromised.",
				},
				"deny_usage_count": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					Default:             int64default.StaticInt64(0),
					MarkdownDescription: "The reuse limit specifies the maximum number of times a user can reuse a previous password.",
				},
				"strength_regexes": schema.SetAttribute{
					Required:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "The regular expression to enforce the minimum and maximum character count, minimum number of numeric and special characters and whether to include lowercase or uppercase letters in a password.",
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
				},
				"change_enforcement": schema.SingleNestedAttribute{
					Optional: true,
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"expiration_in_days": schema.Int64Attribute{
							MarkdownDescription: "The number of days allowed before a password must be changed.",
							Optional:            true,
						},
						"notify_user_before_in_days": schema.Int64Attribute{
							MarkdownDescription: "Number of days before password expiry to notify the user.",
							Optional:            true,
						},
					},
					Default: objectdefault.StaticValue(
						types.ObjectValueMust(
							map[string]attr.Type{
								"expiration_in_days":         types.Int64Type,
								"notify_user_before_in_days": types.Int64Type,
							},
							map[string]attr.Value{
								"expiration_in_days":         types.Int64Value(0),
								"notify_user_before_in_days": types.Int64Value(0),
							},
						),
					),
				},
			},
		},
	},
}

func (r *PasswordPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PasswordPolicyConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	payload := cidaas.PasswordPolicyModel{
		PolicyName: plan.PolicyName.ValueString(),
		PasswordPolicy: &cidaas.Policy{
			BlockCompromised: plan.passwordPolicy.BlockCompromised.ValueBool(),
			DenyUsageCount:   plan.passwordPolicy.DenyUsageCount.ValueInt64(),
		},
	}

	diags := plan.passwordPolicy.StrengthRegexes.ElementsAs(ctx, &payload.PasswordPolicy.StrengthRegexes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to extract strength regexes", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	if !plan.passwordPolicy.ChangeEnforcement.IsNull() {
		payload.PasswordPolicy.ChangeEnforcement = cidaas.ChangeEnforcement{
			ExpirationInDays:       plan.passwordPolicy.changeEnforcement.ExpirationInDays.ValueInt64(),
			NotifyUserBeforeInDays: plan.passwordPolicy.changeEnforcement.NotifyUserBeforeInDays.ValueInt64(),
		}
	}

	res, err := r.cidaasClient.PasswordPolicy.Create(ctx, payload)
	if err != nil {
		tflog.Error(ctx, "failed to create password policy via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create password policy", fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "successfully created password policy via API", util.H{
		"policy_id": res.Data.ID,
	})

	plan.ID = types.StringValue(res.Data.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource password policy created successfully", util.H{
		"policy_id":   res.Data.ID,
		"policy_name": plan.PolicyName.ValueString(),
	})
}

func (r *PasswordPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PasswordPolicyConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.cidaasClient.PasswordPolicy.Get(ctx, state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read password policy via API", util.H{
			"policy_id": state.ID.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to read password policy", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	state.PolicyName = types.StringValue(res.Data.PolicyName)

	changeEnforcementConfig := map[string]attr.Type{
		"expiration_in_days":         types.Int64Type,
		"notify_user_before_in_days": types.Int64Type,
	}
	passwordPolicyType := map[string]attr.Type{
		"block_compromised":  types.BoolType,
		"deny_usage_count":   types.Int64Type,
		"strength_regexes":   types.SetType{ElemType: types.StringType},
		"change_enforcement": types.ObjectType{AttrTypes: changeEnforcementConfig},
	}

	if res.Data.PasswordPolicy != nil {
		tflog.Debug(ctx, "processing password policy configuration", util.H{
			"block_compromised": res.Data.PasswordPolicy.BlockCompromised,
			"deny_usage_count":  res.Data.PasswordPolicy.DenyUsageCount,
		})

		passwordPolicy, diags := types.ObjectValue(passwordPolicyType, map[string]attr.Value{
			"block_compromised": util.BoolValueOrNull(&res.Data.PasswordPolicy.BlockCompromised),
			"deny_usage_count":  util.Int64ValueOrNull(&res.Data.PasswordPolicy.DenyUsageCount),
			"strength_regexes":  util.SetValueOrNull(res.Data.PasswordPolicy.StrengthRegexes),
			"change_enforcement": types.ObjectValueMust(changeEnforcementConfig, map[string]attr.Value{
				"expiration_in_days":         util.Int64ValueOrNull(&res.Data.PasswordPolicy.ChangeEnforcement.ExpirationInDays),
				"notify_user_before_in_days": util.Int64ValueOrNull(&res.Data.PasswordPolicy.ChangeEnforcement.NotifyUserBeforeInDays),
			}),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "failed to process password policy configuration", util.H{
				"errors": resp.Diagnostics.Errors(),
			})
			return
		}
		state.PasswordPolicy = passwordPolicy
		tflog.Debug(ctx, "successfully processed password policy configuration")
	} else {
		tflog.Debug(ctx, "no password policy configuration found in response")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "successfully completed password policy read", util.H{
		"policy_id":   state.ID.ValueString(),
		"policy_name": res.Data.PolicyName,
	})
}

func (r *PasswordPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PasswordPolicyConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or extract configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	payload := cidaas.PasswordPolicyModel{
		ID:         state.ID.ValueString(),
		PolicyName: plan.PolicyName.ValueString(),
		PasswordPolicy: &cidaas.Policy{
			BlockCompromised: plan.passwordPolicy.BlockCompromised.ValueBool(),
			DenyUsageCount:   plan.passwordPolicy.DenyUsageCount.ValueInt64(),
		},
	}

	diags := plan.passwordPolicy.StrengthRegexes.ElementsAs(ctx, &payload.PasswordPolicy.StrengthRegexes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to extract strength regexes for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	if !plan.passwordPolicy.ChangeEnforcement.IsNull() {
		payload.PasswordPolicy.ChangeEnforcement = cidaas.ChangeEnforcement{
			ExpirationInDays:       plan.passwordPolicy.changeEnforcement.ExpirationInDays.ValueInt64(),
			NotifyUserBeforeInDays: plan.passwordPolicy.changeEnforcement.NotifyUserBeforeInDays.ValueInt64(),
		}
	}

	res, err := r.cidaasClient.PasswordPolicy.Update(ctx, payload)
	if err != nil {
		tflog.Error(ctx, "Failed to update password policy via API", util.H{
			"policy_id": state.ID.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to update password policy", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	if !res.Data {
		tflog.Error(ctx, "Password policy update returned unsuccessful response", util.H{
			"policy_id": state.ID.ValueString(),
			"response":  fmt.Sprintf("%+v", res),
		})
		resp.Diagnostics.AddError("failed to update password policy", fmt.Sprintf("Response: %+v", res))
		return
	}
	tflog.Info(ctx, "Successfully updated password policy via API", util.H{
		"policy_id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "resource password policy updated successfully", util.H{
		"policy_id":   state.ID.ValueString(),
		"policy_name": plan.PolicyName.ValueString(),
	})
}

func (r *PasswordPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state PasswordPolicyConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.PasswordPolicy.Delete(ctx, state.ID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete password policy via API", util.H{
			"policy_id": state.ID.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete password policy", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "password policy deleted successfully", util.H{
		"policy_id": state.ID.ValueString(),
	})
}

func (r *PasswordPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
