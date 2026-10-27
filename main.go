// Command cryptoctl is the Webitel build of github.com/webitel/crypto cryptoctl: the
// Webitel field schemas are compiled in as the baseline.
package main

import (
	"github.com/webitel/crypto/cmd/cryptoctl/cli"
	"github.com/webitel/crypto/cryptostore/schema"

	wcrypto "github.com/webitel/webitel-cryptoctl/schema"
)

func init() {
	for _, c := range []*schema.Config{
		wcrypto.Directory,
		wcrypto.Chat,
		wcrypto.CallCenter,
		wcrypto.Email,
		wcrypto.KB,
	} {
		schema.Register(c)
	}
}

func main() { cli.Main() }
