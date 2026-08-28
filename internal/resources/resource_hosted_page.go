package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var allowedHotedPageIDs = []string{
	"register_success", "password_forgot_init", "verification_init", "verification_complete", "reactivate_verification_method",
	"device_init_code", "password_set", "password_set_success", "register_additional_info", "consent_preview", "mfa_required", "consent_scopes",
	"logout_success", "status", "group_selection", "login", "register", "error", "account_deduplication", "device_success_page",
	"suggest_verification_methods", "login_success",
}

const GroupOwner = "client"

type HostedPageResource struct {
	BaseResource
}

func NewHostedPageResource() resource.Resource {
	return &HostedPageResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_HOSTED_PAGE,
				Schema: &hostedPageSchema,
			},
		),
	}
}

type HostedPageConfig struct {
	ID                  types.String `tfsdk:"id"`
	HostedPageGroupName types.String `tfsdk:"hosted_page_group_name"`
	DefaultLocale       types.String `tfsdk:"default_locale"`
	HostedPages         types.Set    `tfsdk:"hosted_pages"`
	hostedPages         []*HostedPage
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

type HostedPage struct {
	HostedPageID types.String `tfsdk:"hosted_page_id"`
	Locale       types.String `tfsdk:"locale"`
	URL          types.String `tfsdk:"url"`
	Content      types.String `tfsdk:"content"`
}

func (h *HostedPageConfig) extractHostedPages(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if !h.HostedPages.IsNull() {
		h.hostedPages = make([]*HostedPage, 0, len(h.HostedPages.Elements()))
		diags = h.HostedPages.ElementsAs(ctx, &h.hostedPages, false)
	}
	return diags
}

var hostedPageSchema = schema.Schema{
	MarkdownDescription: "The Hosted Page resource in the provider allows you to define and manage hosted pages within the Cidaas system." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:hosted_pages_write" +
		"\n- cidaas:hosted_pages_read" +
		"\n- cidaas:hosted_pages_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"hosted_page_group_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the hosted page group. This must be unique across the cidaas system and cannot be updated for an existing state.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"default_locale": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The default locale for hosted pages e.g. `en-US`.",
			Default:             stringdefault.StaticString("en"),
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
			// if hosted_page not found by the local provided in the hosted_pages map, the api throws ambigious data error.
			// TODO: add a custom plan modifier later to validate the same and throw plan time error
		},
		"hosted_pages": schema.SetNestedAttribute{
			Required:            true,
			MarkdownDescription: "List of hosted pages with their respective attributes",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"hosted_page_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The identifier for the hosted page, e.g., `register_success`.",
						Validators: []validator.String{
							stringvalidator.OneOf(allowedHotedPageIDs...),
						},
					},
					"locale": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The locale for the hosted page, e.g., `en-US`.",
						Default:             stringdefault.StaticString("en"),
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
					"url": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The URL for the hosted page.",
					},
					"content": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The conent of the hosted page.",
					},
				},
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

func (r *HostedPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { //nolint:dupl
	var plan HostedPageConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extractHostedPages(ctx)...)
	hpPayload := prepareHostedPageModel(ctx, plan)
	res, err := r.cidaasClient.HostedPages.Upsert(ctx, *hpPayload)
	if err != nil {
		tflog.Error(ctx, "failed to create hosted page via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create hosted page", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created hosted page via API", util.H{
		"hosted_page_id": res.Data.ID,
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

	tflog.Info(ctx, "resource hosted page created successfully", util.H{
		"hosted_page_id": res.Data.ID,
	})
}

func (r *HostedPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HostedPageConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	res, err := r.cidaasClient.HostedPages.Get(ctx, state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read hosted page via API", util.H{
			"hosted_page_id": state.ID.ValueString(),
			"error":          err.Error(),
		})
		resp.Diagnostics.AddError("failed to read hosted page", util.FormatErrorMessage(err))
		return
	}

	// Update state with API response
	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.HostedPageGroupName = util.StringValueOrNull(&res.Data.ID)
	state.DefaultLocale = util.StringValueOrNull(&res.Data.DefaultLocale)
	state.CreatedAt = util.StringValueOrNull(&res.Data.CreatedTime)
	state.UpdatedAt = util.StringValueOrNull(&res.Data.UpdatedTime)

	hostedPages := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"hosted_page_id": types.StringType,
			"locale":         types.StringType,
			"url":            types.StringType,
			"content":        types.StringType,
		},
	}

	var objectValues []attr.Value
	for _, sc := range res.Data.HostedPages {
		hostedPageID := sc.HostedPageID
		local := sc.Locale
		url := sc.URL
		content := sc.Content
		objValue := types.ObjectValueMust(hostedPages.AttrTypes, map[string]attr.Value{
			"hosted_page_id": util.StringValueOrNull(&hostedPageID),
			"locale":         util.StringValueOrNull(&local),
			"url":            util.StringValueOrNull(&url),
			"content":        util.StringValueOrNull(&content),
		})
		objectValues = append(objectValues, objValue)
	}

	hps, diags := types.SetValueFrom(ctx, hostedPages, objectValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to process hosted pages data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	state.HostedPages = hps
	tflog.Debug(ctx, "successfully processed hosted pages data")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource hosted page read successfully", util.H{
		"hosted_page_id": state.ID.ValueString(),
	})
}

func (r *HostedPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl
	var plan, state HostedPageConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(plan.extractHostedPages(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	hpPayload := prepareHostedPageModel(ctx, plan)
	_, err := r.cidaasClient.HostedPages.Upsert(ctx, *hpPayload)
	if err != nil {
		resp.Diagnostics.AddError("failed to update hosted page", util.FormatErrorMessage(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HostedPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state HostedPageConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.HostedPages.Delete(ctx, state.ID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete hosted page via API", util.H{
			"hosted_page_id": state.ID.ValueString(),
			"error":          err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete hosted page", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource hosted page deleted successfully", util.H{
		"hosted_page_id": state.ID.ValueString(),
	})
}

func (r *HostedPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func prepareHostedPageModel(_ context.Context, plan HostedPageConfig) *cidaas.HostedPageModel {
	hostedPage := cidaas.HostedPageModel{
		ID:            plan.HostedPageGroupName.ValueString(),
		DefaultLocale: plan.DefaultLocale.ValueString(),
		GroupOwner:    GroupOwner,
	}
	var hps []cidaas.HostedPageData
	for _, hp := range plan.hostedPages {
		hps = append(hps, cidaas.HostedPageData{
			HostedPageID: hp.HostedPageID.ValueString(),
			Locale:       hp.Locale.ValueString(),
			URL:          hp.URL.ValueString(),
			Content:      hp.Content.ValueString(),
		})
	}
	hostedPage.HostedPages = hps
	return &hostedPage
}
