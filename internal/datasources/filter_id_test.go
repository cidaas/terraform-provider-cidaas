package datasources

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFilterDatasourceID_StableAcrossOrder(t *testing.T) {
	a := FiltersModelType{
		{
			Name:    types.StringValue("role"),
			MatchBy: types.StringValue("exact"),
			Values:  []types.String{types.StringValue("b"), types.StringValue("a")},
		},
		{
			Name:    types.StringValue("name"),
			MatchBy: types.StringValue("substring"),
			Values:  []types.String{types.StringValue("x")},
		},
	}
	b := FiltersModelType{
		{
			Name:    types.StringValue("name"),
			MatchBy: types.StringValue("substring"),
			Values:  []types.String{types.StringValue("x")},
		},
		{
			Name:    types.StringValue("role"),
			MatchBy: types.StringValue("exact"),
			Values:  []types.String{types.StringValue("a"), types.StringValue("b")},
		},
	}

	idA := filterDatasourceID("cidaas_role", a)
	idB := filterDatasourceID("cidaas_role", b)
	if idA != idB {
		t.Fatalf("expected stable ID across filter/value order, got %q vs %q", idA, idB)
	}
	if !strings.HasPrefix(idA, "cidaas_role-") || len(idA) != len("cidaas_role-")+8 {
		t.Fatalf("unexpected ID format: %q", idA)
	}

	emptyA := filterDatasourceID("cidaas_role", nil)
	emptyB := filterDatasourceID("cidaas_role", FiltersModelType{})
	if emptyA != emptyB {
		t.Fatalf("empty filters should match, got %q vs %q", emptyA, emptyB)
	}
	if emptyA == idA {
		t.Fatalf("empty and non-empty filters must differ")
	}
}
