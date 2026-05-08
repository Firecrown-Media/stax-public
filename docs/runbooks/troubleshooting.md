Common stax errors with exact error messages and how to fix them.

---

### "DDEV not running"

Run `ddev status` to check the current state. If DDEV is stopped, start it:

```bash
stax start
```

If DDEV is running but stax still reports this, verify you are in the correct project directory and that `.stax.yml` exists there.

---

### "go: command not found" (make build)

Go is installed via Homebrew, which puts it in `/opt/homebrew/bin`. The `make` command uses `/bin/sh`, which does not source your shell profile.

```bash
PATH="/opt/homebrew/bin:$PATH" make build
```

---

### "failed to get WPEngine credentials"

Your WPEngine credentials are not stored in Keychain. Run setup:

```bash
stax setup
```

To verify credentials exist, check Keychain Access for entries under `com.firecrownmedia.stax`, or use `stax credentials get`.

---

### "rsync: [sender] change_dir failed: No such file or directory"

The remote path is incorrect or the install name in `.stax.yml` doesn't match your WPEngine install. Run:

```bash
stax doctor --fix
```

This checks the `provider_config.install` value and verifies that the remote path is reachable over SSH.

---

### "permission denied (publickey)" (SSH)

Your SSH key is not loaded or the path stored in Keychain doesn't match the key registered with WPEngine.

Verify the key works directly:

```bash
ssh -T install@ssh.wpengine.net
```

If this fails, check that the SSH key registered in the WPEngine portal matches the private key path you provided during `stax setup`.

---

### "phpcs not found" (stax migrate audit)

Install phpcs with the WordPress VIP coding standards:

```bash
composer global require automattic/vip-coding-standards
```

Verify phpcs is on your PATH:

```bash
phpcs -i
```

---

### "VIP CLI not found" (stax migrate import)

Install the VIP CLI:

```bash
npm install -g @automattic/vip
```

---

### "migration.destination is not set"

The `stax migrate` commands require the destination to be configured. Add the following to `.stax.yml`:

```yaml
migration:
  destination: vip
```

Alternatively, pass the flag for a one-off run:

```bash
stax migrate export --destination=vip
```

---

### Zero URL replacements after db pull

The search-replace ran without errors but the site URL was not updated. This usually means the `provider_config.install` value in `.stax.yml` doesn't match the actual WPEngine install name, so the pattern-based source URL is wrong.

Check:

```bash
stax doctor
```

The doctor command validates that the install name resolves to a real WPEngine site and reports if there is a mismatch.

---

### "nginx error" after stax media setup

The generated nginx config is missing a `server {}` wrapper. Run setup again:

```bash
stax media setup
```

Then inspect the generated file to confirm it has the correct structure:

```bash
cat .ddev/nginx_full/media-proxy.conf
```

The file should contain a `server { ... }` block wrapping all directives. If it doesn't, re-run `stax media setup` — the command regenerates the file from scratch.
