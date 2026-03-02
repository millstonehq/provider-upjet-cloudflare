// Package dns contains configuration for Cloudflare DNS resources.
package dns

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures the DNS Record resource.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("cloudflare_dns_record", func(r *config.Resource) {
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "dns"

		// Kind will be Record
		r.Kind = "Record"

		r.UseAsync = false
	})
}
