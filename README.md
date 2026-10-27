# webitel-cryptoctl

Схеми шифрування полів Webitel і їхня поставка для `cryptoctl`.

- **Go-модуль** `github.com/webitel/webitel-cryptoctl/schema` (окремий `schema/go.mod`, лише
  `cryptostore`): baseline-схеми, які сервіси вкомпільовують (`schema.Register(wcrypto.CallCenter)`).
- **Webitel-збірка `cryptoctl`** (корінь репо, `main.go`): `cli.Main()` з
  `github.com/webitel/crypto/cmd/cryptoctl/cli`, а схеми вкомпільовано як **baseline**
  (`schema.Register`): адмін-конфіг їх не послабить і не загубить.
- **Deb-пакет `webitel-cryptoctl`**: цей `cryptoctl`, конфіг сховища і обгортка
  `webitel-cryptoctl` для ops (`status`, `usage`, `migrate`, `can-retire`, …).

Бібліотека й утиліта живуть у [`github.com/webitel/crypto`](https://github.com/webitel/crypto)
і не містять нічого Webitel-специфічного. Тут лише те, що стосується Webitel.

## Схеми

Файли розкладені **за схемами БД**, і кожна таблиця описана рівно в одному з них:

| Файл | Таблиці | Сервіси |
|---|---|---|
| `schema/schema.d/directory.jsonc` | `directory.*` | webitel.go |
| `schema/schema.d/chat.jsonc` | `chat.*` | chat_manager |
| `schema/schema.d/call_center.jsonc` | `call_center.*` | engine, flow_manager |
| `schema/schema.d/email.jsonc` | `email.*` | webitel-emails (лише `main`) |
| `schema/schema.d/kb.jsonc` | `kb.*` | webitel-kb (лише `main`) |

webitel-emails і webitel-kb шифрують поля через `cryptostore.Codec` напряму, а не через
`cryptostore/schema`, тож для них ці файли — копія їхнього коду: зміна набору шифрованих
полів у сервісі йде разом із PR сюди.

Формат — `cryptostore/schema.Config` у JSONC. Ключі `props` у `directory.wbt_session` на
кшталт `oauth.id_token` **пласкі** (ключ з крапкою), це не вкладені шляхи.

### Як змінити схему

1. PR у цей репо: зміна `.jsonc` у `schema/schema.d/`. Новий файл додається в
   `schema/schema.go` і в список `main.go`.
2. Теги `schema/vX.Y.Z` і deb `vA.B.C` з `A.B.C >= X.Y.Z` (лише на гілці релізу, див. нижче),
   а потім `go get github.com/webitel/webitel-cryptoctl/schema@vX.Y.Z` у **всіх** сервісах, які пишуть у змінену таблицю. Вони виходять разом: якщо один сервіс
   шифрує поле, а інший про це не знає, вони ламають дані одне одному в БД.
3. Deb `webitel-cryptoctl` публікується **раніше** за сервіси або одночасно з ними. Сервіс на
   гілці релізу залежить від `webitel-cryptoctl (>= версія з його go.mod)`.

## Версії та гілки

- Semver, два модулі з власними тегами:
  - `schema/vX.Y.Z` — Go-модуль схем, його версія стоїть у `go.mod` сервісів;
  - `vA.B.C` — deb `webitel-cryptoctl` (схеми + `cryptoctl`).
- **Версія deb ≥ версії схем.** Сервіс на релізі залежить від `webitel-cryptoctl (>= X.Y.Z)`,
  де `X.Y.Z` — версія `schema` у його `go.mod`. Тож зміна схеми завжди виходить разом із deb
  такої самої або вищої версії; новий `cryptoctl` без зміни схем — лише тег deb.
- Кожна лінійка продукту має свою гілку `vYY.MM` і свій minor (26.10 → `0.1.x`).
  Хотфікси — patch-теги на гілці релізу.
- Теги ставляться **лише на гілках `vYY.MM`**, не на `main`. Тег публікується в компонент своєї
  гілки (`26.10-releases`), тож інсталяції різних релізів не бачать пакетів одна одної.
- `main` збирається як dev (`~dev`-версія). На `main` сервіси залежать від
  `webitel-cryptoctl` без версії: `~dev` у dpkg менша за реліз, тож `>=` там не виконався б.

### Оновити `cryptoctl` у лінійці

```bash
git switch v26.10
go get github.com/webitel/crypto/cmd/cryptoctl@vA.B.C
go mod tidy
# PR → тег наступного patch на v26.10
```

## Пакет `webitel-cryptoctl`

```
/usr/share/webitel/crypto/store.d/pg.jsonc   одне postgres-сховище (схеми — baseline у бінарнику)
/usr/local/bin/cryptoctl                     Webitel-збірка cryptoctl
/usr/bin/webitel-cryptoctl                   обгортка
/etc/default/webitel-cryptoctl               налаштування (conffile, 0640)
```

`webitel-cryptoctl` бере налаштування з `/etc/default/webitel-cryptoctl` (підключення `PG*`,
перевизначення `CRYPTO_*`), задає `CRYPTO_STORE_DIR` на `store.d` пакета і запускає
`cryptoctl`. Змінна, задана під час виклику, важить більше за файл:

```bash
webitel-cryptoctl status
PGHOST=db2 webitel-cryptoctl can-retire 1
```

Ключі ті самі, що в сервісів: `CRYPTO_CIPHER_*`/`CRYPTO_SEARCH_*` або
`/var/lib/webitel/crypto/.cipher` / `.search`. Цей каталог вшитий у збірку через
`-X github.com/webitel/crypto/cryptobox.defaultSystemDir=/var/lib/webitel/crypto`.

## Розробка

```bash
go test ./...                  # схеми + пакувальні файли (обгортка, store.d, .env.example)
TEST_SHELL=/bin/dash go test . # обгортка під dash (як /bin/sh у Debian)
```
