All stax command groups with one-line descriptions.

| Command | Description |
|---------|-------------|
| `stax setup` | Configure WPEngine and GitHub credentials |
| `stax init` | Initialize a new stax project and create `.stax.yml` |
| `stax start` | Start the DDEV environment |
| `stax stop` | Stop the DDEV environment |
| `stax restart` | Restart the DDEV environment |
| `stax status` | Show DDEV and project status |
| `stax dev` | Start development mode (DDEV + file watcher) |
| `stax build` | Build the project (Composer, npm) |
| `stax db` | Database pull, push, and snapshot operations |
| `stax files` | Sync wp-content files via rsync |
| `stax media` | Configure the nginx remote media proxy |
| `stax migrate` | Orchestrate WPEngine → VIP migration |
| `stax config` | Read, validate, and manage `.stax.yml` |
| `stax validate` | Validate project configuration |
| `stax doctor` | Check prerequisites and auto-fix common issues |
| `stax repo` | Git repository operations |
| `stax actions` | Set up GitHub Actions workflows |
| `stax lint` | Run PHP CodeSniffer checks |
| `stax list` | List available WPEngine installs |
| `stax wpengine` | Global WPEngine discovery and management |
| `stax version` | Print the stax version and feature status |
| `stax man` | Generate man pages |
| `stax completion` | Generate shell completion scripts |

## Detailed reference

- [stax db](db.md)
- [stax files](files.md)
- [stax migrate](migrate.md)
- [stax config](config.md)

## Global flags

These flags are available on every command:

| Flag | Description |
|------|-------------|
| `-c, --config string` | Config file path (default: `.stax.yml`) |
| `-d, --debug` | Enable debug logging |
| `--no-color` | Disable colored output |
| `--project-dir string` | Project directory (default: current directory) |
| `-q, --quiet` | Suppress non-error output |
| `-v, --verbose` | Enable verbose output |
