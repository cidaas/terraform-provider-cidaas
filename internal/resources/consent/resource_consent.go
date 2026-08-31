package consent

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
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

type ConsentConfig struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConsentGroupID types.String `tfsdk:"consent_group_id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

type ConsentResource struct {
	base.BaseResource
	cidaasClient *cidaas.Client
}

func NewConsentResource() resource.Resource {
	return &ConsentResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_CONSENT,
				Schema: &consentSchema,
			},
		),
	}
}

var consentSchema = schema.Schema{
	MarkdownDescription: "The consent resource in the provider is used to define and manage consents within the Cidaas system." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:consent_read" +
		"\n- cidaas:consent_write" +
		"\n- cidaas:consent_delete",
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
			MarkdownDescription: "The flag to enable or disable a specific consent. By default, the value is set to `true`",
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

func (r *ConsentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	if c.Capabilities.TargetVersion == "4.x" {
		resp.Diagnostics.AddError(
			"Consent Resource Coming Soon on v4",
			"Consent resources (`cidaas_consent`, `cidaas_consent_group`, `cidaas_consent_version`) are supported on cidaas v3.x. Support for cidaas v4.x (Trustdesk) is coming soon. Please set cidaas_version = \"3.x\" in your provider block to manage v3 consent resources.",
		)
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
	r.cidaasClient = c.CidaasClient
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
		if base.ReadHandleNotFound(ctx, resp, err) {
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

func (r *ConsentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
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
