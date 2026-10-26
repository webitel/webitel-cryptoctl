// Package schema holds the Webitel field-encryption schemas, one per database schema
// (directory, chat, call_center), as github.com/webitel/crypto/cryptostore/schema
// configs.
//
// Each table is described in exactly one file, so services sharing a table share its
// definition. Services register the config of their database schema as the immutable
// baseline:
//
//	schema.Register(wcrypto.CallCenter) // cryptostore/schema; wcrypto = this package
//
// The same .jsonc files ship in the webitel-cryptoctl package (schema.d) for cryptoctl,
// so the tool knows exactly what the services encrypt.
package schema
