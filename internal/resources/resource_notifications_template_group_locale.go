package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NotificationsTemplateGroupLocaleResource struct {
	BaseResource
}

func NewNotificationsTemplateGroupLocaleResource() resource.Resource {
	return &NotificationsTemplateGroupLocaleResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_NOTIFICATIONS_TEMPLATE_GROUP_LOCALE, //nolint:revive
				Schema: &notificationsTemplateGroupLocaleSchema,
			},
		),
	}
}

var notificationsTemplateGroupLocaleSchema = schema.Schema{
	MarkdownDescription: "Manages **one locale** in a notification-srv template group: copy templates on create and bulk-delete on destroy.\n\n" +
		"Use with **`cidaas_notifications_template_group`** (group metadata only). After group create, notification-srv may already have locales from `default`; " +
		"if copy returns \"already templates found\", create succeeds when the locale appears in `templatefilters`.\n\n" +
		"Extra API locales not managed by locale resources are **not** auto-removed.\n\n" +
		"**Scopes:** `cidaas:templates_read`, `cidaas:templates_write`, `cidaas:templates_delete`.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "`{group_id}/{locale}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Template group id (`_id`).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"locale": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Target locale in the group (BCP47, e.g. `en`, `de-DE`).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"copy_from_group_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("default"),
			MarkdownDescription: "Source group for template copy (API `copy.fromGroupID`).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"copy_from_locale": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("en"),
			MarkdownDescription: "Source locale on `copy_from_group_id` (API `copy.locale[].from`).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}

type notificationsTemplateGroupLocaleModel struct {
	ID              types.String `tfsdk:"id"`
	GroupID         types.String `tfsdk:"group_id"`
	Locale          types.String `tfsdk:"locale"`
	CopyFromGroupID types.String `tfsdk:"copy_from_group_id"`
	CopyFromLocale  types.String `tfsdk:"copy_from_locale"`
}

func (r *NotificationsTemplateGroupLocaleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationsTemplateGroupLocaleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := plan.GroupID.ValueString()
	locale := plan.Locale.ValueString()

	localePresent, err := r.localeExistsInGroup(ctx, groupID, locale)
	if err != nil {
		if util.IsResourceNotFound(err) {
			resp.Diagnostics.AddError(
				"template group not found",
				fmt.Sprintf("group %q does not exist: %s", groupID, util.FormatErrorMessage(err)),
			)
			return
		}
		resp.Diagnostics.AddError("failed to read template group templatefilters", util.FormatErrorMessage(err))
		return
	}
	if !localePresent {
		copyReq := cidaas.NotificationsSrvCopy{
			FromGroupID: plan.CopyFromGroupID.ValueString(),
			Locale: []cidaas.NotificationsSrvLocaleMapping{
				{From: plan.CopyFromLocale.ValueString(), To: locale},
			},
		}
		if err := r.cidaasClient.NotificationsSrvTemplateGroup.CopyLocales(ctx, groupID, copyReq); err != nil {
			if !cidaas.IsNotificationSrvTemplatesAlreadyExistError(err) {
				resp.Diagnostics.AddError("failed to copy templates for locale", util.FormatErrorMessage(err))
				return
			}
			ok, verifyErr := r.localeExistsInGroup(ctx, groupID, locale)
			if verifyErr != nil {
				resp.Diagnostics.AddError("failed to verify locale after copy error", util.FormatErrorMessage(verifyErr))
				return
			}
			if !ok {
				resp.Diagnostics.AddError(
					"failed to copy templates for locale",
					fmt.Sprintf("%s: locale %q is not present in templatefilters after copy", util.FormatErrorMessage(err), locale),
				)
				return
			}
			tflog.Info(ctx, "locale templates already exist; skipping copy", util.H{"group_id": groupID, "locale": locale})
		}
	}

	state := localeModelFromPlan(plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationsTemplateGroupLocaleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationsTemplateGroupLocaleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := state.GroupID.ValueString()
	locale := state.Locale.ValueString()

	_, err := r.cidaasClient.NotificationsSrvTemplateGroup.Get(ctx, groupID)
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read notifications template group", util.FormatErrorMessage(err))
		return
	}

	present, err := r.localeExistsInGroup(ctx, groupID, locale)
	if err != nil {
		resp.Diagnostics.AddError("failed to read template group templatefilters", util.FormatErrorMessage(err))
		return
	}
	if !present {
		tflog.Info(ctx, "locale not found in templatefilters; removing from state", util.H{"group_id": groupID, "locale": locale})
		resp.State.RemoveResource(ctx)
		return
	}

	out := localeModelFromPlan(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationsTemplateGroupLocaleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationsTemplateGroupLocaleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := localeModelFromPlan(plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationsTemplateGroupLocaleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationsTemplateGroupLocaleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := state.GroupID.ValueString()
	locale := state.Locale.ValueString()

	group, err := r.cidaasClient.NotificationsSrvTemplateGroup.Get(ctx, groupID)
	if err != nil {
		if util.IsResourceNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("failed to read notifications template group", util.FormatErrorMessage(err))
		return
	}
	if strings.EqualFold(strings.TrimSpace(group.DefaultLocale), locale) {
		resp.Diagnostics.AddError(
			"Invalid locale deletion",
			fmt.Sprintf("cannot delete locale %q while it is default_locale on the template group; change default_locale on cidaas_notifications_template_group first", locale),
		)
		return
	}

	if err := r.cidaasClient.NotificationsSrvTemplate.DeleteByGroupAndLocales(ctx, groupID, []string{locale}); err != nil {
		resp.Diagnostics.AddError("failed to delete templates for locale", util.FormatErrorMessage(err))
		return
	}
}

func (r *NotificationsTemplateGroupLocaleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	groupID, locale, ok := parseNotificationsTemplateGroupLocaleImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("expected {group_id}/{locale}, got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), groupID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("locale"), locale)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), groupID+"/"+locale)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("copy_from_group_id"), "default")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("copy_from_locale"), "en")...)
}

func (r *NotificationsTemplateGroupLocaleResource) localeExistsInGroup(ctx context.Context, groupID, locale string) (bool, error) {
	locales, err := r.cidaasClient.NotificationsSrvTemplateGroup.ListTemplateFiltersLocales(ctx, groupID)
	if err != nil {
		return false, err
	}
	for _, l := range locales {
		if l == locale {
			return true, nil
		}
	}
	return false, nil
}

func localeModelFromPlan(m notificationsTemplateGroupLocaleModel) notificationsTemplateGroupLocaleModel {
	return notificationsTemplateGroupLocaleModel{
		ID:              types.StringValue(m.GroupID.ValueString() + "/" + m.Locale.ValueString()),
		GroupID:         m.GroupID,
		Locale:          m.Locale,
		CopyFromGroupID: m.CopyFromGroupID,
		CopyFromLocale:  m.CopyFromLocale,
	}
}

func parseNotificationsTemplateGroupLocaleImportID(id string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(id), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}
