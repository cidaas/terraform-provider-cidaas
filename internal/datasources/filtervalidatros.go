package datasources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type filterNameValidator struct {
	FilterConfig FilterConfig
}

func (v filterNameValidator) Description(_ context.Context) string {
	return "validate that the provided field is filterable"
}

func (v filterNameValidator) MarkdownDescription(_ context.Context) string {
	return "validate that the provided field is filterable"
}

func (v filterNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	fieldName := req.ConfigValue.ValueString()
	_, ok := v.FilterConfig[fieldName]

	if !ok {
		// Aggregate filterable attributes
		var filterableAttributes []string

		for k := range v.FilterConfig {
			filterableAttributes = append(filterableAttributes, k)
		}

		sort.Strings(filterableAttributes)

		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Non-Filterable Field",
			fmt.Sprintf(
				"Field %q is not filterable.\nFilterable Fields: %s",
				fieldName,
				strings.Join(filterableAttributes, ", "),
			),
		)
		return
	}
}
