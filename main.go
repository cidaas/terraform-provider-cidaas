package main

import (
	"context"
	"flag"
	"log"

	"github.com/Cidaas/terraform-provider-cidaas/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name cidaas

var (
	version = "dev"
	commit  = "none"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debugger support")
	flag.Parse()

	if debug {
		log.Printf("cidaas provider version=%s commit=%s", version, commit)
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/Cidaas/cidaas",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
