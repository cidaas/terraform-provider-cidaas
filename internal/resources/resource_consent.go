package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ConsentResource struct {
	BaseResource
}

func NewConsentResource() resource.Resource {
	return &ConsentResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_CONSENT,
				Schema: &consentSchema,
			},
		),
	}
}

type ConsentConfig struct {
	ID             types.String `tfsdk:"id"`
	ConsentGroupID types.String `tfsdk:"consent_group_id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

var consentSchema = schema.Schema{
	MarkdownDescription: "The Consent resource in the provider allows you to manage different consents within a specific consent group in Cidaas." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:tenant_consent_read" +
		"\n- cidaas:tenant_consent_write" +
		"\n- cidaas:tenant_consent_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier of the consent resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required: true,
			MarkdownDescription: "The name of the consent. Cidaas stores this value in lowercase; " +
				"uppercase input is accepted and normalized on plan.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				validators.ToLower{},
				validators.CaseInsensitiveUniqueIdentifier{},
			},
		},
		"consent_group_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The `consent_group_id` to which the consent belongs.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The flag to enable or disable a speicific consent. By default, the value is set to `true`",
			Default:             booldefault.StaticBool(true),
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The timestamp when the consent version was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The timestamp when the consent version was last updated.",
		},
	},
}

func (r *ConsentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConsentConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	consent := cidaas.ConsentModel{
		ConsentName:    plan.Name.ValueString(),
		ConsentGroupID: plan.ConsentGroupID.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
	}

	res, err := r.cidaasClient.Consent.Upsert(ctx, consent)
	if err != nil {
		tflog.Error(ctx, "failed to create consent via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create consent", fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "successfully created consent via API", util.H{
		"consent_id": res.Data.ID,
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

	tflog.Info(ctx, "resource consent created successfully")
}

func (r *ConsentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConsentConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.Consent.GetConsentInstances(ctx, state.ConsentGroupID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read consent instances via API", util.H{
			"consent_group_id": state.ConsentGroupID.ValueString(),
			"error":            err.Error(),
		})
		resp.Diagnostics.AddError("Failed to read consent", fmt.Sprintf("Error: %s ", err.Error()))
		return
	}

	if !res.Success && res.Status == http.StatusNoContent {
		tflog.Error(ctx, "invalid consent group ID - no consent group found", util.H{
			"consent_group_id": state.ConsentGroupID.ValueString(),
			"status":           res.Status,
		})
		resp.Diagnostics.AddError("Invalid consent_group_id", fmt.Sprintf("No consent group found for the provided consent_group_id %+v", state.ConsentGroupID.String()))
		return
	}

	tflog.Debug(ctx, "processing consent instances")
	isAvailable := false
	if len(res.Data) > 0 {
		for _, instance := range res.Data {
			if strings.EqualFold(instance.ConsentName, state.Name.ValueString()) {
				isAvailable = true
				id := instance.ID
				consentName := instance.ConsentName
				createdTime := instance.CreatedTime
				updatedTime := instance.UpdatedTime
				state.ID = util.StringValueOrNull(&id)
				state.Name = util.StringValueOrNull(&consentName)
				state.Enabled = types.BoolValue(instance.Enabled)
				state.CreatedAt = util.StringValueOrNull(&createdTime)
				state.UpdatedAt = util.StringValueOrNull(&updatedTime)
				break
			}
		}
	}

	if !isAvailable {
		tflog.Error(ctx, "consent not found in consent group", util.H{
			"consent_name":     state.Name.ValueString(),
			"consent_group_id": state.ConsentGroupID.ValueString(),
		})
		resp.Diagnostics.AddError("Consent not found", fmt.Sprintf("consent %s not found for the provided consent_group_id %s", state.Name.String(), state.ConsentGroupID.String()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource consent read successfully")
}

func (r *ConsentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConsentConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan or state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	consent := cidaas.ConsentModel{
		ID:             state.ID.ValueString(),
		ConsentName:    plan.Name.ValueString(),
		ConsentGroupID: plan.ConsentGroupID.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
	}
	res, err := r.cidaasClient.Consent.Upsert(ctx, consent)
	if err != nil {
		tflog.Error(ctx, "failed to update consent via API", util.H{
			"consent_id": state.ID.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to update consent", fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "successfully updated consent via API", util.H{
		"consent_id": state.ID.ValueString(),
	})

	plan.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "resource consent updated successfully")
}

func (r *ConsentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConsentConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	err := r.cidaasClient.Consent.Delete(ctx, state.ID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete consent via API", util.H{
			"consent_id": state.ID.ValueString(),
			"error":      err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete consent", fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "resource consent deleted successfully", util.H{
		"consent_id": state.ID.ValueString(),
	})
}

func (r *ConsentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	tflog.Debug(ctx, "processing import identifier", util.H{
		"import_id": id,
	})

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		tflog.Error(ctx, "invalid import identifier format", util.H{
			"import_id": id,
		})
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'consent_group_id:name', got: %s", id),
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("consent_group_id"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("name"), parts[1])
}
