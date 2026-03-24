// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package main is the entry point for the terraform-provider-opnsense binary.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

// these will be set by the goreleaser configuration
// to appropriate values for the compiled binary.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/matthew-on-git/opnsense",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
