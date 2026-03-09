// Package main provides the entrypoint for the AAP Terraform provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashi-demo-lab/terraform-provider-aap/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// Example version string that can be overwritten by a release process
	version = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/hashi-demo-lab/aap",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
