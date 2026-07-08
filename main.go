package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/mulesoft-anypoint/terraform-provider-anypoint/anypoint"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: "anypoint.mulesoft.com/automation/anypoint",
		ProviderFunc: func() *schema.Provider {
			return anypoint.Provider()
		},
	}

	plugin.Serve(opts)
}
