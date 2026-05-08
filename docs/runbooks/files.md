Sync wp-content files between WPEngine and your local environment using rsync over SSH.

## Pull all wp-content

```bash
stax files pull
```

Pulls everything under `wp-content/` except uploads (excluded by default because the uploads directory is often large).

## Pull specific directories

```bash
stax files pull --themes-only
stax files pull --plugins-only
stax files pull --mu-plugins-only
```

Restricts the sync to a single subdirectory. Useful when you only need to update one area and don't want to wait for a full sync.

## Dry run

```bash
stax files pull --dry-run
```

Shows what would be transferred without making changes. Use this before any pull that could overwrite local modifications.

## Exclude uploads

```bash
stax files pull --exclude-uploads
```

Explicitly excludes `wp-content/uploads/`. Uploads are typically large and can be served directly from WPEngine using the media proxy — see the [media proxy runbook](media-proxy.md).

## Delete local files not on remote

```bash
stax files pull --delete
```

Removes any local files under the synced directory that do not exist on the remote. Use carefully — this is destructive and cannot be undone.

## Bandwidth throttle

```bash
stax files pull --bandwidth-limit=500
```

Caps rsync bandwidth at 500 KB/s. Useful on slow or shared connections to avoid saturating the network.

## Push files

Always review before pushing:

```bash
stax files push --dry-run
```

Then push to staging:

```bash
stax files push --environment=staging
```

A push to production requires interactive confirmation at the prompt.

## Exclude patterns

```bash
stax files pull --exclude="*.log,cache/"
```

Comma-separated list of patterns to exclude from the sync. Patterns follow rsync syntax — trailing `/` matches directories, `*` matches within a path component.

## What not to do

- Never run `stax files push` to production without reviewing `--dry-run` output first. Files on the remote are overwritten.
- Do not push the `uploads/` directory without a deliberate decision. It can be extremely large and pushing it to production can overwrite user-uploaded content.
