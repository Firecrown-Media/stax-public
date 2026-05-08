Reference for all `stax migrate` subcommands and flags.

Requires `migration.destination: vip` in `.stax.yml`, or pass `--destination=vip` on any command.

---

## stax migrate pull

Download wp-content from the source provider. Uploads are excluded by default.

```bash
stax migrate pull [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Show what would be transferred without pulling |
| `--themes-only` | bool | false | Pull only `wp-content/themes/` |
| `--plugins-only` | bool | false | Pull only `wp-content/plugins/` |
| `--mu-plugins-only` | bool | false | Pull only `wp-content/mu-plugins/` |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate pull
stax migrate pull --themes-only
stax migrate pull --dry-run
```

---

## stax migrate export

Export the source database with mysqldump flags required by WordPress VIP (`--hex-blob`, `--quote-names`, `--default-character-set=utf8mb4`).

```bash
stax migrate export [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `.stax/<install>-export.sql` | Path for the SQL dump file |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate export
stax migrate export --output=mysite-export.sql
```

---

## stax migrate audit

Scan plugins, themes, and client-mu-plugins against the WordPress-VIP-Go phpcs ruleset.

```bash
stax migrate audit [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | `<project>/wp-content` | Path to wp-content directory |
| `--severity` | int | 1 | Minimum phpcs severity level (1–5) |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate audit
stax migrate audit --path=../my-site/wp-content
stax migrate audit --severity=3
```

---

## stax migrate compare

Diff the downloaded WPEngine wp-content against a local VIP repo checkout.

```bash
stax migrate compare [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--vip-repo` | string | — | Path to local VIP repo checkout (required) |
| `--path` | string | `<project>/wp-content` | Path to downloaded wp-content |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate compare --vip-repo=../vip-repo
stax migrate compare --path=../wpe/wp-content --vip-repo=../vip-repo
```

---

## stax migrate import

Validate the SQL dump and import it into the VIP destination using the VIP CLI.

```bash
stax migrate import [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sql` | string | — | Path to the SQL dump file (required) |
| `--slug` | string | — | VIP environment slug |
| `--dry-run` | bool | false | Validate without importing |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate import --sql=.stax/mysite-export.sql
stax migrate import --sql=export.sql --slug=my-vip-env
stax migrate import --sql=export.sql --dry-run
```

---

## stax migrate report

Run audit and compare, then write a combined migration report to `.stax/migration-report.md`.

```bash
stax migrate report [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--vip-repo` | string | — | Path to local VIP repo checkout |
| `--path` | string | `<project>/wp-content` | Path to wp-content |
| `--sql` | string | — | Path to SQL dump file (optional) |
| `--output` | string | `.stax/migration-report.md` | Output path for the report |
| `--destination` | string | from config | Override migration destination |

**Examples:**

```bash
stax migrate report --vip-repo=../vip-repo
stax migrate report --path=../wpe/wp-content --vip-repo=../vip-repo --sql=.stax/mysite-export.sql
```

---

## stax migrate status

Print migration configuration and the presence of key artifacts (SQL export, report file).

```bash
stax migrate status
```

No flags beyond `--destination` and the global flags.
