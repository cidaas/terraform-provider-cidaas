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

type ScopesDataSource struct {
	BaseDataSource
}

type ScopesFilterModel struct {
	BaseModel
	Scope []Scope `tfsdk:"scope"`
}

type Scope struct {
	ID                   types.String `tfsdk:"id"`
	SecurityLevel        types.String `tfsdk:"security_level"`
	ScopeKey             types.String `tfsdk:"scope_key"`
	GroupName            types.Set    `tfsdk:"group_name"`
	RequiredUserConsent  types.Bool   `tfsdk:"required_user_consent"`
	LocalizedDescription types.List   `tfsdk:"localized_descriptions"`
	ScopeOwner           types.String `tfsdk:"scope_owner"`
}

type LocalDescription struct {
	Locale      types.String `tfsdk:"locale"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

var scopeFilter = FilterConfig{
	"scope_key":             {TypeFunc: FilterTypeString},
	"security_level":        {TypeFunc: FilterTypeString},
	"group_name":            {TypeFunc: FilterTypeString},
	"required_user_consent": {TypeFunc: FilterTypeBool},
}

var scopeSchema = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the scope.",
	},
	"security_level": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The security level of the scope, `PUBLIC` or `CONFIDENTIAL`.",
	},
	"scope_key": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Unique identifier(name) for the scope.",
	},
	"group_name": schema.SetAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		MarkdownDescription: "List of scope_groups associated with the scope.",
	},
	"required_user_consent": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "Indicates whether user consent is required for the scope.",
	},
	"scope_owner": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The owner of the scope. e.g. `ADMIN`.",
	},
	"localized_descriptions": schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"locale": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "The locale for the scope, e.g., `en-US`.",
				},
				"title": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "The title of the scope in the configured locale.",
				},
				"description": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "The description of the scope in the configured locale.",
				},
			},
		},
	},
}

var scopesDataSourceSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of scopes available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", SCOPE_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": scopeFilter.Schema(),
		"scope": schema.ListNestedBlock{
			Description: "The returned list of scopes.",
			NestedObject: schema.NestedBlockObject{
				Attributes: scopeSchema,
			},
		},
	},
}

func NewScope() datasource.DataSource {
	return &ScopesDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   SCOPE_DATASOURCE,
				Schema: &scopesDataSourceSchema,
			},
		),
	}
}

func (d *ScopesDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data ScopesFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := scopeFilter.GetAndFilter(ctx, d.Client, data.Filters, listScopes)
	if diag != nil {
		tflog.Error(ctx, "failed to filter scopes data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered scopes data")

	data.Scope = parseModel(AnySliceToTyped[cidaas.ScopeModel](result), parseScope)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read scopes data source")
}

func listScopes(ctx context.Context, client *cidaas.Client) ([]any, error) {
	scopes, err := client.Scopes.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(scopes), nil
}

func parseScope(scope cidaas.ScopeModel) Scope {
	r := Scope{
		ID:                  types.StringValue(scope.ID),
		ScopeKey:            types.StringValue(scope.ScopeKey),
		GroupName:           util.SetValueOrNull(scope.GroupName),
		SecurityLevel:       types.StringValue(scope.SecurityLevel),
		RequiredUserConsent: types.BoolValue(scope.RequiredUserConsent),
	}

	var objectValues []attr.Value
	localeDescription := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"locale":      types.StringType,
			"title":       types.StringType,
			"description": types.StringType,
		},
	}

	for _, sc := range scope.LocaleWiseDescription {
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
	r.LocalizedDescription = types.ListValueMust(localeDescription, objectValues)
	return r
}
