package schema_test

import cryptoschema "github.com/webitel/crypto/cryptostore/schema"

// reference is the developer's consolidated Webitel schema (cmd/cryptoctl/webitel.go,
// built with -tags=webitel before this repository existed), kept verbatim: the split
// per-database-schema files must merge back into exactly this.
var reference = cryptoschema.Config{
	Version: 1,
	Units: map[string]*cryptoschema.Unit{
		// ---------------------------------------------- //
		//   schema:   directory                          //
		//   project:  webitel.go                         //
		// ---------------------------------------------- //
		"directory.wbt_oauth_service": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"client_secret": {},
			},
		},
		"directory.wbt_user_token": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"token": {Search: true},
			},
		},
		"directory.wbt_session": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"code":    {Search: true},
				"access":  {Search: true},
				"refresh": {Search: true},
				"props": {Nested: []cryptoschema.FieldNested{
					{Path: []string{"csrf"}},                // Cookie
					{Path: []string{"pn-rpid"}},             // SSE PUSH
					{Path: []string{"oauth.id_token"}},      // SSO Creds
					{Path: []string{"oauth.access_token"}},  // SSO Creds
					{Path: []string{"oauth.refresh_token"}}, // SSO Creds
				}},
			},
		},
		"directory.wbt_ldap": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"password": {},
			},
		},
		"directory.wbt_totp": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"key": {},
			},
		},
		// ---------------------------------------------- //
		//   schema:   chat                               //
		//   project:  chat_manager                       //
		// ---------------------------------------------- //
		"chat.bot": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"metadata": {Nested: []cryptoschema.FieldNested{
					// custom
					{Path: []string{"secret"}},
					// viber
					// telegram
					{Path: []string{"token"}},
					// infobip_whatsapp
					{Path: []string{"api_key"}},
					// gotd
					{Path: []string{"api_hash"}},
					// gotd: backup data
					{Path: []string{".auth"}},
					{Path: []string{".gotd"}},
					// messenger
					{Path: []string{"client_secret"}},
					{Path: []string{"whatsapp_token"}},
					// messenger: backup data
					{Path: []string{"fb"}},
					{Path: []string{"ig"}},
					{Path: []string{"wa"}},
				}},
			},
		},
		// ---------------------------------------------- //
		//   schema:   call_center                        //
		//   project:  engine & flow_manager              //
		// ---------------------------------------------- //
		"call_center.cc_email_profile": {
			Fields: map[string]*cryptoschema.FieldPolicy{
				"password": {},
				"params": {Nested: []cryptoschema.FieldNested{
					{Path: []string{"$.oauth2.client_secret"}},
				}},
				"token": {Nested: []cryptoschema.FieldNested{
					{Path: []string{"access_token"}},
					{Path: []string{"refresh_token"}},
				}},
			},
		},
	},
}
