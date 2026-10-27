package main

import (
	"path/filepath"
	"testing"

	cryptoschema "github.com/webitel/crypto/cryptostore/schema"
)

// TestBaselineHasAllSchemas: every table of schema/schema.d is in the compiled-in baseline.
func TestBaselineHasAllSchemas(t *testing.T) {
	base := cryptoschema.Baseline()
	if base == nil {
		t.Fatal("no baseline registered")
	}

	files, err := filepath.Glob("schema/schema.d/*.jsonc")
	if err != nil || len(files) == 0 {
		t.Fatalf("schema files: %v %v", files, err)
	}

	for _, f := range files {
		c, err := cryptoschema.LoadConfigFile(f)
		if err != nil {
			t.Fatal(err)
		}

		for unit := range c.Units {
			if base.Units[unit] == nil {
				t.Errorf("baseline lacks unit %s (%s)", unit, f)
			}
		}
	}
}
