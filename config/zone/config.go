// Package zone contains configuration for Cloudflare Zone resources.
package zone

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures the Zone resource.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("cloudflare_zone", func(r *config.Resource) {
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "zone"

		// Kind will be Zone
		r.Kind = "Zone"

		r.UseAsync = false
	})
}
