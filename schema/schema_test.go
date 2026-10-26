package schema_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	cryptoschema "github.com/webitel/crypto/cryptostore/schema"

	"github.com/webitel/webitel-cryptoctl/schema"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMergedEqualsReference(t *testing.T) {
	got := cryptoschema.MergeConfig(cryptoschema.MergeConfig(schema.Directory, schema.Chat), schema.CallCenter)
	if !reflect.DeepEqual(got, &reference) {
		t.Fatalf("merged schema differs from reference:\n got %s\nwant %s", mustJSON(t, got), mustJSON(t, &reference))
	}
}

var configs = map[string]*cryptoschema.Config{
	"directory":   schema.Directory,
	"chat":        schema.Chat,
	"call_center": schema.CallCenter,
}

func TestUnitsDisjoint(t *testing.T) {
	owner := map[string]string{}
	for name, c := range configs {
		for unit := range c.Units {
			if prev, dup := owner[unit]; dup {
				t.Errorf("unit %q is described in both %s and %s", unit, prev, name)
			}
			owner[unit] = name
		}
	}
}

func TestUnitsMatchFilePrefix(t *testing.T) {
	for name, c := range configs {
		if len(c.Units) == 0 {
			t.Errorf("%s: no units", name)
		}
		for unit := range c.Units {
			if !strings.HasPrefix(unit, name+".") {
				t.Errorf("%s: unit %q does not belong to schema %q", name, unit, name)
			}
		}
	}
}

// TestFlatDottedKeyPreserved: props keys like "oauth.id_token" are flat keys, not paths.
func TestFlatDottedKeyPreserved(t *testing.T) {
	props := schema.Directory.Units["directory.wbt_session"].Fields["props"]
	for _, attr := range props.Nested {
		if strings.HasPrefix(attr.Path[0], "oauth.") && len(attr.Path) != 1 {
			t.Errorf("props attribute %v split into segments; want one flat key", attr.Path)
		}
	}
	want := cryptoschema.FieldNested{Path: []string{"oauth.id_token"}}
	for _, attr := range props.Nested {
		if reflect.DeepEqual(attr, want) {
			return
		}
	}
	t.Errorf("props has no flat key %q", "oauth.id_token")
}
