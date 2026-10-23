package datasources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/verification"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = (*VerificationMethodDataSource)(nil)

// VerificationMethodDataSource is the cidaas_verification_method data source.
type VerificationMethodDataSource struct {
	client *client.Client
}

var verificationMethodDataSourceSchema = buildVerificationMethodDataSourceSchema()

// buildVerificationMethodDataSourceSchema derives the data source schema from the resource schema;
// verification_type is the lookup key.
func buildVerificationMethodDataSourceSchema() schema.Schema {
	attrs := computedAttributesFromResource(verification.MethodResourceSchema().Attributes)
	attrs["verification_type"] = schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Verification method type to look up, e.g. `EMAIL`, `TOTP` or the name of a pluggable method.",
	}
	return schema.Schema{
		MarkdownDescription: fmt.Sprintf("The data source `%s` reads one MFA verification method (system or pluggable) from "+
			"`verification-actions-srv/config` on cidaas v4 and exposes the same attributes as the `cidaas_verification_method` resource.",
			VERIFICATION_METHOD_DATASOURCE),
		Attributes: attrs,
	}
}

// NewVerificationMethod returns the cidaas_verification_method data source.
func NewVerificationMethod() datasource.DataSource {
	return &VerificationMethodDataSource{}
}

// Metadata sets the data source type name.
func (d *VerificationMethodDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = VERIFICATION_METHOD_DATASOURCE
}

// Schema returns the data source schema.
func (d *VerificationMethodDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = verificationMethodDataSourceSchema
}

// Configure blocks the data source unless the provider targets cidaas 4.x.
func (d *VerificationMethodDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	if !c.RequireV4(VERIFICATION_METHOD_DATASOURCE, &resp.Diagnostics) {
		return
	}
	d.client = c
}

// Read looks up the method by verification_type; a missing type lists the available ones.
func (d *VerificationMethodDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var verificationType string
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("verification_type"), &verificationType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	methods, err := d.client.VerificationMethod.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read verification methods", err.Error())
		return
	}
	var found *client.VerificationMethodModel
	available := make([]string, 0, len(methods))
	for i := range methods {
		available = append(available, methods[i].VerificationType)
		if methods[i].VerificationType == verificationType {
			found = &methods[i]
		}
	}
	if found == nil {
		sort.Strings(available)
		resp.Diagnostics.AddError("Verification method not found",
			fmt.Sprintf("No verification method %q; available verification types: %s", verificationType, strings.Join(available, ", ")))
		return
	}

	state, diags := verification.FlattenVerificationMethod(ctx, found, nil)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
