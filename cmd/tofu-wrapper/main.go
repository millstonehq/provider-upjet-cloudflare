// Package main provides a terraform-to-tofu wrapper that fixes Cloudflare v5
// provider compatibility with upjet. The Cloudflare v5 terraform provider's
// Read function crashes when called with an empty ID, which upjet triggers on
// new resources during terraform's implicit refresh. This wrapper strips
// empty-ID resources from tfstate before apply so terraform correctly plans a
// Create instead of crashing during refresh.
//
// Install as /usr/local/bin/terraform (before the tofu symlink in PATH).
// Pair with TF_CLI_ARGS_apply=-refresh=false to prevent Read during apply.
package main

import (
	"encoding/json"
	"os"
	"strings"
	"syscall"
)

func main() {
	args := os.Args[1:]

	if isApplyCommand(args) {
		stripEmptyIDResources("terraform.tfstate")
	}

	tofu := "/usr/bin/tofu"
	syscall.Exec(tofu, append([]string{"tofu"}, args...), os.Environ())
}

func isApplyCommand(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg == "apply"
	}
	return false
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
