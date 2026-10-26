package schema_test

import (
	"encoding/json"
	"reflect"
	"slices"
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
	"email":       schema.Email,
	"kb":          schema.KB,
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

func TestDevSchemas(t *testing.T) {
	want := map[string][]string{
		"email.profile":      {"oauth_client_secret", "oauth_refresh_token", "password"},
		"kb.embedding_model": {"config"},
	}
	for _, c := range []*cryptoschema.Config{schema.Email, schema.KB} {
		for unit, u := range c.Units {
			var got []string
			for f, p := range u.Fields {
				if p.Search || len(p.Nested) > 0 {
					t.Errorf("%s.%s: want a plain scalar field", unit, f)
				}
				got = append(got, f)
			}
			slices.Sort(got)
			if !slices.Equal(got, want[unit]) {
				t.Errorf("%s fields = %q, want %q", unit, got, want[unit])
			}
			delete(want, unit)
		}
	}
	for unit := range want {
		t.Errorf("unit %s missing", unit)
	}
}
