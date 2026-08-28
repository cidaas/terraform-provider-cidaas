package datasources

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	EXACT     = "exact"
	SUBSTRING = "substring"
	REGEX     = "regex"
)

// ListFunc is a wrapper for functions that will list and return values from the API.
type ListFunc func(ctx context.Context, client *cidaas.Client) ([]any, error)

type FilterModel struct {
	Name    types.String   `tfsdk:"name" json:"name"`
	Values  []types.String `tfsdk:"values" json:"values"`
	MatchBy types.String   `tfsdk:"match_by" json:"match_by"`
}

type FiltersModelType []FilterModel

// FilterTypeFunc is a function that takes in a filter name and value,
// and returns the value converted to the correct filter type
type FilterTypeFunc func(value string) (any, error)

// FilterAttribute is used to configure filtering for an individual response field
type FilterAttribute struct {
	TypeFunc FilterTypeFunc
}

// Config is the root configuration type for filter data sources
type FilterConfig map[string]FilterAttribute

// schema for the `filter` attributes
func (f FilterConfig) Schema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						f.validateFilterable(),
					},
					Description: "The name of the attribute to filter on.",
				},
				"values": schema.SetAttribute{
					Required:    true,
					Description: "The value(s) to be used in the filter.",
					ElementType: types.StringType,
				},
				"match_by": schema.StringAttribute{
					Optional:    true,
					Description: "The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`",
					Validators: []validator.String{
						stringvalidator.OneOfCaseInsensitive(
							"exact", "substring", "regex",
						),
					},
				},
			},
		},
	}
}

// GetAndFilter will run all filter operations given the parameters
func (f FilterConfig) GetAndFilter(ctx context.Context, client *cidaas.Client, filters []FilterModel, listFunc ListFunc) ([]any, diag.Diagnostic) {
	elems, err := listFunc(ctx, client)
	if err != nil {
		return nil, diag.NewErrorDiagnostic(
			"Failed to list resources",
			err.Error(),
		)
	}

	filteredElements, d := f.applyFiltering(filters, elems)
	if d != nil {
		return nil, d
	}
	return filteredElements, nil
}

func FilterTypeString(value string) (any, error) {
	return value, nil
}

func FilterTypeInt(value string) (any, error) {
	return strconv.Atoi(value)
}

func FilterTypeBool(value string) (any, error) {
	return strconv.ParseBool(value)
}

func AnySliceToTyped[T any](obj []any) []T {
	result := make([]T, len(obj))

	for i, v := range obj {
		result[i] = v.(T) //nolint:forcetypeassert
	}

	return result
}

func TypedSliceToAny[T any](obj []T) []any {
	result := make([]any, len(obj))

	for i, v := range obj {
		result[i] = v
	}

	return result
}

// applyFiltering filters the records from the reposne based on the filters provider in the config
func (f FilterConfig) applyFiltering(filterSet []FilterModel, data []any) ([]any, diag.Diagnostic) {
	result := make([]any, 0)

	for _, elem := range data {
		match, d := f.matchesFilter(filterSet, elem)
		if d != nil {
			return nil, d
		}
		if !match {
			continue
		}
		result = append(result, elem)
	}

	return result, nil
}

// matchesFilter checks whether an object matches the given filter set
func (f FilterConfig) matchesFilter(filterSet []FilterModel, elem any) (bool, diag.Diagnostic) {
	for _, filter := range filterSet {
		filterName := filter.Name.ValueString()

		matchingField, d := resolveStructValueByJSON(elem, filterName)
		if d != nil {
			return false, d
		}

		match, d := f.checkFieldMatchesFilter(matchingField, filter)
		if d != nil {
			return false, d
		}

		if !match {
			return false, nil
		}
	}

	return true, nil
}

// checkFieldMatchesFilter checks whether an individual field
// meets the condition for the given filter
func (f FilterConfig) checkFieldMatchesFilter(field any, filter FilterModel) (bool, diag.Diagnostic) {
	rField := reflect.ValueOf(field)

	// Recursively filter on list elements
	if rField.Kind() == reflect.Slice {
		for i := 0; i < rField.Len(); i++ {
			match, d := f.checkFieldMatchesFilter(rField.Index(i).Interface(), filter)
			if d != nil {
				return false, d
			}

			if match {
				return true, nil
			}
		}

		return false, nil
	}

	normalizedValue, d := normalizeValue(field)
	if d != nil {
		return false, d
	}

	result := false
	d = nil

	switch strings.ToLower(filter.MatchBy.ValueString()) {
	case EXACT, "":
		result = checkFilterExact(filter.Values, normalizedValue)
	case SUBSTRING:
		result = checkFilterSubString(filter.Values, normalizedValue)
	case REGEX:
		result, d = checkFilterRegex(filter.Values, normalizedValue)
	}

	return result, d
}

// normalizeValue converts the given field into a comparable string
func normalizeValue(field any) (string, diag.Diagnostic) {
	rField := reflect.ValueOf(field)

	// Dereference if the value is a pointer
	for rField.Kind() == reflect.Pointer {
		// Null pointer; assume empty
		if rField.IsNil() {
			return "", nil
		}

		rField = reflect.Indirect(rField)
	}
	const timeFormat = "2006-01-02T15:04:05Z"

	// Special handler for time.Time values
	if t, ok := rField.Interface().(time.Time); ok {
		return t.Format(timeFormat), nil
	}

	switch rField.Kind() {
	case reflect.String:
		return rField.String(), nil
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(rField.Int(), 10), nil
	case reflect.Bool:
		return strconv.FormatBool(rField.Bool()), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rField.Float(), 'f', 0, 64), nil
	default:
		return "", diag.NewErrorDiagnostic(
			"Invalid field type",
			fmt.Sprintf("Invalid type for field: %s", rField.Type().String()),
		)
	}
}

func checkFilterExact(values []types.String, actualValue string) bool {
	for _, value := range values {
		if reflect.DeepEqual(actualValue, value.ValueString()) {
			return true
		}
	}

	return false
}

func checkFilterSubString(values []types.String, actualValue string) bool {
	for _, value := range values {
		if strings.Contains(actualValue, value.ValueString()) {
			return true
		}
	}

	return false
}

func checkFilterRegex(values []types.String, actualValue string) (bool, diag.Diagnostic) {
	for _, value := range values {
		r, err := regexp.Compile(value.ValueString())
		if err != nil {
			return false, diag.NewErrorDiagnostic(
				"failed to compile regex",
				err.Error(),
			)
		}

		if r.MatchString(actualValue) {
			return true, nil
		}
	}

	return false, nil
}

// resolveStructFieldByJSON resolves the struct field resolves the StructField
// with the given JSON tag.
func resolveStructFieldByJSON(val any, field string) (reflect.StructField, diag.Diagnostic) {
	rType := reflect.TypeOf(val)

	for i := 0; i < rType.NumField(); i++ {
		currentField := rType.Field(i)
		if tag, ok := currentField.Tag.Lookup("json"); ok && tag == field {
			return currentField, nil
		}

		// If there is no JSON tag, compare against the lowercase field name.
		// This is a workaround for untagged fields
		if strings.EqualFold(field, currentField.Name) {
			return currentField, nil
		}

		// If there is a json tag in camelCase
		if strings.ReplaceAll(field, "_", "") == strings.ToLower(currentField.Name) {
			return currentField, nil
		}
	}

	return reflect.StructField{}, diag.NewErrorDiagnostic(
		"Missing field",
		fmt.Sprintf("Failed to find field %s in struct.", field),
	)
}

// resolveStructValueByJSON resolves the corresponding value of a struct field
// given a JSON tag.
func resolveStructValueByJSON(val any, field string) (any, diag.Diagnostic) {
	structField, d := resolveStructFieldByJSON(val, field)
	if d != nil {
		return nil, d
	}

	targetField := reflect.ValueOf(val).FieldByName(structField.Name)

	if !targetField.IsValid() {
		return nil, diag.NewErrorDiagnostic(
			"Field not found",
			fmt.Sprintf("Could not find JSON tag in target struct: %s", field),
		)
	}

	return targetField.Interface(), nil
}

func (f FilterConfig) validateFilterable() filterNameValidator {
	return filterNameValidator{
		FilterConfig: f,
	}
}
