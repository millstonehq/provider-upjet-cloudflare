// Package dns contains configuration for Cloudflare DNS resources.
package dns

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

	if f.name != "cloudflare_dns_record" {
		t.Fatalf("registered for %q, want cloudflare_dns_record", f.name)
	}

	// Execute the captured configurator on a fresh Resource and assert fields.
	r := &config.Resource{}
	f.fn(r)

	if r.ExternalName.GetExternalNameFn == nil {
		t.Error("ExternalName not configured (expected IdentifierFromProvider)")
	}
	if r.ShortGroup != "dns" {
		t.Errorf("ShortGroup = %q, want %q", r.ShortGroup, "dns")
	}
	if r.Kind != "Record" {
		t.Errorf("Kind = %q, want %q", r.Kind, "Record")
	}
	if r.UseAsync {
		t.Error("UseAsync = true, want false")
	}
}
