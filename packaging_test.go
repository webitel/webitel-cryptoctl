package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/webitel/crypto/encoding/jsonc"
)

const (
	wrapperEnvFile = "/etc/default/webitel-cryptoctl"
	wrapperBinary  = "/usr/local/bin/cryptoctl"
)

// wrapper runs a copy of bin/webitel-cryptoctl against a fake cryptoctl that prints its env and args.
type wrapper struct {
	dir, script, envFile string
}

func newWrapper(t *testing.T) *wrapper {
	t.Helper()
	src, err := os.ReadFile("bin/webitel-cryptoctl")
	if err != nil {
		t.Fatal(err)
	}
	w := &wrapper{dir: t.TempDir()}
	w.envFile = filepath.Join(w.dir, "default")
	fake := filepath.Join(w.dir, "cryptoctl")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nenv\nfor a in \"$@\"; do printf 'ARG=%s\\n' \"$a\"; done\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := strings.ReplaceAll(string(src), wrapperEnvFile, w.envFile)
	script = strings.ReplaceAll(script, wrapperBinary, fake)
	w.script = filepath.Join(w.dir, "webitel-cryptoctl")
	if err := os.WriteFile(w.script, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return w
}

func (w *wrapper) writeEnvFile(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(w.envFile, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
}

func (w *wrapper) run(t *testing.T, env []string, args ...string) (vars map[string]string, argv []string) {
	t.Helper()
	shell := os.Getenv("TEST_SHELL") // e.g. /bin/dash
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell, append([]string{w.script}, args...)...)
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH")}, env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wrapper: %v\n%s", err, out)
	}
	vars = map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		if arg, ok := strings.CutPrefix(line, "ARG="); ok {
			argv = append(argv, arg)
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			vars[k] = v
		}
	}
	return vars, argv
}

func TestWrapperDefaults(t *testing.T) {
	w := newWrapper(t)
	vars, argv := w.run(t, nil, "status", "a b")
	if got, want := vars["CRYPTO_STORE_DIR"], "/usr/share/webitel/crypto/store.d"; got != want {
		t.Errorf("CRYPTO_STORE_DIR = %q, want %q", got, want)
	}
	if want := []string{"status", "a b"}; !slices.Equal(argv, want) {
		t.Errorf("args = %q, want %q", argv, want)
	}
}

func TestWrapperEnvFileApplied(t *testing.T) {
	w := newWrapper(t)
	w.writeEnvFile(t, "# comment\nPGHOST=db1\n#PGPORT=5432\n\nCRYPTO_DIR=/srv/keys\n")
	vars, _ := w.run(t, nil)
	if vars["PGHOST"] != "db1" || vars["CRYPTO_DIR"] != "/srv/keys" {
		t.Errorf("PGHOST = %q, CRYPTO_DIR = %q; want db1, /srv/keys", vars["PGHOST"], vars["CRYPTO_DIR"])
	}
	if _, set := vars["PGPORT"]; set {
		t.Errorf("commented PGPORT was exported")
	}
}

func TestWrapperQuotedValue(t *testing.T) {
	w := newWrapper(t)
	w.writeEnvFile(t, "PGPASSWORD=\"p w\"\nPGUSER='u x'\n")
	vars, _ := w.run(t, nil)
	if vars["PGPASSWORD"] != "p w" || vars["PGUSER"] != "u x" {
		t.Errorf("PGPASSWORD = %q, PGUSER = %q; want \"p w\", \"u x\"", vars["PGPASSWORD"], vars["PGUSER"])
	}
}

func TestWrapperCallerEnvWins(t *testing.T) {
	w := newWrapper(t)
	w.writeEnvFile(t, "PGHOST=db1\nPGUSER=file\n")
	vars, _ := w.run(t, []string{"PGHOST=db2"})
	if vars["PGHOST"] != "db2" {
		t.Errorf("PGHOST = %q, want caller's db2 over the env file", vars["PGHOST"])
	}
	if vars["PGUSER"] != "file" {
		t.Errorf("PGUSER = %q, want file (caller did not set it)", vars["PGUSER"])
	}
}

func TestWrapperMissingEnvFile(t *testing.T) {
	w := newWrapper(t) // no env file written
	vars, _ := w.run(t, nil)
	if vars["CRYPTO_STORE_DIR"] == "" {
		t.Errorf("fake cryptoctl not reached or CRYPTO_STORE_DIR unset")
	}
}

func TestWrapperStoreDirOverride(t *testing.T) {
	w := newWrapper(t)
	vars, _ := w.run(t, []string{"CRYPTO_STORE_DIR=/x"})
	if vars["CRYPTO_STORE_DIR"] != "/x" {
		t.Errorf("CRYPTO_STORE_DIR = %q, want /x", vars["CRYPTO_STORE_DIR"])
	}
}

// TestStoreConfigUsesBaseline: the store carries no schema of its own; the compiled-in
// baseline (main.go) applies to it.
func TestStoreConfigUsesBaseline(t *testing.T) {
	raw, err := os.ReadFile("store.d/pg.jsonc")
	if err != nil {
		t.Fatal(err)
	}
	var store struct {
		DSN    string          `json:"dsn"`
		Schema json.RawMessage `json:"schema"`
	}
	if err := jsonc.Unmarshal(raw, &store); err != nil {
		t.Fatalf("store.d/pg.jsonc: %v", err)
	}
	if store.DSN != "postgres:" {
		t.Errorf("dsn = %q, want %q (connection from libpq environment)", store.DSN, "postgres:")
	}
	if len(store.Schema) != 0 {
		t.Errorf("store.d/pg.jsonc has a schema section %s; the baseline is compiled in", store.Schema)
	}
}

func TestEnvExampleAllCommented(t *testing.T) {
	raw, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			t.Errorf(".env.example:%d: %q is not commented out", i+1, line)
		}
	}
}

// TestWrapperValuesAreLiteral: values are literal, as in a systemd EnvironmentFile.
func TestWrapperValuesAreLiteral(t *testing.T) {
	w := newWrapper(t)
	marker := filepath.Join(w.dir, "executed")
	w.writeEnvFile(t, "PGPASSWORD=pa$$word\n"+
		"CRYPTO_CIPHER_KEYRING=[1]ab$PWD;x&y|z\n"+
		"PGUSER=\"q$(touch "+marker+")\"\n"+
		"PGHOST=a;touch "+marker+"\n")
	vars, _ := w.run(t, nil)
	want := map[string]string{
		"PGPASSWORD":            "pa$$word",
		"CRYPTO_CIPHER_KEYRING": "[1]ab$PWD;x&y|z",
		"PGUSER":                "q$(touch " + marker + ")",
		"PGHOST":                "a;touch " + marker,
	}
	for k, v := range want {
		if vars[k] != v {
			t.Errorf("%s = %q, want literal %q", k, vars[k], v)
		}
	}
	if _, err := os.Stat(marker); err == nil {
		t.Errorf("env file content was executed")
	}
}

func TestWrapperLastAssignmentWins(t *testing.T) {
	w := newWrapper(t)
	w.writeEnvFile(t, "PGHOST=db1\nPGHOST=db2\n")
	if vars, _ := w.run(t, nil); vars["PGHOST"] != "db2" {
		t.Errorf("PGHOST = %q, want the last assignment db2", vars["PGHOST"])
	}
}
