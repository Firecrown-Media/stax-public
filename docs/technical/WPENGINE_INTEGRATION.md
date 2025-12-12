# WPEngine Integration Strategy

## Overview

Stax integrates deeply with WPEngine to provide seamless database and file synchronization, environment matching, and deployment workflows. This document details the technical implementation of WPEngine integration, including API authentication, SSH Gateway access, database operations, file syncing, and remote media proxying.

## Authentication Architecture

### Multi-Factor Authentication

Stax uses two authentication methods for WPEngine:

1. **API Authentication**: Username/password for WPEngine API
2. **SSH Authentication**: SSH key for WPEngine SSH Gateway

Both credentials are stored securely in macOS Keychain (see CONFIG_SPEC.md for details).

### API Authentication

**Endpoint**: `https://api.wpengineapi.com/v1`

**Authentication Method**: HTTP Basic Authentication

**Headers**:
```http
Authorization: Basic <base64(username:password)>
Content-Type: application/json
```

**Credential Storage**:
```json
{
  "api_user": "myuser@example.com",
  "api_password": "mypassword"
}
```

**Go Implementation**:
```go
// pkg/wpengine/client.go
type Client struct {
    baseURL    string
    httpClient *http.Client
    username   string
    password   string
}

func NewClient(username, password string) *Client {
    return &Client{
        baseURL: "https://api.wpengineapi.com/v1",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        username: username,
        password: password,
    }
}

func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    var buf io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        buf = bytes.NewBuffer(data)
    }

    req, err := http.NewRequest(method, c.baseURL+path, buf)
    if err != nil {
        return nil, err
    }

    req.SetBasicAuth(c.username, c.password)
    req.Header.Set("Content-Type", "application/json")

    return c.httpClient.Do(req)
}
```

### SSH Gateway Authentication

**Hostname**: `ssh.wpengine.net`

**Port**: 22

**Authentication Method**: SSH public key authentication

**SSH User Format**: `<install_name>@<install_name>`

**Example**: `fsmultisite@fsmultisite`

**SSH Config**:
```
Host wpengine-*
    HostName ssh.wpengine.net
    Port 22
    User %h
    IdentityFile ~/.ssh/wpengine_rsa
    StrictHostKeyChecking accept-new
```

**Go Implementation**:
```go
// pkg/wpengine/ssh.go
import (
    "golang.org/x/crypto/ssh"
    "github.com/firecrown-media/stax/pkg/security"
)

func NewSSHClient(config SSHConfig) (*SSHClient, error) {
    signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
    if err != nil {
        return nil, err
    }

    // WPEngine uses install-specific SSH gateways: {install}.ssh.wpengine.net
    if config.Host == "" {
        config.Host = fmt.Sprintf("%s.ssh.wpengine.net", config.Install)
    }

    // Initialize known hosts manager for secure host key verification
    // Uses TOFU (Trust On First Use) pattern with ~/.stax/known_hosts
    khManager, err := security.NewKnownHostsManager()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize known hosts manager: %w", err)
    }

    sshConfig := &ssh.ClientConfig{
        User: config.Install,  // SSH user is just the install name
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: khManager.GetHostKeyCallback(),  // Secure TOFU verification
        Timeout:         30 * time.Second,
    }

    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", config.Host), sshConfig)
    if err != nil {
        return nil, err
    }

    return &SSHClient{client: client, config: config}, nil
}

func (c *SSHClient) ExecuteCommand(cmd string) (string, error) {
    session, err := c.client.NewSession()
    if err != nil {
        return "", err
    }
    defer session.Close()

    output, err := session.CombinedOutput(cmd)
    return string(output), err
}
```

## WPEngine API Endpoints

### 1. List Installs

**Endpoint**: `GET /installs`

**Purpose**: List all WPEngine installations for the account

**Response**:
```json
{
  "results": [
    {
      "id": "abc123",
      "name": "fsmultisite",
      "primary_domain": "fsmultisite.wpengine.com",
      "php_version": "8.2",
      "environment": "production"
    }
  ]
}
```

**Go Implementation**:
```go
type Install struct {
    ID            string `json:"id"`
    Name          string `json:"name"`
    PrimaryDomain string `json:"primary_domain"`
    PHPVersion    string `json:"php_version"`
    Environment   string `json:"environment"`
}

func (c *Client) ListInstalls() ([]Install, error) {
    resp, err := c.makeRequest("GET", "/installs", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Results []Install `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Results, nil
}
```

### 2. Get Install Details

**Endpoint**: `GET /installs/{install_id}`

**Purpose**: Get detailed information about a specific installation

**Response**:
```json
{
  "id": "abc123",
  "name": "fsmultisite",
  "primary_domain": "fsmultisite.wpengine.com",
  "php_version": "8.2",
  "mysql_version": "8.0",
  "wordpress_version": "6.4.2",
  "environment": "production",
  "disk_usage": {
    "used": 2500000000,
    "total": 10000000000
  },
  "domains": [
    "fsmultisite.wpengine.com",
    "firecrown.com",
    "flyingmag.com",
    "planeandpilotmag.com"
  ]
}
```

**Go Implementation**:
```go
type InstallDetails struct {
    ID               string   `json:"id"`
    Name             string   `json:"name"`
    PrimaryDomain    string   `json:"primary_domain"`
    PHPVersion       string   `json:"php_version"`
    MySQLVersion     string   `json:"mysql_version"`
    WordPressVersion string   `json:"wordpress_version"`
    Environment      string   `json:"environment"`
    DiskUsage        struct {
        Used  int64 `json:"used"`
        Total int64 `json:"total"`
    } `json:"disk_usage"`
    Domains []string `json:"domains"`
}

func (c *Client) GetInstallDetails(installID string) (*InstallDetails, error) {
    resp, err := c.makeRequest("GET", fmt.Sprintf("/installs/%s", installID), nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var details InstallDetails
    if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
        return nil, err
    }

    return &details, nil
}
```

### 3. List Backups

**Endpoint**: `GET /installs/{install_id}/backups`

**Purpose**: List available database backups

**Response**:
```json
{
  "results": [
    {
      "id": "backup123",
      "type": "automatic",
      "created_at": "2025-11-08T14:30:00Z",
      "size": 256000000,
      "status": "complete"
    },
    {
      "id": "backup124",
      "type": "manual",
      "created_at": "2025-11-07T09:15:00Z",
      "size": 255000000,
      "status": "complete"
    }
  ]
}
```

**Go Implementation**:
```go
type Backup struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    CreatedAt time.Time `json:"created_at"`
    Size      int64     `json:"size"`
    Status    string    `json:"status"`
}

func (c *Client) ListBackups(installID string) ([]Backup, error) {
    resp, err := c.makeRequest("GET", fmt.Sprintf("/installs/%s/backups", installID), nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Results []Backup `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Results, nil
}
```

### 4. Create Backup

**Endpoint**: `POST /installs/{install_id}/backups`

**Purpose**: Trigger a manual backup

**Request**:
```json
{
  "description": "Backup before stax db:push"
}
```

**Response**:
```json
{
  "id": "backup125",
  "status": "pending"
}
```

**Go Implementation**:
```go
func (c *Client) CreateBackup(installID, description string) (string, error) {
    body := map[string]string{
        "description": description,
    }

    resp, err := c.makeRequest("POST", fmt.Sprintf("/installs/%s/backups", installID), body)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        ID string `json:"id"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    return result.ID, nil
}
```

## SSH Gateway Operations

WPEngine's SSH Gateway provides direct access to the WordPress installation filesystem and MySQL database.

### SSH Gateway Capabilities

1. **File Access**: Full access to installation directory
2. **Database Access**: Via `wp db` commands
3. **WP-CLI**: Full WP-CLI command suite
4. **Rsync**: File synchronization

### Connection Pattern

**Connection String**: `ssh <install>@<install>@ssh.wpengine.net`

**Example**: `ssh fsmultisite@fsmultisite@ssh.wpengine.net`

**Base Directory**: `/sites/<install>/`

**Key Directories**:
- `/sites/<install>/wp-content/` - WordPress content
- `/sites/<install>/wp-content/uploads/` - Media uploads
- `/sites/<install>/wp-config.php` - WordPress configuration

### File Operations via SSH

**List files**:
```bash
ssh fsmultisite@fsmultisite@ssh.wpengine.net "ls -la /sites/fsmultisite/wp-content/"
```

**Download single file**:
```bash
scp fsmultisite@fsmultisite@ssh.wpengine.net:/sites/fsmultisite/wp-config.php ./
```

**Rsync directory**:
```bash
rsync -avz --progress \
  fsmultisite@fsmultisite@ssh.wpengine.net:/sites/fsmultisite/wp-content/uploads/ \
  ./wp-content/uploads/
```

**Go Implementation**:
```go
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
    session, err := c.client.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()

    // Create local file
    localFile, err := os.Create(localPath)
    if err != nil {
        return err
    }
    defer localFile.Close()

    // Setup remote command output to local file
    session.Stdout = localFile

    // Execute cat command
    cmd := fmt.Sprintf("cat %s", remotePath)
    return session.Run(cmd)
}
```

### Database Operations via SSH Gateway

WPEngine provides WP-CLI access via SSH, enabling direct database operations.

**Query database**:
```bash
ssh fsmultisite@fsmultisite@ssh.wpengine.net \
  "wp db query 'SELECT option_value FROM wp_options WHERE option_name = \"siteurl\"'"
```

**Export database**:
```bash
ssh fsmultisite@fsmultisite@ssh.wpengine.net \
  "wp db export --add-drop-table -" > database.sql
```

**Export with exclusions**:
```bash
ssh fsmultisite@fsmultisite@ssh.wpengine.net \
  "wp db export --add-drop-table --exclude_tables=wp_actionscheduler_logs -" > database.sql
```

**Go Implementation**:
```go
func (c *SSHClient) ExportDatabase(excludeTables []string) (io.Reader, error) {
    cmd := "wp db export --add-drop-table"

    if len(excludeTables) > 0 {
        cmd += fmt.Sprintf(" --exclude_tables=%s", strings.Join(excludeTables, ","))
    }

    cmd += " -"  // Output to stdout

    session, err := c.client.NewSession()
    if err != nil {
        return nil, err
    }

    stdout, err := session.StdoutPipe()
    if err != nil {
        return nil, err
    }

    if err := session.Start(cmd); err != nil {
        return nil, err
    }

    return stdout, nil
}
```

## Database Download Strategy

### Partial Export Optimization

WPEngine databases can be large (1GB+). Stax optimizes download time by excluding unnecessary tables.

**Default Exclusions**:
```go
var DefaultExcludedTables = []string{
    // Action Scheduler logs (large, regenerates)
    "wp_actionscheduler_logs",
    "wp_actionscheduler_actions",

    // Transients (temporary data)
    "wp_options WHERE option_name LIKE '_transient_%'",

    // Spam and trash (not needed locally)
    "wp_comments WHERE comment_approved = 'spam'",
    "wp_comments WHERE comment_approved = 'trash'",

    // Admin notes (large, not critical)
    "wp_wc_admin_notes",
    "wp_wc_admin_note_actions",

    // Search index (rebuilds locally)
    "wp_relevanssi",
}
```

**Table Prefix Detection**:
```go
func (c *SSHClient) DetectTablePrefix() (string, error) {
    cmd := `wp db query "SHOW TABLES LIKE '%_options'" --skip-column-names`

    output, err := c.ExecuteCommand(cmd)
    if err != nil {
        return "", err
    }

    // Extract prefix from first table
    // Example output: "wp_options" → prefix is "wp_"
    tableName := strings.TrimSpace(output)
    if !strings.Contains(tableName, "_options") {
        return "", errors.New("could not detect table prefix")
    }

    prefix := strings.TrimSuffix(tableName, "options")
    return prefix, nil
}
```

**Conditional Export**:
```go
func (c *SSHClient) ExportDatabaseOptimized(options ExportOptions) (io.Reader, error) {
    prefix, err := c.DetectTablePrefix()
    if err != nil {
        return nil, err
    }

    var excludedTables []string

    // Skip logs
    if options.SkipLogs {
        excludedTables = append(excludedTables,
            prefix+"actionscheduler_logs",
            prefix+"actionscheduler_actions",
        )
    }

    // Skip transients (requires custom SQL)
    if options.SkipTransients {
        // Note: Transients must be handled post-export
    }

    // Skip spam
    if options.SkipSpam {
        // Note: Spam filtering requires post-processing
    }

    // Add user-specified exclusions
    for _, table := range options.ExcludeTables {
        excludedTables = append(excludedTables, prefix+table)
    }

    return c.ExportDatabase(excludedTables)
}
```

### Database Import Process

**Import Flow**:
```
1. Download from WPEngine SSH Gateway
2. Stream to temporary file
3. Import to DDEV MySQL container
4. Run search-replace operations
5. Flush WordPress cache
```

**Go Implementation**:
```go
func ImportDatabase(reader io.Reader, ddevProject string) error {
    // Create temporary file for SQL
    tmpFile, err := ioutil.TempFile("", "stax-import-*.sql")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name())

    // Write database dump to temp file
    _, err = io.Copy(tmpFile, reader)
    if err != nil {
        return err
    }
    tmpFile.Close()

    // Import via DDEV
    cmd := exec.Command("ddev", "import-db", "--src="+tmpFile.Name())
    cmd.Dir = ddevProject

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("import failed: %s", output)
    }

    return nil
}
```

### Search-Replace Strategy

After database import, Stax performs multi-step search-replace to update URLs.

**Search-Replace Sequence**:

1. **Network URL** (affects all sites)
   ```bash
   wp search-replace fsmultisite.wpenginepowered.com firecrown.local --network
   ```

2. **Individual Site URLs** (per-site)
   ```bash
   wp search-replace flyingmag.com flyingmag.firecrown.local --url=flyingmag.com
   wp search-replace planeandpilotmag.com planeandpilot.firecrown.local --url=planeandpilotmag.com
   wp search-replace finescale.com finescale.firecrown.local --url=finescale.com
   wp search-replace avweb.com avweb.firecrown.local --url=avweb.com
   ```

3. **Protocol Updates** (HTTPS → HTTPS, but local cert)
   ```bash
   wp search-replace https://flyingmag.com https://flyingmag.firecrown.local --url=flyingmag.com
   ```

**Column Exclusions**:
```go
var SearchReplaceSkipColumns = []string{
    "guid",  // WordPress post GUIDs should never change
}
```

**Go Implementation**:
```go
type SearchReplaceOperation struct {
    Old string
    New string
    URL string  // For multisite, specify which site
}

func PerformSearchReplace(operations []SearchReplaceOperation, ddevProject string) error {
    for _, op := range operations {
        args := []string{"wp", "search-replace", op.Old, op.New}

        if op.URL != "" {
            args = append(args, "--url="+op.URL)
        }

        // Skip GUID column
        args = append(args, "--skip-columns=guid")

        // Dry run first
        dryRunArgs := append(args, "--dry-run")
        cmd := exec.Command("ddev", dryRunArgs...)
        cmd.Dir = ddevProject

        output, err := cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("dry run failed: %s", output)
        }

        // Parse dry run output to show replacements
        fmt.Printf("Will replace %s → %s\n", op.Old, op.New)

        // Execute actual replacement
        cmd = exec.Command("ddev", args...)
        cmd.Dir = ddevProject

        output, err = cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("search-replace failed: %s", output)
        }
    }

    return nil
}
```

## File Sync Strategy

### Rsync-Based Synchronization

Stax uses rsync over SSH for efficient file synchronization.

**Rsync Command Template**:
```bash
rsync -rlDvz \
  --size-only \
  --progress \
  --exclude="*.log" \
  --exclude="cache/" \
  <install>@<install>@ssh.wpengine.net:/sites/<install>/wp-content/uploads/ \
  ./wp-content/uploads/
```

**Rsync Flags Explained**:
- `-r`: Recursive
- `-l`: Copy symlinks as symlinks
- `-D`: Preserve device files and special files
- `-v`: Verbose
- `-z`: Compress during transfer
- `--size-only`: Skip files of same size (faster than checksum)
- `--progress`: Show progress during transfer

**Go Implementation**:
```go
type RsyncOptions struct {
    Source      string
    Destination string
    Exclude     []string
    Delete      bool  // Delete files not on remote
    DryRun      bool
    BandwidthLimit int  // KB/s
}

func (c *SSHClient) Rsync(options RsyncOptions) error {
    args := []string{
        "-rlDvz",
        "--size-only",
        "--progress",
    }

    // Add exclusions
    for _, pattern := range options.Exclude {
        args = append(args, "--exclude="+pattern)
    }

    // Delete local files not on remote
    if options.Delete {
        args = append(args, "--delete")
    }

    // Bandwidth limit
    if options.BandwidthLimit > 0 {
        args = append(args, fmt.Sprintf("--bwlimit=%d", options.BandwidthLimit))
    }

    // Dry run
    if options.DryRun {
        args = append(args, "--dry-run")
    }

    // Source and destination
    args = append(args, options.Source, options.Destination)

    cmd := exec.Command("rsync", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}
```

**Default Exclusions**:
```go
var DefaultRsyncExclusions = []string{
    "*.log",
    "cache/",
    ".DS_Store",
    "Thumbs.db",
    "*.tmp",
    "*.swp",
}
```

### Selective File Sync

Sync only specific directories:

```bash
# Sync uploads only
stax wpe:sync wp-content/uploads

# Sync plugins
stax wpe:sync wp-content/plugins

# Sync specific theme
stax wpe:sync wp-content/themes/my-theme
```

### Incremental Sync

For large uploads directories, use incremental sync:

```go
func (c *SSHClient) IncrementalSync(remotePath, localPath string) error {
    // First sync: Get file list
    listCmd := fmt.Sprintf("find %s -type f -printf '%%P\\n'", remotePath)
    remoteFiles, err := c.ExecuteCommand(listCmd)
    if err != nil {
        return err
    }

    // Compare with local files
    localFiles := getLocalFileList(localPath)

    // Determine files to sync
    filesToSync := diffFileLists(remoteFiles, localFiles)

    // Sync only changed files
    for _, file := range filesToSync {
        source := fmt.Sprintf("%s@%s@ssh.wpengine.net:%s/%s",
            c.install, c.install, remotePath, file)
        dest := filepath.Join(localPath, file)

        if err := c.Rsync(RsyncOptions{
            Source:      source,
            Destination: dest,
        }); err != nil {
            return err
        }
    }

    return nil
}
```

## Remote Media Proxying

Instead of downloading all media files locally, Stax configures Nginx to proxy media requests to remote sources (BunnyCDN or WPEngine).

### Nginx Configuration

**DDEV Nginx Config** (`.ddev/nginx-site.conf`):

```nginx
# Proxy wp-content/uploads requests to remote sources
location ~ ^/wp-content/uploads/(.*)$ {
    # Try local file first
    try_files $uri @proxy_media;
}

location @proxy_media {
    # Set variables
    set $upstream_bunnycdn https://cdn.firecrown.com;
    set $upstream_wpengine https://fsmultisite.wpengine.com;

    # Try BunnyCDN first
    proxy_pass $upstream_bunnycdn$request_uri;
    proxy_ssl_server_name on;
    proxy_intercept_errors on;
    recursive_error_pages on;

    # If BunnyCDN returns 404, try WPEngine
    error_page 404 = @proxy_wpengine;

    # Caching
    proxy_cache media_cache;
    proxy_cache_valid 200 24h;
    proxy_cache_valid 404 1m;
    proxy_cache_key "$scheme$request_method$host$request_uri";

    # Headers
    proxy_set_header Host cdn.firecrown.com;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

    # Hide upstream headers
    proxy_hide_header Set-Cookie;
    proxy_ignore_headers Set-Cookie;
    add_header X-Proxy-Cache $upstream_cache_status;
}

location @proxy_wpengine {
    proxy_pass $upstream_wpengine$request_uri;
    proxy_ssl_server_name on;

    # Caching
    proxy_cache media_cache;
    proxy_cache_valid 200 24h;
    proxy_cache_key "$scheme$request_method$host$request_uri";

    # Headers
    proxy_set_header Host fsmultisite.wpengine.com;
    proxy_hide_header Set-Cookie;
    proxy_ignore_headers Set-Cookie;
    add_header X-Proxy-Cache $upstream_cache_status;
}

# Cache configuration
proxy_cache_path /var/cache/nginx/media
    levels=1:2
    keys_zone=media_cache:10m
    max_size=1g
    inactive=24h;
```

**DDEV Config Generation**:

```go
type MediaProxyConfig struct {
    Enabled         bool
    BunnyCDNHost    string
    WPEngineHost    string
    CacheEnabled    bool
    CacheMaxSize    string
    CacheTTL        string
}

func GenerateNginxMediaProxy(config MediaProxyConfig) (string, error) {
    tmpl := template.Must(template.ParseFiles("templates/ddev/nginx-site.conf.tmpl"))

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, config); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

**Template** (`templates/ddev/nginx-site.conf.tmpl`):

```nginx
{{if .Enabled}}
# Remote media proxy configuration
location ~ ^/wp-content/uploads/(.*)$ {
    try_files $uri @proxy_media;
}

location @proxy_media {
    set $upstream_bunnycdn {{.BunnyCDNHost}};
    set $upstream_wpengine {{.WPEngineHost}};

    proxy_pass $upstream_bunnycdn$request_uri;
    proxy_ssl_server_name on;
    proxy_intercept_errors on;
    recursive_error_pages on;
    error_page 404 = @proxy_wpengine;

    {{if .CacheEnabled}}
    proxy_cache media_cache;
    proxy_cache_valid 200 {{.CacheTTL}};
    proxy_cache_valid 404 1m;
    add_header X-Proxy-Cache $upstream_cache_status;
    {{end}}
}

location @proxy_wpengine {
    proxy_pass $upstream_wpengine$request_uri;
    proxy_ssl_server_name on;

    {{if .CacheEnabled}}
    proxy_cache media_cache;
    proxy_cache_valid 200 {{.CacheTTL}};
    {{end}}
}

{{if .CacheEnabled}}
proxy_cache_path /var/cache/nginx/media
    levels=1:2
    keys_zone=media_cache:10m
    max_size={{.CacheMaxSize}}
    inactive={{.CacheTTL}};
{{end}}
{{end}}
```

### Media Proxy Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Browser Request                              │
│              GET /wp-content/uploads/2025/11/image.jpg          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Nginx (DDEV)                                 │
│  1. Check local file: wp-content/uploads/2025/11/image.jpg     │
│     - Found: Serve directly                                     │
│     - Not found: Proxy to remote                                │
└────────────────────────┬────────────────────────────────────────┘
                         │ (not found)
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Check Nginx Cache                            │
│  - Cache key: https://cdn.firecrown.com/.../image.jpg          │
│  - Cache hit: Serve from cache                                  │
│  - Cache miss: Fetch from upstream                              │
└────────────────────────┬────────────────────────────────────────┘
                         │ (cache miss)
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Proxy to BunnyCDN                            │
│  GET https://cdn.firecrown.com/wp-content/uploads/.../image.jpg │
│  - 200 OK: Cache and serve                                      │
│  - 404 Not Found: Try WPEngine                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │ (404)
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Proxy to WPEngine                            │
│  GET https://fsmultisite.wpengine.com/.../image.jpg            │
│  - 200 OK: Cache and serve                                      │
│  - 404 Not Found: Return 404 to browser                         │
└─────────────────────────────────────────────────────────────────┘
```

### Benefits of Media Proxying

1. **No Local Storage**: Saves disk space (WPEngine uploads can be 10GB+)
2. **Faster Setup**: No time downloading all media files
3. **Always Fresh**: Media is always current from production
4. **Bandwidth Efficient**: Only downloads media when accessed
5. **Optional Caching**: Cache frequently accessed files locally

### Cache Management

**Cache Statistics**:
```bash
# View cache stats (inside DDEV container)
stax ssh
du -sh /var/cache/nginx/media
find /var/cache/nginx/media -type f | wc -l
```

**Clear Cache**:
```bash
# Clear all media cache
stax ssh
rm -rf /var/cache/nginx/media/*

# Or via Stax command
stax media:clear-cache
```

**Go Implementation**:
```go
func ClearMediaCache(ddevProject string) error {
    cmd := exec.Command("ddev", "exec", "rm", "-rf", "/var/cache/nginx/media/*")
    cmd.Dir = ddevProject
    return cmd.Run()
}

func GetCacheStats(ddevProject string) (*CacheStats, error) {
    // Get cache size
    sizeCmd := exec.Command("ddev", "exec", "du", "-sb", "/var/cache/nginx/media")
    sizeCmd.Dir = ddevProject
    sizeOutput, err := sizeCmd.Output()
    if err != nil {
        return nil, err
    }

    // Get file count
    countCmd := exec.Command("ddev", "exec", "find", "/var/cache/nginx/media", "-type", "f")
    countCmd.Dir = ddevProject
    countOutput, err := countCmd.Output()
    if err != nil {
        return nil, err
    }

    files := strings.Split(string(countOutput), "\n")
    fileCount := len(files) - 1  // Subtract empty line

    // Parse size
    sizeParts := strings.Fields(string(sizeOutput))
    size, _ := strconv.ParseInt(sizeParts[0], 10, 64)

    return &CacheStats{
        SizeBytes: size,
        FileCount: fileCount,
    }, nil
}
```

## GitHub Workflow Integration

### Deployment Workflow

Stax integrates with existing GitHub Actions workflows for WPEngine deployments.

**GitHub Action** (`.github/workflows/deploy-wpengine.yml`):

```yaml
name: Deploy to WPEngine

on:
  push:
    branches:
      - main
      - staging
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy to'
        required: true
        type: choice
        options:
          - production
          - staging

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Setup PHP
        uses: shivammathur/setup-php@v2
        with:
          php-version: '8.1'

      - name: Install Composer dependencies
        run: composer install --no-dev --optimize-autoloader

      - name: Install NPM dependencies
        run: npm install --production

      - name: Build assets
        run: bash scripts/build.sh

      - name: Deploy to WPEngine
        uses: wpengine/github-action-wpe-site-deploy@v3
        with:
          WPE_SSHG_KEY_PRIVATE: ${{ secrets.WPE_SSHG_KEY_PRIVATE }}
          WPE_ENV: ${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}
          SRC_PATH: "."
          REMOTE_PATH: "wp-content/"
```

**Stax Integration**:

```go
// pkg/github/workflows.go
type WorkflowDispatchInput struct {
    Environment string `json:"environment"`
}

func (c *Client) TriggerDeployment(repo, workflow, environment string) (int64, error) {
    ctx := context.Background()

    inputs := map[string]interface{}{
        "environment": environment,
    }

    event := github.CreateWorkflowDispatchEventRequest{
        Ref:    "main",
        Inputs: inputs,
    }

    _, resp, err := c.client.Actions.CreateWorkflowDispatchEventByFileName(
        ctx, c.owner, repo, workflow, event)
    if err != nil {
        return 0, err
    }

    // Get run ID from response location header
    location := resp.Header.Get("Location")
    runID := extractRunIDFromLocation(location)

    return runID, nil
}

func (c *Client) WatchWorkflowRun(repo string, runID int64) error {
    ctx := context.Background()
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            run, _, err := c.client.Actions.GetWorkflowRunByID(ctx, c.owner, repo, runID)
            if err != nil {
                return err
            }

            fmt.Printf("Status: %s\n", run.GetStatus())

            if run.GetStatus() == "completed" {
                if run.GetConclusion() == "success" {
                    fmt.Println("✓ Deployment successful!")
                    return nil
                } else {
                    return fmt.Errorf("deployment failed: %s", run.GetConclusion())
                }
            }
        }
    }
}
```

**Stax Command**:
```bash
stax wpe:deploy --environment=staging --watch
```

## Error Handling and Retries

### Connection Errors

**Retry Logic**:
```go
type RetryConfig struct {
    MaxAttempts int
    Delay       time.Duration
    Backoff     float64  // Exponential backoff multiplier
}

func (c *Client) makeRequestWithRetry(method, path string, body interface{}, retryConfig RetryConfig) (*http.Response, error) {
    var lastErr error
    delay := retryConfig.Delay

    for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
        resp, err := c.makeRequest(method, path, body)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        lastErr = err
        if attempt < retryConfig.MaxAttempts {
            fmt.Printf("Attempt %d/%d failed, retrying in %v...\n",
                attempt, retryConfig.MaxAttempts, delay)
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * retryConfig.Backoff)
        }
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### API Rate Limiting

**Rate Limit Handling**:
```go
func (c *Client) handleRateLimit(resp *http.Response) error {
    if resp.StatusCode != 429 {
        return nil
    }

    // Check Retry-After header
    retryAfter := resp.Header.Get("Retry-After")
    if retryAfter == "" {
        return errors.New("rate limited but no Retry-After header")
    }

    seconds, err := strconv.Atoi(retryAfter)
    if err != nil {
        return err
    }

    fmt.Printf("Rate limited, waiting %d seconds...\n", seconds)
    time.Sleep(time.Duration(seconds) * time.Second)

    return nil
}
```

### SSH Connection Errors

**Connection Pool**:
```go
type SSHPool struct {
    clients []*ssh.Client
    mu      sync.Mutex
}

func NewSSHPool(size int, installName, privateKey string) (*SSHPool, error) {
    pool := &SSHPool{
        clients: make([]*ssh.Client, 0, size),
    }

    for i := 0; i < size; i++ {
        client, err := NewSSHClient(installName, privateKey)
        if err != nil {
            return nil, err
        }
        pool.clients = append(pool.clients, client)
    }

    return pool, nil
}

func (p *SSHPool) Get() *ssh.Client {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.clients) == 0 {
        return nil
    }

    client := p.clients[0]
    p.clients = p.clients[1:]
    return client
}

func (p *SSHPool) Put(client *ssh.Client) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.clients = append(p.clients, client)
}
```

## Performance Optimization

### Parallel Downloads

**Download Multiple Files**:
```go
func (c *SSHClient) DownloadFilesParallel(files []string, localDir string, maxWorkers int) error {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, maxWorkers)
    errors := make(chan error, len(files))

    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release

            localPath := filepath.Join(localDir, filepath.Base(f))
            if err := c.DownloadFile(f, localPath); err != nil {
                errors <- err
            }
        }(file)
    }

    wg.Wait()
    close(errors)

    // Collect errors
    var errs []error
    for err := range errors {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return fmt.Errorf("failed to download %d files", len(errs))
    }

    return nil
}
```

### Database Streaming

**Stream Database Import** (avoid temp files):
```go
func StreamDatabaseImport(sshClient *SSHClient, ddevProject string) error {
    // Export from WPEngine (returns io.Reader)
    reader, err := sshClient.ExportDatabase(nil)
    if err != nil {
        return err
    }

    // Import directly to DDEV (pipe)
    cmd := exec.Command("ddev", "import-db", "--src=-")
    cmd.Dir = ddevProject
    cmd.Stdin = reader

    return cmd.Run()
}
```

## Testing and Validation

### Connection Testing

```go
func TestWPEngineConnection(username, password, installName, privateKey string) error {
    // Test API connection
    client := NewClient(username, password)
    installs, err := client.ListInstalls()
    if err != nil {
        return fmt.Errorf("API connection failed: %w", err)
    }

    // Verify install exists
    found := false
    for _, install := range installs {
        if install.Name == installName {
            found = true
            break
        }
    }
    if !found {
        return fmt.Errorf("install '%s' not found", installName)
    }

    // Test SSH connection
    sshClient, err := NewSSHClient(installName, privateKey)
    if err != nil {
        return fmt.Errorf("SSH connection failed: %w", err)
    }
    defer sshClient.Close()

    // Test WP-CLI access
    output, err := sshClient.ExecuteCommand("wp core version")
    if err != nil {
        return fmt.Errorf("WP-CLI test failed: %w", err)
    }

    fmt.Printf("✓ WPEngine connection successful\n")
    fmt.Printf("✓ WordPress version: %s\n", strings.TrimSpace(output))

    return nil
}
```

## Summary

WPEngine integration provides:

- **Dual Authentication**: API for metadata, SSH for data
- **Optimized Database Downloads**: Partial exports, table exclusions
- **Efficient File Sync**: Rsync with intelligent exclusions
- **Remote Media Proxying**: No local storage required
- **Search-Replace Automation**: Multi-step URL updates
- **GitHub Workflow Integration**: Automated deployments
- **Error Handling**: Retries, rate limiting, connection pooling
- **Performance**: Parallel operations, streaming, caching

This comprehensive integration ensures Stax can seamlessly pull from and push to WPEngine environments while maintaining data integrity and optimal performance.
