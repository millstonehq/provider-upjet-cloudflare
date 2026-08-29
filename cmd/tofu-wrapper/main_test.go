package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// THE REGRESSION THIS FILE EXISTS FOR. Stripping and -refresh=false used to sit behind one
// predicate, so `apply -refresh-only` -- the command upjet runs for Observe -- got neither, and
// the Observe path was protected instead by a hand-written Go decorator naming every resource
// individually. That does not scale to a provider family's 211 resources. These tests pin the
// split: strip on ANY apply, inject -refresh=false on plain apply only.

func TestIsApply(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"plain apply", []string{"apply"}, true},
		// The case the whole change is about.
		{"refresh-only apply", []string{"apply", "-refresh-only"}, true},
		{"flag before subcommand", []string{"-chdir=/w", "apply", "-refresh-only"}, true},
		{"apply with other flags", []string{"apply", "-auto-approve", "-input=false"}, true},
		{"plan is not apply", []string{"plan"}, false},
		{"init is not apply", []string{"init"}, false},
		{"no args", []string{}, false},
		{"flags only", []string{"-version"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isApply(tc.args); got != tc.want {
				t.Errorf("isApply(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestIsPlainApply(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"plain apply", []string{"apply"}, true},
		{"apply with flags", []string{"apply", "-auto-approve"}, true},
		// -refresh=false must NOT reach the refresh-only path: it would disable the
		// refresh the command exists to perform and Observe would stop seeing drift.
		{"refresh-only is excluded", []string{"apply", "-refresh-only"}, false},
		{"refresh-only before subcommand", []string{"-refresh-only", "apply"}, false},
		{"plan", []string{"plan"}, false},
		{"no args", []string{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPlainApply(tc.args); got != tc.want {
				t.Errorf("isPlainApply(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// writeState writes a tfstate containing one resource per supplied id.
func writeState(t *testing.T, ids ...string) string {
	t.Helper()
	resources := make([]any, 0, len(ids))
	for i, id := range ids {
		resources = append(resources, map[string]any{
			"type": "cloudflare_dns_record",
			"name": string(rune('a' + i)),
			"instances": []any{
				map[string]any{"attributes": map[string]any{"id": id}},
			},
		})
	}
	b, err := json.Marshal(map[string]any{"version": 4, "resources": resources})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "terraform.tfstate")
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func remainingIDs(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var st map[string]any
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range st["resources"].([]any) {
		inst := r.(map[string]any)["instances"].([]any)[0].(map[string]any)
		out = append(out, inst["attributes"].(map[string]any)["id"].(string))
	}
	return out
}

func TestStripEmptyIDResources(t *testing.T) {
	t.Run("drops empty ids and keeps the rest", func(t *testing.T) {
		p := writeState(t, "abc123", "", "def456")
		stripEmptyIDResources(p)
		got := remainingIDs(t, p)
		if len(got) != 2 || got[0] != "abc123" || got[1] != "def456" {
			t.Errorf("remaining ids = %q, want [abc123 def456]", got)
		}
	})

	t.Run("leaves a clean state untouched", func(t *testing.T) {
		p := writeState(t, "abc123")
		before, _ := os.ReadFile(p)
		stripEmptyIDResources(p)
		after, _ := os.ReadFile(p)
		if string(before) != string(after) {
			t.Error("state with no empty ids was rewritten; it should be left byte-identical")
		}
	})

	// The wrapper execs tofu regardless, so a state it cannot parse must not stop the
	// build -- a crash here would break every terraform call, not just the edge case.
	t.Run("survives a missing file", func(t *testing.T) {
		stripEmptyIDResources(filepath.Join(t.TempDir(), "absent.tfstate"))
	})

	t.Run("survives malformed json", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "terraform.tfstate")
		if err := os.WriteFile(p, []byte("{not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		stripEmptyIDResources(p)
	})
}
