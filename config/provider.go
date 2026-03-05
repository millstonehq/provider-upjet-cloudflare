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

	// Fix resources that generate Go reserved word package names
	reservedWordFixes := map[string]struct {
		shortGroup string
		kind       string
	}{
		"cloudflare_address_map": {
			shortGroup: "addressmap",
			kind:       "AddressMap",
		},
		"cloudflare_workers_for_platforms_dispatch_namespace": {
			shortGroup: "workersplatforms",
			kind:       "DispatchNamespace",
		},
		"cloudflare_zero_trust_device_default_profile": {
			shortGroup: "zerotrust",
			kind:       "DeviceDefaultProfile",
		},
		"cloudflare_zero_trust_device_default_profile_certificates": {
			shortGroup: "zerotrust",
			kind:       "DeviceDefaultProfileCertificates",
		},
		"cloudflare_zero_trust_device_default_profile_local_domain_fallback": {
			shortGroup: "zerotrust",
			kind:       "DeviceDefaultProfileLocalDomainFallback",
		},
		"cloudflare_snippet": {
			shortGroup: "snippet",
			kind:       "Snippet",
		},
	}

	for name, fix := range reservedWordFixes {
		pc.AddResourceConfigurator(name, func(r *tjconfig.Resource) {
			r.ShortGroup = fix.shortGroup
			r.Kind = fix.kind
		})
	}

	pc.ConfigureResources()
	return pc
}
