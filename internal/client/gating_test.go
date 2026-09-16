package client

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestRequireV4_AllowsFourX(t *testing.T) {
	c := &Client{Capabilities: Capabilities{TargetVersion: "4.x"}}
	var diags diag.Diagnostics
	if !c.RequireV4("cidaas_app_configuration", &diags) {
		t.Fatal("expected RequireV4 to succeed on 4.x")
	}
	if diags.HasError() {
		t.Fatal("expected no diagnostics")
	}
}

func TestRequireV4_RejectsThreeX(t *testing.T) {
	c := &Client{Capabilities: Capabilities{TargetVersion: "3.x"}}
	var diags diag.Diagnostics
	if c.RequireV4("cidaas_federation_provider", &diags) {
		t.Fatal("expected RequireV4 to fail on 3.x")
	}
	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}
}

func TestValidateResourceVersion_V4Compatible(t *testing.T) {
	c := &Client{
		Capabilities: Capabilities{
			TargetVersion: "4.x",
		},
	}

	var diags diag.Diagnostics
	ok := c.ValidateResourceVersion("cidaas_app_configuration", &diags)
	if !ok {
		t.Fatal("expected compatibility to succeed")
	}
	if diags.HasError() {
		t.Fatal("expected no diagnostics error")
	}
}

func TestValidateResourceVersion_V3Incompatible(t *testing.T) {
	c := &Client{
		Capabilities: Capabilities{
			TargetVersion: "3.x",
		},
	}

	var diags diag.Diagnostics
	ok := c.ValidateResourceVersion("cidaas_app_configuration", &diags)
	if ok {
		t.Fatal("expected compatibility to fail")
	}
	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}
}

func TestValidateResourceVersion_LegacyCompatibleOnV3(t *testing.T) {
	c := &Client{
		Capabilities: Capabilities{
			TargetVersion: "3.x",
		},
	}

	var diags diag.Diagnostics
	ok := c.ValidateResourceVersion("cidaas_app", &diags)
	if !ok {
		t.Fatal("expected legacy app compatibility to succeed on 3.x")
	}
	if diags.HasError() {
		t.Fatal("expected no diagnostics error")
	}
}

func TestValidateResourceVersion_FederationRequiresV4(t *testing.T) {
	c := &Client{Capabilities: Capabilities{TargetVersion: "3.x"}}
	var diags diag.Diagnostics
	if c.ValidateResourceVersion("cidaas_federation_provider", &diags) {
		t.Fatal("expected federation to require 4.x")
	}
}
