// Package main provides a terraform-to-tofu wrapper that fixes Cloudflare v5
// provider compatibility with upjet. The Cloudflare v5 terraform provider's
// Read function crashes when called with an empty ID, which upjet triggers on
// new resources during terraform's implicit refresh. This wrapper:
//
//  1. Strips empty-ID resources from tfstate before ANY apply, including
//     `apply -refresh-only`, which is what upjet runs for Observe. A resource with
//     an empty ID does not exist yet, so removing it from state means terraform
//     never calls Read on it. This covers every resource without a list, which is
//     what a provider family needs -- see internal/controller/setup_custom.go for
//     the per-resource Go decorator this replaces.
//  2. Injects -refresh=false for plain apply ONLY (not -refresh-only) to prevent
//     terraform from calling Read with the now-valid ID during apply's implicit
//     refresh step. On -refresh-only it would disable the refresh the command
//     exists to perform, and Observe would stop detecting drift.
//
// Install as /usr/local/bin/terraform (before the tofu symlink in PATH).
package main

import (
	"encoding/json"
	"os"
	"strings"
	"syscall"
)

func main() {
	args := os.Args[1:]

	// STRIPPING RUNS ON *ANY* apply, INCLUDING -refresh-only, AND THAT IS THE POINT.
	//
	// These two behaviours used to sit behind one predicate, so Observe -- which upjet
	// runs as `apply -refresh-only` -- got neither. That gap was covered instead by a Go
	// decorator in internal/controller/setup_custom.go which short-circuits Observe per
	// resource, naming each one by hand. Readable for two resources; impossible for the
	// 211 a provider family publishes. So the protection moves here: this wrapper already
	// intercepts every terraform invocation the provider makes, so it covers every
	// resource without a list to maintain.
	//
	// Stripping is correct on the refresh path for the same reason it is correct on
	// apply. A resource with an empty ID does not exist yet; removing it from state means
	// terraform never calls Read on it. Cloudflare v5's Read crashes on an empty ID, and a
	// resource absent from state is never read -- the same outcome the Go decorator
	// produces by returning ResourceExists: false.
	if isApply(args) {
		stripEmptyIDResources("terraform.tfstate")
	}

	// -refresh=false STAYS apply-ONLY. On -refresh-only it would disable the very refresh
	// the command exists to perform, and Observe would stop detecting drift.
	if isPlainApply(args) {
		args = append(args, "-refresh=false")
	}

	tofu := "/usr/bin/tofu"
	syscall.Exec(tofu, append([]string{"tofu"}, args...), os.Environ())
}

// isPlainApply returns true for "apply" but false for "apply -refresh-only".
// isApply reports whether the invocation is any form of `apply`, including
// `apply -refresh-only`. Contrast isPlainApply, which deliberately excludes the
// refresh-only form.
func isApply(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg == "apply"
	}
	return false
}

func isPlainApply(args []string) bool {
	isApply := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if arg == "-refresh-only" {
				return false
			}
			continue
		}
		if !isApply {
			if arg == "apply" {
				isApply = true
			} else {
				return false
			}
		}
	}
	return isApply
}

func stripEmptyIDResources(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	resources, ok := state["resources"].([]interface{})
	if !ok || len(resources) == 0 {
		return
	}

	filtered := make([]interface{}, 0, len(resources))
	modified := false

	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		if !ok {
			filtered = append(filtered, r)
			continue
		}

		instances, ok := res["instances"].([]interface{})
		if !ok || len(instances) == 0 {
			filtered = append(filtered, r)
			continue
		}

		inst, ok := instances[0].(map[string]interface{})
		if !ok {
			filtered = append(filtered, r)
			continue
		}

		attrs, ok := inst["attributes"].(map[string]interface{})
		if !ok {
			filtered = append(filtered, r)
			continue
		}

		id, _ := attrs["id"].(string)
		if id == "" {
			modified = true
			continue // Strip this resource
		}

		filtered = append(filtered, r)
	}

	if !modified {
		return
	}

	state["resources"] = filtered
	out, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, out, 0644)
}
