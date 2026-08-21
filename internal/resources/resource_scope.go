package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type ScopeResource struct {
	BaseResource
}

func NewScopeResource() resource.Resource {
	return &ScopeResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_SCOPE,
				Schema: &scopeSchema,
			},
		),
	}
}

type ScopeConfig struct {
	ID                    types.String `tfsdk:"id"`
	SecurityLevel         types.String `tfsdk:"security_level"`
	ScopeKey              types.String `tfsdk:"scope_key"`
	GroupName             types.Set    `tfsdk:"group_name"`
	RequiredUserConsent   types.Bool   `tfsdk:"required_user_consent"`
	LocalizedDescription  types.List   `tfsdk:"localized_descriptions"`
	localizedDescriptions []*LocalDescription
	ScopeOwner            types.String `tfsdk:"scope_owner"`
}

type LocalDescription struct {
	Locale      types.String `tfsdk:"locale"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

func (sc *ScopeConfig) extractLocalizedDescription(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if !sc.LocalizedDescription.IsNull() {
		sc.localizedDescriptions = make([]*LocalDescription, 0, len(sc.LocalizedDescription.Elements()))
		diags = sc.LocalizedDescription.ElementsAs(ctx, &sc.localizedDescriptions, false)
	}
	return diags
}

var scopeSchema = schema.Schema{
	MarkdownDescription: "The Scope resource allows to manage scopes in Cidaas system. Scopes define the level of access and permissions granted to an application (client)." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:scopes_read" +
		"\n- cidaas:scopes_write" +
		"\n- cidaas:scopes_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"security_level": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The security level of the scope, e.g., `PUBLIC`. Allowed values are `PUBLIC` and `CONFIDENTIAL`",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"PUBLIC", "CONFIDENTIAL"}...),
			},
			Default: stringdefault.StaticString("PUBLIC"),
		},
		"scope_key": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Unique identifier for the scope. This cannot be updated for an existing state.",
			PlanModifiers: []planmodifier.String{
				// the destroy will throw the error too
				validators.UniqueIdentifier{},
			},
		},
		"group_name": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "List of scope_groups to associate the scope with.",
			// group_name validator can be added by fetching the group_names using api
		},
		"required_user_consent": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Indicates whether user consent is required for the scope.",
			Default:             booldefault.StaticBool(false),
		},
		"scope_owner": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The owner of the scope. e.g. `ADMIN`",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"localized_descriptions": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"locale": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The locale for the scope, e.g., `en-US`.",
						Default:             stringdefault.StaticString("en-US"),
						Validators: []validator.String{
							stringvalidator.OneOf(
								func() []string {
									validLocals := make([]string, len(util.Locales))
									for i, locale := range util.Locales {
										validLocals[i] = locale.LocaleString
									}
									return validLocals
								}()...),
						},
					},
					"title": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The title of the scope in the configured locale.",
					},
					"description": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The description of the scope in the configured locale.",
						Validators: []validator.String{
							stringvalidator.LengthBetween(0, 256),
						},
					},
				},
			},
		},
	},
}

func (r *ScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScopeConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extractLocalizedDescription(ctx)...)
	scopePayload, d := generateScopeModel(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare scope model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	response, err := r.cidaasClient.Scopes.Upsert(ctx, *scopePayload)
	if err != nil {
		tflog.Error(ctx, "failed to create scope via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create scope", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created scope via API", util.H{
		"scope_id": response.Data.ID,
	})

	plan.ScopeOwner = util.StringValueOrNull(&response.Data.ScopeOwner)
	plan.ID = util.StringValueOrNull(&response.Data.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource scope created successfully", util.H{
		"scope_id":  response.Data.ID,
		"scope_key": plan.ScopeKey.ValueString(),
	})
}

func (r *ScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScopeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.Scopes.Get(ctx, state.ScopeKey.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read scope via API", util.H{
			"scope_key": state.ScopeKey.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to read scope", util.FormatErrorMessage(err))
		return
	}

	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.ScopeKey = util.StringValueOrNull(&res.Data.ScopeKey)
	state.SecurityLevel = util.StringValueOrNull(&res.Data.SecurityLevel)
	state.RequiredUserConsent = util.BoolValueOrNull(&res.Data.RequiredUserConsent)
	state.ScopeOwner = util.StringValueOrNull(&res.Data.ScopeOwner)
	state.GroupName = util.SetValueOrNull(res.Data.GroupName)

	var diag diag.Diagnostics
	var objectValues []attr.Value
	localeDescription := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"locale":      types.StringType,
			"title":       types.StringType,
			"description": types.StringType,
		},
	}

	for _, sc := range res.Data.LocaleWiseDescription {
		local := sc.Locale
		title := sc.Title
		description := sc.Description
		objValue := types.ObjectValueMust(localeDescription.AttrTypes, map[string]attr.Value{
			"locale":      util.StringValueOrNull(&local),
			"title":       util.StringValueOrNull(&title),
			"description": util.StringValueOrNull(&description),
		})
		objectValues = append(objectValues, objValue)
	}

	state.LocalizedDescription, diag = types.ListValueFrom(ctx, localeDescription, objectValues)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to process localized descriptions", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully processed localized descriptions")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource scope read successfully", util.H{
		"scope_id": res.Data.ID,
	})
}

func (r *ScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl
	var plan, state ScopeConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(plan.extractLocalizedDescription(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or extract localized descriptions", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	scopePayload, d := generateScopeModel(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to generate scope model for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	_, err := r.cidaasClient.Scopes.Upsert(ctx, *scopePayload)
	if err != nil {
		tflog.Error(ctx, "Failed to update scope via API", util.H{
			"scope_key": plan.ScopeKey.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to update scope", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated scope via API", util.H{
		"scope_key": plan.ScopeKey.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource scope updated successfully", util.H{
		"scope_key": plan.ScopeKey.ValueString(),
	})
}

func (r *ScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state ScopeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.Scopes.Delete(ctx, state.ScopeKey.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete scope via API", util.H{
			"scope_key": state.ScopeKey.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete scope", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource scope deleted successfully", util.H{
		"scope_key": state.ScopeKey.ValueString(),
	})
}

func (r *ScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("scope_key"), req, resp)
}

func generateScopeModel(ctx context.Context, plan ScopeConfig) (*cidaas.ScopeModel, diag.Diagnostics) {
	scope := cidaas.ScopeModel{
		SecurityLevel:       plan.SecurityLevel.ValueString(),
		ScopeKey:            plan.ScopeKey.ValueString(),
		RequiredUserConsent: plan.RequiredUserConsent.ValueBool(),
	}
	diag := plan.GroupName.ElementsAs(ctx, &scope.GroupName, false)
	if diag.HasError() {
		return nil, diag
	}

	for _, ld := range plan.localizedDescriptions {
		scope.LocaleWiseDescription = append(scope.LocaleWiseDescription, cidaas.ScopeLocalDescription{
			Locale:      ld.Locale.ValueString(),
			Language:    util.GetLanguageForLocale(ld.Locale.ValueString()),
			Title:       ld.Title.ValueString(),
			Description: ld.Description.ValueString(),
		})
	}
	return &scope, nil
}
