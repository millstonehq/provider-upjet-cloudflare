// Package zone contains configuration for Cloudflare Zone resources.
package zone

import (
	"testing"

	"github.com/crossplane/upjet/v2/pkg/config"
)

type fakeAdder struct {
	name string
	fn   config.ResourceConfiguratorFn
}

func (f *fakeAdder) AddResourceConfigurator(name string, rc config.ResourceConfiguratorFn) {
	f.name = name
	f.fn = rc
}

func TestConfigureRegistersConfigurator(t *testing.T) {
	f := &fakeAdder{}
	configureWithAdder(f)

	if f.name != "cloudflare_zone" {
		t.Fatalf("registered for %q, want cloudflare_zone", f.name)
	}

	// Execute the captured configurator on a fresh Resource and assert fields.
	r := &config.Resource{}
	f.fn(r)

	if r.ExternalName.GetExternalNameFn == nil {
		t.Error("ExternalName not configured (expected IdentifierFromProvider)")
	}
	if r.ShortGroup != "zone" {
		t.Errorf("ShortGroup = %q, want %q", r.ShortGroup, "zone")
	}
	if r.Kind != "Zone" {
		t.Errorf("Kind = %q, want %q", r.Kind, "Zone")
	}
	if r.UseAsync {
		t.Error("UseAsync = true, want false")
	}
}
