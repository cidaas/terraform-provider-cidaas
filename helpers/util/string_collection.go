package util //nolint:revive

import "github.com/hashicorp/terraform-plugin-framework/types"

// ListToStrings converts a Terraform string list to []string.
// Null/unknown lists return nil. Null or unknown elements are skipped so
// resource attribute references unresolved at plan time do not panic with
// "Received unknown value... Target Type: string" (ElementsAs into []string).
func ListToStrings(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	elems := l.Elements()
	out := make([]string, 0, len(elems))
	for _, e := range elems {
		if e.IsUnknown() || e.IsNull() {
			continue
		}
		sv, ok := e.(types.String)
		if !ok {
			continue
		}
		out = append(out, sv.ValueString())
	}
	return out
}

// SetToStrings converts a Terraform string set to []string.
// Null/unknown sets return nil. Null or unknown elements are skipped for the
// same reason as ListToStrings.
func SetToStrings(s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	elems := s.Elements()
	out := make([]string, 0, len(elems))
	for _, e := range elems {
		if e.IsUnknown() || e.IsNull() {
			continue
		}
		sv, ok := e.(types.String)
		if !ok {
			continue
		}
		out = append(out, sv.ValueString())
	}
	return out
}

// ListContainsUnknown reports whether any element in a known, non-null list is unknown.
func ListContainsUnknown(l types.List) bool {
	if l.IsNull() || l.IsUnknown() {
		return false
	}
	for _, e := range l.Elements() {
		if e.IsUnknown() {
			return true
		}
	}
	return false
}

// SetContainsUnknown reports whether any element in a known, non-null set is unknown.
func SetContainsUnknown(s types.Set) bool {
	if s.IsNull() || s.IsUnknown() {
		return false
	}
	for _, e := range s.Elements() {
		if e.IsUnknown() {
			return true
		}
	}
	return false
}
