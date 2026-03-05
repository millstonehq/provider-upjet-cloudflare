// Package config contains the provider configuration.
package config

import (
	_ "embed"

	tjconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/millstonehq/provider-upjet-cloudflare/config/dns"
	"github.com/millstonehq/provider-upjet-cloudflare/config/zone"
)

const (
	resourcePrefix = "cloudflare"
	modulePath     = "github.com/millstonehq/provider-upjet-cloudflare"
)

//go:embed schema.json
var providerSchema []byte

// GetProvider returns provider configuration
func GetProvider() *tjconfig.Provider {
	pc := tjconfig.NewProvider(
		providerSchema,  // Schema extracted by OpenTofu
		resourcePrefix,
		modulePath,
		[]byte{},        // Empty metadata
		tjconfig.WithRootGroup("cloudflare.millstone.tech"),
		tjconfig.WithFeaturesPackage("internal/features"),
		tjconfig.WithIncludeList([]string{
			"cloudflare_.*",
		}),
		tjconfig.WithDefaultResourceOptions(
			func(r *tjconfig.Resource) {
				r.ExternalName = tjconfig.IdentifierFromProvider
			},
		),
	)

	// Configure individual resources
	for _, configure := range []func(*tjconfig.Provider){
		dns.Configure,
		zone.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
