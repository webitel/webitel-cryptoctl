package schema

import (
	_ "embed"

	cryptoschema "github.com/webitel/crypto/cryptostore/schema"
)

var (
	//go:embed schema.d/directory.jsonc
	directoryJSON string
	//go:embed schema.d/chat.jsonc
	chatJSON string
	//go:embed schema.d/call_center.jsonc
	callCenterJSON string
	//go:embed schema.d/email.jsonc
	emailJSON string
	//go:embed schema.d/kb.jsonc
	kbJSON string
)

var (
	// Directory is the schema of the directory.* tables (webitel.go).
	Directory = cryptoschema.MustParseConfig(directoryJSON)
	// Chat is the schema of the chat.* tables (chat_manager).
	Chat = cryptoschema.MustParseConfig(chatJSON)
	// CallCenter is the schema of the call_center.* tables (engine, flow_manager).
	CallCenter = cryptoschema.MustParseConfig(callCenterJSON)
	// Email is the schema of the email.* tables (webitel-emails, dev only).
	Email = cryptoschema.MustParseConfig(emailJSON)
	// KB is the schema of the kb.* tables (webitel-kb, dev only).
	KB = cryptoschema.MustParseConfig(kbJSON)
)
