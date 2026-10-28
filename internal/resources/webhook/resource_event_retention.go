package webhook

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EventRetentionID is the fixed ID of the tenant-wide cidaas_event_retention singleton (also the import ID).
const EventRetentionID = "event_retention"

var (
	_ resource.Resource                = &eventRetentionResource{}
	_ resource.ResourceWithConfigure   = &eventRetentionResource{}
	_ resource.ResourceWithImportState = &eventRetentionResource{}
	_ resource.ResourceWithModifyPlan  = &eventRetentionResource{}

	eventRetentionCatalogAttrTypes = map[string]attr.Type{
		"object_type":          types.StringType,
		"days_to_live":         types.Int64Type,
		"default_days_to_live": types.Int64Type,
		"min_days_to_live":     types.Int64Type,
		"max_days_to_live":     types.Int64Type,
	}
	eventRetentionCatalogType = types.ObjectType{AttrTypes: eventRetentionCatalogAttrTypes}
)

// eventRetentionResource implements cidaas_event_retention (webhook-srv /eventdescriptions/retention).
type eventRetentionResource struct {
	client *client.Client
}

// EventRetentionState is the Terraform model of cidaas_event_retention.
type EventRetentionState struct {
	ID        types.String `tfsdk:"id"`
	Retention types.Map    `tfsdk:"retention"`
	Catalog   types.Map    `tfsdk:"catalog"`
}

// NewEventRetentionResource returns the cidaas_event_retention resource.
func NewEventRetentionResource() resource.Resource {
	return &eventRetentionResource{}
}

func (r *eventRetentionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_retention"
}

func (r *eventRetentionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.RequireV4("cidaas_event_retention", &resp.Diagnostics) {
		return
	}
	r.client = c
}

func (r *eventRetentionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages per-event-type data retention (`daysToLive` of changelog and user activity events) on cidaas v4 via " +
			"`webhook-srv/eventdescriptions/retention`. One instance per tenant. Separate from `cidaas_webhook`.\n\n" +
			"Requires the event retention feature (`custom_reporting`) to be active on the tenant and the scopes " +
			"`cidaas:eventdescriptions_retention_read` and `cidaas:eventdescriptions_retention_write` " +
			"(`cidaas:webhook_read` is optional and only used to tell unknown event types from immutable ones in errors).\n\n" +
			"Only the event types listed in `retention` are managed. Removing an entry, or destroying the resource, " +
			"restores that event type to its default `daysToLive`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `event_retention` (one instance per tenant).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"retention": schema.MapAttribute{
				ElementType: types.Int64Type,
				Required:    true,
				MarkdownDescription: "Map of event type ID (e.g. `ACCOUNT_MODIFIED`) to `days_to_live`. Each event type must be configurable " +
					"(listed in `catalog`) and each value must be within its `min_days_to_live` / `max_days_to_live`.",
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
					mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"catalog": schema.MapAttribute{
				ElementType: eventRetentionCatalogType,
				Computed:    true,
				MarkdownDescription: "All configurable event types on the tenant, keyed by event type ID, with `object_type`, " +
					"current `days_to_live` and the `default_days_to_live` / `min_days_to_live` / `max_days_to_live` bounds.",
			},
		},
	}
}

// ModifyPlan validates the planned retention against the live catalog so errors surface at plan time.
func (r *eventRetentionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.client == nil {
		return
	}
	var plan EventRetentionState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan.Retention.IsUnknown() || plan.Retention.IsNull() {
		return
	}
	catalog, ok := r.listCatalog(ctx, &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(r.validateRetention(ctx, catalog, knownRetention(plan.Retention))...)
}

func (r *eventRetentionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventRetentionState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	catalog, ok := r.listCatalog(ctx, &resp.Diagnostics)
	if !ok {
		return
	}
	want := knownRetention(plan.Retention)
	resp.Diagnostics.Append(r.validateRetention(ctx, catalog, want)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.patch(ctx, catalog, want, &resp.Diagnostics) {
		return
	}
	next, diags := retentionState(plan.Retention, catalog)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *eventRetentionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventRetentionState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	list, err := r.client.EventRetention.List(ctx)
	if err != nil {
		if errors.Is(err, client.ErrRetentionInactive) {
			resp.Diagnostics.AddWarning("Event retention is not active",
				"webhook-srv reports retention is not active on this tenant; cidaas_event_retention was removed from Terraform state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read event retention", err.Error())
		return
	}
	catalog := catalogByID(list)

	// Import (no retention in state yet) adopts every configurable event type; otherwise keep only managed keys.
	values := map[string]attr.Value{}
	if state.Retention.IsNull() || state.Retention.IsUnknown() {
		for id, ed := range catalog {
			values[id] = types.Int64Value(int64(ed.DaysToLive))
		}
	} else {
		for id := range state.Retention.Elements() {
			if ed, ok := catalog[id]; ok {
				values[id] = types.Int64Value(int64(ed.DaysToLive))
			}
		}
	}
	retention, diags := types.MapValue(types.Int64Type, values)
	resp.Diagnostics.Append(diags...)
	next, diags := retentionState(retention, catalog)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Update PATCHes changed entries and restores defaults for entries removed from config.
func (r *eventRetentionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EventRetentionState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	catalog, ok := r.listCatalog(ctx, &resp.Diagnostics)
	if !ok {
		return
	}
	want := knownRetention(plan.Retention)
	resp.Diagnostics.Append(r.validateRetention(ctx, catalog, want)...)
	if resp.Diagnostics.HasError() {
		return
	}
	changes := map[string]int64{}
	for id, days := range want {
		if ed := catalog[id]; int64(ed.DaysToLive) != days {
			changes[id] = days
		}
	}
	for id, days := range defaultsFor(catalog, removedKeys(state.Retention, want)) {
		changes[id] = days
	}
	if !r.patch(ctx, catalog, changes, &resp.Diagnostics) {
		return
	}
	next, diags := retentionState(plan.Retention, catalog)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

// Delete restores every managed event type to its default daysToLive.
func (r *eventRetentionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventRetentionState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	list, err := r.client.EventRetention.List(ctx)
	if err != nil {
		if errors.Is(err, client.ErrRetentionInactive) {
			return // nothing to restore when the feature is off
		}
		resp.Diagnostics.AddError("Failed to read event retention", err.Error())
		return
	}
	catalog := catalogByID(list)
	restore := map[string]int64{}
	for id, days := range defaultsFor(catalog, removedKeys(state.Retention, nil)) {
		if int64(catalog[id].DaysToLive) != days {
			restore[id] = days
		}
	}
	r.patch(ctx, catalog, restore, &resp.Diagnostics)
}

// ImportState imports the tenant singleton; the only valid import ID is `event_retention`.
func (r *eventRetentionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != EventRetentionID {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("cidaas_event_retention is a tenant singleton; import it with ID %q, got %q.", EventRetentionID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), EventRetentionID)...)
}

func (r *eventRetentionResource) listCatalog(ctx context.Context, diags *diag.Diagnostics) (map[string]client.RetentionEventDescription, bool) {
	list, err := r.client.EventRetention.List(ctx)
	if err != nil {
		if errors.Is(err, client.ErrRetentionInactive) {
			diags.AddError("Event retention is not active",
				"webhook-srv reports retention is not active on this tenant. Enable the event retention feature "+
					"(custom_reporting) before managing cidaas_event_retention. API: "+err.Error())
			return nil, false
		}
		diags.AddError("Failed to read event retention catalog", err.Error())
		return nil, false
	}
	return catalogByID(list), true
}

// patch sends changes (if any) and mirrors the new values into catalog for the state that follows.
func (r *eventRetentionResource) patch(ctx context.Context, catalog map[string]client.RetentionEventDescription, changes map[string]int64, diags *diag.Diagnostics) bool {
	if len(changes) == 0 {
		return true // webhook-srv answers 404 for an empty PATCH
	}
	body := make(map[string]int, len(changes))
	for id, days := range changes {
		body[id] = int(days)
	}
	items, err := r.client.EventRetention.Update(ctx, body)
	if err != nil {
		diags.AddError("Failed to update event retention", err.Error())
		return false
	}
	for _, item := range items {
		if ed, ok := catalog[item.EventID]; ok {
			ed.DaysToLive = item.DaysToLive
			catalog[item.EventID] = ed
		}
	}
	return true
}

// validateRetention mirrors webhook-srv PATCH checks so nothing is half-applied (the API applies entries one by one).
func (r *eventRetentionResource) validateRetention(ctx context.Context, catalog map[string]client.RetentionEventDescription, want map[string]int64) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, id := range sortedKeys(want) {
		days := want[id]
		at := path.Root("retention").AtMapKey(id)
		ed, ok := catalog[id]
		if !ok || ed.Retention == nil {
			diags.AddAttributeError(at, "Event type not configurable", r.notConfigurableReason(ctx, id))
			continue
		}
		if b := ed.Retention; days < int64(b.MinDaysToLive) || days > int64(b.MaxDaysToLive) {
			diags.AddAttributeError(at, "days_to_live out of range",
				fmt.Sprintf("days_to_live for %s must be between %d and %d (default %d), got %d.", id, b.MinDaysToLive, b.MaxDaysToLive, b.DefaultDaysToLive, days))
		}
	}
	return diags
}

func (r *eventRetentionResource) notConfigurableReason(ctx context.Context, id string) string {
	exists, err := r.client.EventRetention.EventTypeExists(ctx, id)
	switch {
	case err != nil:
		return fmt.Sprintf("event type %s is not in the retention catalog (see the catalog attribute).", id)
	case !exists:
		return fmt.Sprintf("unknown event type: %s.", id)
	default:
		return fmt.Sprintf("retention for event type %s cannot be modified (immutable).", id)
	}
}

func retentionState(retention types.Map, catalog map[string]client.RetentionEventDescription) (EventRetentionState, diag.Diagnostics) {
	var diags diag.Diagnostics
	entries := make(map[string]attr.Value, len(catalog))
	for id, ed := range catalog {
		var def, minDays, maxDays int
		if ed.Retention != nil {
			def, minDays, maxDays = ed.Retention.DefaultDaysToLive, ed.Retention.MinDaysToLive, ed.Retention.MaxDaysToLive
		}
		obj, d := types.ObjectValue(eventRetentionCatalogAttrTypes, map[string]attr.Value{
			"object_type":          types.StringValue(ed.ObjectType),
			"days_to_live":         types.Int64Value(int64(ed.DaysToLive)),
			"default_days_to_live": types.Int64Value(int64(def)),
			"min_days_to_live":     types.Int64Value(int64(minDays)),
			"max_days_to_live":     types.Int64Value(int64(maxDays)),
		})
		diags.Append(d...)
		entries[id] = obj
	}
	catalogValue, d := types.MapValue(eventRetentionCatalogType, entries)
	diags.Append(d...)
	return EventRetentionState{
		ID:        types.StringValue(EventRetentionID),
		Retention: retention,
		Catalog:   catalogValue,
	}, diags
}

func catalogByID(list []client.RetentionEventDescription) map[string]client.RetentionEventDescription {
	out := make(map[string]client.RetentionEventDescription, len(list))
	for _, ed := range list {
		out[ed.ID] = ed
	}
	return out
}

// knownRetention returns the known entries of a retention map (unknown values are skipped).
func knownRetention(m types.Map) map[string]int64 {
	out := map[string]int64{}
	if m.IsNull() || m.IsUnknown() {
		return out
	}
	for id, v := range m.Elements() {
		if days, ok := v.(types.Int64); ok && !days.IsNull() && !days.IsUnknown() {
			out[id] = days.ValueInt64()
		}
	}
	return out
}

// removedKeys returns the keys of previous that are not in keep.
func removedKeys(previous types.Map, keep map[string]int64) []string {
	var out []string
	for id := range knownRetention(previous) {
		if _, ok := keep[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}

// defaultsFor maps each id still configurable in catalog to its defaultDaysToLive.
func defaultsFor(catalog map[string]client.RetentionEventDescription, ids []string) map[string]int64 {
	out := map[string]int64{}
	for _, id := range ids {
		if ed, ok := catalog[id]; ok && ed.Retention != nil {
			out[id] = int64(ed.Retention.DefaultDaysToLive)
		}
	}
	return out
}

func sortedKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
