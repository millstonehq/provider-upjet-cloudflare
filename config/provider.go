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
		providerSchema, // Schema extracted by OpenTofu
		resourcePrefix,
		modulePath,
		[]byte{}, // Empty metadata
		tjconfig.WithRootGroup("cloudflare.millstone.tech"),
		tjconfig.WithFeaturesPackage("internal/features"),
		// ONLY THE RESOURCES THIS PROVIDER ACTUALLY MANAGES, AND THE WILDCARD THAT USED TO BE HERE
		// COST A CLUSTER OUTAGE.
		//
		// "cloudflare_.*" matches every resource in the Cloudflare Terraform provider, so this
		// generated 211 CRDs to manage 12 objects -- 11 dns.Record and 1 zone.Zone, the only two
		// kinds ever configured below. Measured on a live cluster 2026-08-28: all 211 enumerated, 209
		// with zero instances.
		//
		// WHAT THE OTHER 209 COST. Each apiserver rebuilds an aggregated OpenAPI spec over every
		// CRD. On one such cluster that loop stopped converging on a node: 43 "slow openapi aggregation"
		// entries per 10 minutes at 1.5-5s each against 2 on a healthy peer, 1857m of apiserver CPU
		// against ~450m, and request handlers timing out on the lease endpoints. kube-scheduler lost
		// its lease and exited every ~4 minutes, and flannel could not watch Nodes through KubePrism,
		// restarted, rebuilt its VXLAN device, and left every pod on that node unreachable from every
		// other node -- which took the amd64 buildkitd offline and blocked all publishes. Nothing
		// reported any of it: the pods stayed Running and Ready throughout.
		//
		// ANCHORED AT BOTH ENDS, WHICH IS NOT DECORATION. upjet matches these with MatchString, so an
		// unanchored "cloudflare_zone" also selects cloudflare_zone_cache_reserve, _dnssec, _hold,
		// _lockdown and four more -- the eight CRDs the zone group had before this change.
		//
		// ADDING A RESOURCE MEANS ADDING A LINE HERE AND A configurator below, deliberately: the
		// generated surface should be a decision, not a default.
		tjconfig.WithIncludeList([]string{
			"^cloudflare_dns_record$",
			"^cloudflare_zone$",
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

	// NO RESERVED-WORD FIXES ARE NEEDED, and the block that used to sit here is gone rather than
	// kept "just in case". It renamed six resources whose Terraform names produce Go reserved-word
	// package names -- address_map, workers_for_platforms_dispatch_namespace, three
	// zero_trust_device_default_profile variants and snippet. None of them is generated any more,
	// so every entry was a configurator for a resource that does not exist: dead code that reads as
	// enforcement. If a future include-list line reintroduces one, git has the block.

	pc.ConfigureResources()
	return pc
}
