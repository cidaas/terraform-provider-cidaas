// Package notification manages cidaas notification templates and service setup resources.
package notification

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	allowedTemplateTypes = []string{"EMAIL", "SMS", "IVR", "PUSH"}
	cidaasClient         *cidaas.Client
)

type TemplateConfig struct {
	ID               types.String `tfsdk:"id"`
	Locale           types.String `tfsdk:"locale"`
	TemplateKey      types.String `tfsdk:"template_key"`
	TemplateType     types.String `tfsdk:"template_type"`
	Content          types.String `tfsdk:"content"`
	Subject          types.String `tfsdk:"subject"`
	TemplateOwner    types.String `tfsdk:"template_owner"`
	UsageType        types.String `tfsdk:"usage_type"`
	ProcessingType   types.String `tfsdk:"processing_type"`
	VerificationType types.String `tfsdk:"verification_type"`
	Language         types.String `tfsdk:"language"`
	GroupID          types.String `tfsdk:"group_id"`
	IsSystemTemplate types.Bool   `tfsdk:"is_system_template"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

type TemplateResource struct {
	base.BaseResource
}

func NewTemplateResource() resource.Resource {
	return &TemplateResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_TEMPLATE,
				Schema: &templateSchema,
			},
		),
	}
}

var templateSchema = schema.Schema{
	MarkdownDescription: "**Deprecated for new designs:** this resource uses legacy **templates-srv**. " +
		"For notification-srv (`/{notifications_context_path}/templates`), use **`cidaas_notification_template`** instead.\n\n" +
		"The Template resource in the provider is used to define and manage templates within the Cidaas system." +
		" Templates are used for emails, SMS, IVR, and push notifications." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:templates_read" +
		"\n- cidaas:templates_write" +
		"\n- cidaas:templates_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the template resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"locale": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The BCP47 locale of the template (e.g. `en-US`, `en-GB`, `de-DE`, or a language-only tag like `en`). Use lowercase for the language subtag and uppercase for the region when present. Must be one of the allowed locale tags; see the Allowed Locales section below. It cannot be updated for an existing state.",
			Validators: []validator.String{
				stringvalidator.OneOf(util.AllowedBCP47Locales...),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"template_key": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The unique name of the template. It cannot be updated for an existing state.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[A-Z0-9_-]+$`),
					`must be a valid string consisting only of uppercase letters, digits (0-9), underscores (_), and hyphens (-). Example: SAMPLE, 12345, SAMPLE-TEMPLATE, SAMPLE_TEMPLATE, SAMPLE12345, SAMPLE-1234`,
				),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"template_type": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The type of the template. Allowed template_types are EMAIL, SMS, IVR and PUSH. Template types are case sensitive. It cannot be updated for an existing state.",
			Validators: []validator.String{
				stringvalidator.OneOf(allowedTemplateTypes...),
			},
			PlanModifiers: []planmodifier.String{
				&validators.UniqueIdentifier{},
			},
		},
		"content": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The content of the template.",
		},
		"subject": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Applicable only for template_type EMAIL. It represents the subject of an email.",
		},
		"template_owner": schema.StringAttribute{
			Computed: true,
			// Do not use UseStateForUnknown: API may return a different owner after update
			// (e.g. client vs admin) and the plan must match the Upsert response.
			MarkdownDescription: "The template owner of the template.",
		},
		"usage_type": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The usage_type attribute specifies the specific use case or application for the template. Only applicable for SYSTEM templates." +
				" It should be set to `GENERAL` when cidaas does not provide an allowed list of values.",
			Computed: true,
		},
		"processing_type": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The processing_type attribute specifies the method by which the template information is processed and delivered. Only applicable for SYSTEM templates." +
				" It should be set to `GENERAL` when cidaas does not provide an allowed list of values.",
		},
		"verification_type": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The verification_type attribute defines the method used for verification. Only applicable for SYSTEM templates.",
		},
		"language": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The language based on the local provided in the configuration.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The `group_id` under which the configured template will be categorized. Only applicable for SYSTEM templates.",
		},
		"is_system_template": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "A boolean flag to decide between SYSTEM and CUSTOM template. When set to true the provider creates a SYSTEM template else CUSTOM",
			Default:             booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				&systemTemplateValidator{},
			},
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "A boolean flag to enable or disable the template.",
			Default:             booldefault.StaticBool(true),
			PlanModifiers: []planmodifier.Bool{
				&systemTemplateValidator{},
			},
		},
	},
}

func (r *TemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config TemplateConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if config.TemplateType.ValueString() == "EMAIL" && config.Subject.IsNull() {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			"The attribute subject can not be empty when template_type is EMAIL",
		)
		return
	}
	if !config.Subject.IsNull() && config.TemplateType.ValueString() != "EMAIL" {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			"The attribute subject is only allowed when template_type is EMAIL",
		)
		return
	}
	if !config.IsSystemTemplate.ValueBool() &&
		(!config.GroupID.IsNull() || !config.ProcessingType.IsNull() || !config.VerificationType.IsNull() || !config.UsageType.IsNull()) {
		message := "Attributes group_id, processing_type, verification_type and usage_type are not allowed in when is_system_template is set to false." +
			"\nTo create a system template, set is_system_template to true. Otherwise, remove these attributes from the configuration for a custom template."
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration",
			message,
		)
		return
	}
}

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config TemplateConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	cidaasClient = r.CidaasClient // global variable cidaasClient is assigned here
	resp.Diagnostics.Append(validateSystemTemplateConfig(ctx, config)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare template or validate configuration", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	template := prepareTemplateModel(plan)
	res, err := r.CidaasClient.Templates.Upsert(ctx, *template, plan.IsSystemTemplate.ValueBool())
	if err != nil {
		tflog.Error(ctx, "failed to create template via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create template", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created template via API", util.H{
		"template_id": res.Data.ID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	plan.TemplateOwner = util.StringValueOrNull(&res.Data.TemplateOwner)
	plan.Language = util.StringValueOrNull(&res.Data.Language)
	plan.UsageType = util.StringValueOrNull(&res.Data.UsageType)
	plan.GroupID = util.StringValueOrNull(&res.Data.GroupID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource template created successfully", util.H{
		"template_id": res.Data.ID,
	})
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	cidaasClient = r.CidaasClient // global variable cidaasClient is assigned here
	resp.Diagnostics.Append(validateSystemTemplateConfig(ctx, state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data or validate template configuration", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	template := cidaas.TemplateModel{
		Locale:       state.Locale.ValueString(),
		TemplateKey:  state.TemplateKey.ValueString(),
		TemplateType: state.TemplateType.ValueString(),
	}

	if state.IsSystemTemplate.ValueBool() {
		template.ProcessingType = state.ProcessingType.ValueString()
		template.UsageType = state.UsageType.ValueString()
		template.VerificationType = state.VerificationType.ValueString()
		template.GroupID = state.GroupID.ValueString()
	}

	res, err := r.CidaasClient.Templates.Get(ctx, template, state.IsSystemTemplate.ValueBool())
	if err != nil {
		if base.ReadHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read template via API", util.H{
			"template_key": state.TemplateKey.ValueString(),
			"error":        err.Error(),
		})
		resp.Diagnostics.AddError("failed to read template", util.FormatErrorMessage(err))
		return
	}

	state.ID = util.StringValueOrNull(&res.Data.ID)
	state.TemplateOwner = util.StringValueOrNull(&res.Data.TemplateOwner)
	state.UsageType = util.StringValueOrNull(&res.Data.UsageType)
	state.Language = util.StringValueOrNull(&res.Data.Language)
	state.GroupID = util.StringValueOrNull(&res.Data.GroupID)
	state.Content = util.StringValueOrNull(&res.Data.Content)
	state.Enabled = util.BoolValueOrNull(&res.Data.Enabled)

	if state.TemplateOwner.ValueString() == "DEVELOPER" {
		state.IsSystemTemplate = types.BoolValue(false)
	}

	if state.IsSystemTemplate.ValueBool() {
		template.ProcessingType = state.ProcessingType.ValueString()
		template.VerificationType = state.VerificationType.ValueString()
	}

	state.Subject = util.StringValueOrNull(&res.Data.Subject)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource template read successfully", util.H{
		"template_id": res.Data.ID,
	})
}

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl
	var plan, state TemplateConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	cidaasClient = r.CidaasClient // global variable cidaasClient is assigned here
	resp.Diagnostics.Append(validateSystemTemplateConfig(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state data or validate template configuration", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	template := prepareTemplateModel(plan)
	template.ID = state.ID.ValueString()
	res, err := r.CidaasClient.Templates.Upsert(ctx, *template, plan.IsSystemTemplate.ValueBool())
	if err != nil {
		tflog.Error(ctx, "failed to update template via API", util.H{
			"template_id": state.ID.ValueString(),
			"error":       err.Error(),
		})
		resp.Diagnostics.AddError("failed to update template", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated template via API", util.H{
		"template_id": state.ID.ValueString(),
	})

	plan.TemplateOwner = util.StringValueOrNull(&res.Data.TemplateOwner)
	plan.UsageType = util.StringValueOrNull(&res.Data.UsageType)
	plan.Language = util.StringValueOrNull(&res.Data.Language)
	plan.GroupID = util.StringValueOrNull(&res.Data.GroupID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource template updated successfully", util.H{
		"template_id": state.ID.ValueString(),
	})
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	if state.IsSystemTemplate.ValueBool() {
		resp.Diagnostics.AddWarning(
			"The cidaas_template state has been destroyed. However, deleting system template for a specific template_key is not supported in cidaas system.",
			"Alternatively, you can delete the template_group, but please note that this will remove all system templates within that group.",
		)
	} else {
		err := r.CidaasClient.Templates.Delete(ctx, state.TemplateKey.ValueString(), state.TemplateType.ValueString())
		if err != nil {
			tflog.Error(ctx, "Failed to delete template via API", util.H{
				"template_key":  state.TemplateKey.ValueString(),
				"template_type": state.TemplateType.ValueString(),
				"error":         err.Error(),
			})
			resp.Diagnostics.AddError("failed to delete template", util.FormatErrorMessage(err))
			return
		}
		tflog.Info(ctx, "resource template deleted successfully", util.H{
			"template_key":  state.TemplateKey.ValueString(),
			"template_type": state.TemplateType.ValueString(),
		})
	}
}

func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'template_key:template_type:local', got: %s", id),
		)
		return
	}

	templateKey := parts[0]
	templateType := parts[1]
	locale := parts[2]

	if templateKey != strings.ToUpper(templateKey) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected template_key to be in uppercase. Got: %s", templateKey),
		)
		return
	}

	if !util.Contains(allowedTemplateTypes, templateType) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Invalid template_type provided in import identifier. Valid template_types %+v, got: %s", allowedTemplateTypes, templateType),
		)
		return
	}

	if !util.Contains(util.AllowedBCP47Locales, locale) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Invalid locale provided in import identifier. Valid locales %+v, got: %s", util.AllowedBCP47Locales, locale),
		)
		return
	}

	message := "System Templates cannot be imported using Terraform's import functionality." +
		"To import an existing system template, please create a system template configuration and run the \033[1mterraform apply\033[0m command."
	resp.Diagnostics.AddWarning("System Template Import Alert", message)
	resp.State.SetAttribute(ctx, path.Root("template_key"), templateKey)
	resp.State.SetAttribute(ctx, path.Root("template_type"), templateType)
	resp.State.SetAttribute(ctx, path.Root("locale"), locale)
}

func prepareTemplateModel(plan TemplateConfig) *cidaas.TemplateModel {
	var template cidaas.TemplateModel

	template.Locale = plan.Locale.ValueString()
	template.TemplateKey = plan.TemplateKey.ValueString()
	template.TemplateType = plan.TemplateType.ValueString()
	template.Content = plan.Content.ValueString()
	template.Subject = plan.Subject.ValueString()
	template.UsageType = plan.UsageType.ValueString()
	template.GroupID = plan.GroupID.ValueString()
	template.Enabled = plan.Enabled.ValueBool()

	template.ProcessingType = plan.ProcessingType.ValueString()
	template.VerificationType = plan.VerificationType.ValueString()

	return &template
}

var _ planmodifier.Bool = systemTemplateValidator{}

type systemTemplateValidator struct{}

func (v systemTemplateValidator) Description(_ context.Context) string {
	return "Validates system template configuration"
}

func (v systemTemplateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v systemTemplateValidator) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) { //nolint:gocognit
	var config, state TemplateConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !req.StateValue.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if config.IsSystemTemplate.ValueBool() {
		if !req.StateValue.IsNull() {
			if !state.ProcessingType.Equal(config.ProcessingType) {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attribute processing_type can't be modified. Existing value %s, got %s", state.ProcessingType.ValueString(), config.ProcessingType.ValueString()))
			}
			if !state.GroupID.Equal(config.GroupID) {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attribute group_id can't be modified. Existing value %s, got %s", state.GroupID.ValueString(), config.GroupID.ValueString()))
			}
			if !state.VerificationType.Equal(config.VerificationType) {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attribute verification_type can't be modified. Existing value %s, got %s", state.VerificationType.ValueString(), config.VerificationType.ValueString()))
			}
			if !state.UsageType.Equal(config.UsageType) {
				resp.Diagnostics.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attribute usage_type can't be modified. Existing value %s, got %s", state.UsageType.ValueString(), config.UsageType.ValueString()))
			}
		} else if config.GroupID.IsUnknown() {
			resp.Diagnostics.AddError(
				"Unexpected Resource Configuration",
				"Attribute group_id can't be empty when creating a system template.",
			)
		}
	}
}

func validateSystemTemplateConfig(ctx context.Context, config TemplateConfig) diag.Diagnostics { //nolint:gocognit
	var diags diag.Diagnostics
	if config.IsSystemTemplate.ValueBool() {
		masterList, err := cidaasClient.Templates.GetMasterList(ctx, config.GroupID.ValueString())
		if err != nil {
			diags.AddError(
				fmt.Sprintf("Failed to read the settings list for the provided group_id %s. Please check whether the provided group_id is valid.", config.GroupID.ValueString()),
				err.Error(),
			)
			return diags
		}
		templateKeys := make([]string, len(masterList.Data))
		masterListMap := map[string]cidaas.MasterList{}
		for i, v := range masterList.Data {
			templateKeys[i] = v.TemplateKey
			masterListMap[v.TemplateKey] = v
		}

		if !util.Contains(templateKeys, config.TemplateKey.ValueString()) {
			diags.AddError(
				"Unexpected Resource Configuration",
				fmt.Sprintf("Invalid template_key for system template. The template_key must be one of %+v, got: %s", templateKeys, config.TemplateKey.ValueString()),
			)
			return diags
		}
		allowedTemplateTypes := make([]string, len(masterListMap[config.TemplateKey.ValueString()].TemplateTypes))
		processingTypesByTemplateType := map[string][]string{}
		processingTypes := map[string]cidaas.ProcessingType{}

		defaultProcessingType := "GENERAL"

		// must be enabled as well
		for i, v := range masterListMap[config.TemplateKey.ValueString()].TemplateTypes {
			allowedTemplateTypes[i] = v.TemplateType
			var p []string
			for _, value := range v.ProcessingTypes {
				p = append(p, value.ProcessingType)
				key := v.TemplateType + "-" + value.ProcessingType
				processingTypes[key] = value
			}
			if v.TemplateType == config.TemplateType.ValueString() {
				if v.Default.ProcessingType != "" {
					defaultProcessingType = v.Default.ProcessingType
				}
			}
			processingTypesByTemplateType[v.TemplateType] = p
		}
		if !util.Contains(allowedTemplateTypes, config.TemplateType.ValueString()) {
			diags.AddError(
				"Unexpected Resource Configuration",
				fmt.Sprintf("Invalid template_type for system template %s. Allowed template types %+v, got: %s", config.TemplateKey.ValueString(), allowedTemplateTypes, config.TemplateType.ValueString()),
			)
			return diags
		}

		if len(processingTypesByTemplateType[config.TemplateType.ValueString()]) > 0 {
			if config.ProcessingType.IsNull() {
				diags.AddError(
					"Unexpected Resource Configuration",
					fmt.Sprintf("Attribute processing_type is required for system template with template_key %s and template_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString()),
				)
				return diags
			}
			if !util.Contains(processingTypesByTemplateType[config.TemplateType.ValueString()], config.ProcessingType.ValueString()) {
				message := fmt.Sprintf("Invalid processing_type for system template with template_key %s and template_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString()) +
					fmt.Sprintf("Allowed processing types %+v, got: %s", processingTypesByTemplateType[config.TemplateType.ValueString()], config.ProcessingType.String())
				diags.AddError(
					"Unexpected Resource Configuration",
					message,
				)
				return diags
			}
		} else if config.ProcessingType.IsNull() || (!config.ProcessingType.IsNull() && config.ProcessingType.ValueString() != defaultProcessingType) {
			message := fmt.Sprintf("The attribute \033[1mprocessing_type\033[0m must be set to \033[1m%s\033[0m for the provided configuration.", defaultProcessingType)
			diags.AddError(
				"Unexpected Resource Configuration",
				message,
			)
			return diags
		}

		var allowedUsageTypes []string
		processingTypeKey := config.TemplateType.ValueString() + "-" + config.ProcessingType.ValueString()
		if len(processingTypes[processingTypeKey].VerificationTypes) > 0 {
			var allowedVerificationTypes []string
			for _, v := range processingTypes[processingTypeKey].VerificationTypes {
				allowedVerificationTypes = append(allowedVerificationTypes, v.VerificationType)
				allowedUsageTypes = v.UsageTypes
			}
			if config.VerificationType.IsNull() {
				message := fmt.Sprintf("Attribute verification_type is required for system template with template_key %s, template_type %s and processing_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString(), config.ProcessingType.ValueString()) +
					fmt.Sprintf("Allowed verification_types %+v", allowedVerificationTypes)
				diags.AddError(
					"Unexpected Resource Configuration",
					message,
				)
				return diags
			}

			if !util.Contains(allowedVerificationTypes, config.VerificationType.ValueString()) {
				message := fmt.Sprintf("Invalid verification_type for system template with template_key %s, template_type %s and processing_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString(), config.ProcessingType.ValueString()) +
					fmt.Sprintf("Allowed verification_types %+v, got: %s", allowedVerificationTypes, config.VerificationType.String())
				diags.AddError(
					"Unexpected Resource Configuration",
					message,
				)
				return diags
			}
		} else if !config.VerificationType.IsNull() {
			message := fmt.Sprintf("Unsupported attribute verification_type for the system template with template_key %s, template_type %s and processing_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString(), config.ProcessingType.ValueString()) +
				"Please try another combination of template_key, template_type, processing_type and verification_type or remove this attribute from the configuration."
			diags.AddError(
				"Unexpected Resource Configuration",
				message,
			)
			return diags
		}

		if len(allowedUsageTypes) > 0 {
			if config.UsageType.IsNull() {
				message := fmt.Sprintf("Attribute usage_type is required for system template with template_key %s, template_type %s, processing_type %s and verification_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString(), config.ProcessingType.ValueString(), config.VerificationType.ValueString()) +
					fmt.Sprintf("Allowed usage_types %+v", allowedUsageTypes)
				diags.AddError(
					"Unexpected Resource Configuration",
					message,
				)
				return diags
			}

			if !util.Contains(allowedUsageTypes, config.UsageType.ValueString()) {
				message := fmt.Sprintf("Invalid usage_type for system template with template_key %s, template_type %s, processing_type %s and verification_type %s.", config.TemplateKey.ValueString(), config.TemplateType.ValueString(), config.ProcessingType.ValueString(), config.VerificationType.ValueString()) +
					fmt.Sprintf("Allowed usage_types %+v, got: %s", allowedUsageTypes, config.UsageType.String())
				diags.AddError(
					"Unexpected Resource Configuration",
					message,
				)
				return diags
			}
		}
	}
	return diags
}
