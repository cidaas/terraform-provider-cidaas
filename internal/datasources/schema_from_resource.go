package datasources

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// computedAttributesFromResource converts resource attributes into computed data source attributes, so a data
// source exposes exactly what the matching resource stores. Write-only attributes and their *_wo_version
// markers are dropped because data sources cannot have them.
func computedAttributesFromResource(in map[string]rschema.Attribute) map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(in))
	for name, a := range in {
		if a.IsWriteOnly() || strings.HasSuffix(name, "_wo_version") {
			continue
		}
		out[name] = computedAttributeFromResource(name, a)
	}
	return out
}

// computedAttributeFromResource converts a single resource attribute (recursing into nested attributes).
func computedAttributeFromResource(name string, a rschema.Attribute) schema.Attribute {
	desc := a.GetMarkdownDescription()
	sensitive := a.IsSensitive()
	switch t := a.(type) {
	case rschema.StringAttribute:
		return schema.StringAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}
	case rschema.BoolAttribute:
		return schema.BoolAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}
	case rschema.Int64Attribute:
		return schema.Int64Attribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}
	case rschema.ListAttribute:
		return schema.ListAttribute{Computed: true, Sensitive: sensitive, ElementType: t.ElementType, MarkdownDescription: desc}
	case rschema.MapAttribute:
		return schema.MapAttribute{Computed: true, Sensitive: sensitive, ElementType: t.ElementType, MarkdownDescription: desc}
	case rschema.SingleNestedAttribute:
		return schema.SingleNestedAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc,
			Attributes: computedAttributesFromResource(t.Attributes)}
	case rschema.ListNestedAttribute:
		return schema.ListNestedAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc,
			NestedObject: schema.NestedAttributeObject{Attributes: computedAttributesFromResource(t.NestedObject.Attributes)}}
	case rschema.MapNestedAttribute:
		return schema.MapNestedAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc,
			NestedObject: schema.NestedAttributeObject{Attributes: computedAttributesFromResource(t.NestedObject.Attributes)}}
	}
	panic(fmt.Sprintf("data source schema: unsupported resource attribute type %T for %q", a, name))
}
