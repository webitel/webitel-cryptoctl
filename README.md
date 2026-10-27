# webitel-cryptoctl

Field-encryption schemas of the platform services and a
[`cryptoctl`](https://github.com/webitel/crypto/tree/main/cmd/cryptoctl) build with them
compiled in.

Services encrypt sensitive columns (passwords, tokens, OAuth secrets) with these schemas;
`cryptoctl` uses the same schemas to report key usage, re-encrypt data to the current key
(`migrate`) and check whether an old key can be retired.

See the [`cryptoctl` documentation](https://github.com/webitel/crypto/tree/main/cmd/cryptoctl)
for commands, configuration and key management.

## Changing a schema

A new file in `schema/schema.d/` also goes into `schema/schema.go` and `main.go`.

```bash
# webitel-cryptoctl, on the release branch
git switch v26.10
$EDITOR schema/schema.d/call_center.jsonc
go test ./... && (cd schema && go test ./...)
git commit -am "feat(schema): ..." && git push

# schema module and deb; the deb version must be >= the schema version
git tag schema/vX.Y.Z && git tag vA.B.C
git push origin schema/vX.Y.Z vA.B.C

# every service writing the changed table; release them together, after the deb
go get github.com/webitel/webitel-cryptoctl/schema@vX.Y.Z && go mod tidy
```

## Usage

Settings (`PG*`, `CRYPTO_*`) go to `/etc/default/webitel-cryptoctl`; the caller's environment
wins. Keys are the services' keys (`CRYPTO_CIPHER_*` / `CRYPTO_SEARCH_*` or
`/var/lib/webitel/crypto/.cipher` / `.search`).

```bash
webitel-cryptoctl status
PGHOST=db2 webitel-cryptoctl can-retire 1
```

Tables missing from a database (service not installed) are skipped.

## Rollout

1. Migrate column types: encrypted scalars to `bytea`, add `__cs_idx jsonb` to tables with
   searchable fields.
2. Deploy the services.
3. Run `webitel-cryptoctl migrate` **right after**. Until then rows written before the deploy
   have no search tokens, lookups by token miss them and services create duplicates, which
   later break unique indexes on `__cs_idx`.

## Development

```bash
go test ./... && (cd schema && go test ./...)
TEST_SHELL=/bin/dash go test .   # the wrapper under Debian's /bin/sh
```
