package datasources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	LINK    = "LINK"
	CODE    = "CODE"
	IVR     = "IVR"
	EMAIL   = "EMAIL"
	SMS     = "SMS"
	PUSH    = "PUSH"
	GENERAL = "GENERAL"
)

type SystemTemplateOptionsDataSource struct {
	BaseDataSource
}

type SystemTemplateOptionsFilterModel struct {
	BaseModel
	SystemTemplateOption []SystemTemplateOption `tfsdk:"system_template_option"`
}

type SystemTemplateOption struct {
	TemplateKey   types.String `tfsdk:"template_key"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	TemplateTypes types.List   `tfsdk:"template_types"`
}

var systemTemplateFilter = FilterConfig{
	"template_key": {TypeFunc: FilterTypeString},
	"enabled":      {TypeFunc: FilterTypeBool},
	// TODO: add list filter
}

var verificationTypes = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"verification_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The verification type of the template. e.g. `EMAIL`",
			},
			"usage_types": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "The usage type of the template. e.g. `MULTIFACTOR_AUTHENTICATION`",
			},
		},
	},
}

var processingTypes = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"processing_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The processing type of the template. e.g. `LINK` or `CODE` ",
			},
			"supported_tags": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"required": schema.SetAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The required tags in a template. While creating a templates the required tags must be part of the content.",
					},
					"optional": schema.SetAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "This lists provides the optional tags supported in a template content.",
					},
				},
			},
			"verification_types": verificationTypes,
		},
	},
}

var systemTemplateSchema = map[string]schema.Attribute{
	"template_key": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The key of the template.",
	},
	"enabled": schema.BoolAttribute{
		Required:            true,
		MarkdownDescription: "The flag to identify if a system template is enabled.",
	},
	"template_types": schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"template_type": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "The type of the template. e.g. `EMAIL`",
				},
				"processing_types": processingTypes,
			},
		},
	},
}

var systemTemplateDataSourceSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of system templates optionsa that can be"+
		"\nconfigured to create a system template in your Cidaas instance. "+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", SYSTEM_TEMPLATE_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": systemTemplateFilter.Schema(),
		"system_template_option": schema.ListNestedBlock{
			Description: "The returned list of system template options.",
			NestedObject: schema.NestedBlockObject{
				Attributes: systemTemplateSchema,
			},
		},
	},
}

func NewSystemTemplateOption() datasource.DataSource {
	return &SystemTemplateOptionsDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   SYSTEM_TEMPLATE_DATASOURCE,
				Schema: &systemTemplateDataSourceSchema,
			},
		),
	}
}

func (d *SystemTemplateOptionsDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data SystemTemplateOptionsFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := systemTemplateFilter.GetAndFilter(
		ctx,
		d.Client,
		data.Filters,
		listSystemTemplateOptions,
	)
	if diag != nil {
		tflog.Error(ctx, "failed to filter system template options data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered system template options data")
	data.SystemTemplateOption = parseModel(
		AnySliceToTyped[cidaas.MasterList](result),
		parseSystemTemplate,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "successfully read system_template_options data source")
}

func listSystemTemplateOptions(ctx context.Context, client *cidaas.Client) ([]any, error) {
	// system template masterlist is fetched with the groupID "default"
	masterListResp, err := client.Templates.GetMasterList(ctx, "default")
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(masterListResp.Data), nil
}

func parseSystemTemplate(ml cidaas.MasterList) SystemTemplateOption {
	r := SystemTemplateOption{
		TemplateKey: types.StringValue(ml.TemplateKey),
		Enabled:     types.BoolValue(ml.Enabled),
	}

	verificationTypes := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"verification_type": types.StringType,
			"usage_types":       types.SetType{ElemType: types.StringType},
		},
	}

	supportedTagsAttrTypes := map[string]attr.Type{
		"required": types.SetType{ElemType: types.StringType},
		"optional": types.SetType{ElemType: types.StringType},
	}

	processingTypes := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"processing_type": types.StringType,
			"supported_tags": types.ObjectType{
				AttrTypes: supportedTagsAttrTypes,
			},
			"verification_types": types.ListType{
				ElemType: verificationTypes,
			},
		},
	}

	templateTypes := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"template_type": types.StringType,
			"processing_types": types.ListType{
				ElemType: processingTypes,
			},
		},
	}

	var templateTypeobjectValues []attr.Value

	for _, t := range ml.TemplateTypes {
		tt := t.TemplateType
		th := TemaplateTagHandler{
			TemplateData: TemplateData{
				TemplateType: tt,
			},
		}
		th.addSupportedTags()
		var processingTypeobjectValues []attr.Value

		if len(t.ProcessingTypes) > 0 {
			for _, p := range t.ProcessingTypes {
				pt := p.ProcessingType
				th.TemplateData.ProcessingType = pt

				var verificationTypeobjectValues []attr.Value

				for _, v := range p.VerificationTypes {
					vt := v.VerificationType
					vtObjValue := types.ObjectValueMust(
						verificationTypes.AttrTypes,
						map[string]attr.Value{
							"verification_type": util.StringValueOrNull(&vt),
							"usage_types":       util.SetValueOrNull(v.UsageTypes),
						})
					verificationTypeobjectValues = append(verificationTypeobjectValues, vtObjValue)
				}

				ptObjValue := types.ObjectValueMust(processingTypes.AttrTypes, map[string]attr.Value{
					"processing_type": util.StringValueOrNull(&pt),
					"supported_tags": types.ObjectValueMust(
						supportedTagsAttrTypes,
						map[string]attr.Value{
							"required": util.SetValueOrNull(th.getTemplateTags(r.TemplateKey.ValueString()).Required),
							"optional": util.SetValueOrNull(th.getTemplateTags(r.TemplateKey.ValueString()).Optional),
						}),
					"verification_types": types.ListValueMust(verificationTypes, verificationTypeobjectValues),
				})
				processingTypeobjectValues = append(processingTypeobjectValues, ptObjValue)
			}
		} else {
			pt := GENERAL
			th.TemplateData.ProcessingType = pt
			var verificationTypeobjectValues []attr.Value

			ptObjValue := types.ObjectValueMust(processingTypes.AttrTypes, map[string]attr.Value{
				"processing_type": util.StringValueOrNull(&pt),
				"supported_tags": types.ObjectValueMust(
					supportedTagsAttrTypes,
					map[string]attr.Value{
						"required": util.SetValueOrNull(th.getTemplateTags(r.TemplateKey.ValueString()).Required),
						"optional": util.SetValueOrNull(th.getTemplateTags(r.TemplateKey.ValueString()).Optional),
					}),
				"verification_types": types.ListValueMust(verificationTypes, verificationTypeobjectValues),
			})
			processingTypeobjectValues = append(processingTypeobjectValues, ptObjValue)
		}

		ttObjValue := types.ObjectValueMust(templateTypes.AttrTypes, map[string]attr.Value{
			"template_type":    util.StringValueOrNull(&tt),
			"processing_types": types.ListValueMust(processingTypes, processingTypeobjectValues),
		})
		templateTypeobjectValues = append(templateTypeobjectValues, ttObjValue)
	}
	r.TemplateTypes = types.ListValueMust(templateTypes, templateTypeobjectValues)
	return r
}

type TemplateTags struct {
	Required []string
	Optional []string
}

type TemplateData struct {
	TemplateType   string
	ProcessingType string
}

type TagsListPayload struct {
	VerifyAccountEmailSMSLink   TemplateTags
	VerifyAccountEmailSMSCode   TemplateTags
	VerifyAccountIVRCode        TemplateTags
	WelcomeUserEmailSMSLink     TemplateTags
	WelcomeUserIVRLink          TemplateTags
	InviteUserEmail             TemplateTags
	ResetPasswordEmailLink      TemplateTags
	ResetPasswordEmailCode      TemplateTags
	ResetPasswordSMS            TemplateTags
	ResetPasswordIVR            TemplateTags
	AfterChangePasswordEmailSMS TemplateTags
	AfterChangePasswordIVR      TemplateTags
	UserCreatedEmailSMS         TemplateTags
	VerifyUserEmailLink         TemplateTags
	VerifyUserCodeGeneral       TemplateTags
	VerifyUserSMSIVR            TemplateTags
	VerifyUserPush              TemplateTags
	NotifyCommunicationChange   TemplateTags
}

type TemaplateTagHandler struct {
	TemplateData    TemplateData
	TagsListPayload TagsListPayload
}

func (th *TemaplateTagHandler) getTemplateTags(text string) TemplateTags {
	var tags TemplateTags

	switch text {
	case "VERIFY_ACCOUNT":
		switch th.TemplateData.TemplateType {
		case EMAIL, SMS:
			switch th.TemplateData.ProcessingType {
			case LINK:
				tags = th.TagsListPayload.VerifyAccountEmailSMSLink
			case CODE:
				tags = th.TagsListPayload.VerifyAccountEmailSMSCode
			}
		case IVR:
			tags = th.TagsListPayload.VerifyAccountIVRCode
		}
	case "WELCOME_USER":
		switch th.TemplateData.TemplateType {
		case EMAIL, SMS:
			tags = th.TagsListPayload.WelcomeUserEmailSMSLink
		case IVR:
			tags = th.TagsListPayload.WelcomeUserIVRLink
		}
	case "INVITE_USER":
		tags = th.TagsListPayload.InviteUserEmail
	case "RESET_PASSWORD":
		switch th.TemplateData.TemplateType {
		case EMAIL:
			switch th.TemplateData.ProcessingType {
			case LINK:
				tags = th.TagsListPayload.ResetPasswordEmailLink
			case CODE:
				tags = th.TagsListPayload.ResetPasswordEmailCode
			}
		case SMS:
			tags = th.TagsListPayload.ResetPasswordSMS
		case IVR:
			tags = th.TagsListPayload.ResetPasswordIVR
		}
	case "CHANGE_PASSWORD", "AFTER_CHANGE_PASSWORD":
		switch th.TemplateData.TemplateType {
		case EMAIL, SMS:
			tags = th.TagsListPayload.AfterChangePasswordEmailSMS
		case IVR:
			tags = th.TagsListPayload.AfterChangePasswordIVR
		}
	case "NEW_DEVICE", "NEW_LOCATION", "NEW_NETWORK":
		tags = TemplateTags{
			Optional: []string{"{{name}}", "{{account_name}}}"},
		}
	case "USER_CREATED":
		tags = th.TagsListPayload.UserCreatedEmailSMS
	case "VERIFY_USER":
		switch th.TemplateData.TemplateType {
		case EMAIL:
			switch th.TemplateData.ProcessingType {
			case LINK:
				tags = th.TagsListPayload.VerifyUserEmailLink
			case CODE, GENERAL:
				tags = th.TagsListPayload.VerifyUserCodeGeneral
			}
		case SMS, IVR:
			tags = th.TagsListPayload.VerifyUserSMSIVR
		case PUSH:
			tags = th.TagsListPayload.VerifyUserPush
		}
	case "NOTIFY_COMMUNICATION_CHANGE":
		tags = th.TagsListPayload.NotifyCommunicationChange
	}
	return tags
}

func (th *TemaplateTagHandler) addSupportedTags() {
	th.TagsListPayload = TagsListPayload{
		VerifyAccountEmailSMSLink: TemplateTags{
			Required: []string{"{{{verify_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		VerifyAccountEmailSMSCode: TemplateTags{
			Required: []string{"{{code}}", "{{{verify_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		VerifyAccountIVRCode: TemplateTags{
			Required: []string{"{{code}}"},
		},
		WelcomeUserEmailSMSLink: TemplateTags{
			Required: []string{"{{{login_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		WelcomeUserIVRLink: TemplateTags{
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		InviteUserEmail: TemplateTags{
			Required: []string{"{{{invite_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}", "{{{invited_by}}}"},
		},
		ResetPasswordEmailLink: TemplateTags{
			Required: []string{"{{{reset_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		ResetPasswordEmailCode: TemplateTags{
			Required: []string{"{{code}}, {{{reset_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		ResetPasswordSMS: TemplateTags{
			Required: []string{"{{code}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		ResetPasswordIVR: TemplateTags{
			Required: []string{"{{code}}"},
		},
		AfterChangePasswordEmailSMS: TemplateTags{
			Required: []string{"{{{reset_link}}}"},
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		AfterChangePasswordIVR: TemplateTags{
			Optional: []string{"{{name}}", "{{account_name}}}"},
		},
		UserCreatedEmailSMS: TemplateTags{
			Optional: []string{"{{name}}", "{{account_name}}}", "{{user_name}}", "{{password}}", "{{{login_link}}}"},
		},
		VerifyUserEmailLink: TemplateTags{
			Optional: []string{"{{name}}", "{{code}}}"},
			Required: []string{"{{verify_link}}"},
		},
		VerifyUserCodeGeneral: TemplateTags{
			Optional: []string{"{{name}}", "{{verify_link}}}"},
			Required: []string{"{{code}}"},
		},
		VerifyUserSMSIVR: TemplateTags{
			Optional: []string{"{{name}}"},
			Required: []string{"{{code}}"},
		},
		VerifyUserPush: TemplateTags{
			Required: []string{"{{address}}"},
		},
		NotifyCommunicationChange: TemplateTags{
			Optional: []string{"{{name}}", "{{account_name}}}", "{{communication_medium_value}}"},
		},
	}
}
