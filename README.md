# Cloudflare DNS CLI

A command-line tool for managing Cloudflare DNS records, built with Go using the Cobra framework.

## Installation

### Quick Install (recommended)

```bash
# Global install (requires sudo)
curl -fsSL https://raw.githubusercontent.com/coollabsio/cloudflare-cli/main/scripts/install.sh | bash

# User install (no sudo required)
curl -fsSL https://raw.githubusercontent.com/coollabsio/cloudflare-cli/main/scripts/install.sh | bash -s -- --user
```

### Using `go install`

```bash
go install github.com/coollabsio/cloudflare-cli@latest
```

This will install the `cf` binary in your `$GOPATH/bin` directory (usually `~/go/bin`). Make sure this directory is in your `$PATH`.

### Build from source

```bash
git clone https://github.com/coollabsio/cloudflare-cli.git
cd cloudflare-cli
go build -o cf .
```

## Getting Started

### 1. Get your Cloudflare API Token

Create an API token in the [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
- **Zone:Read** - For listing zones and zone details
- **DNS:Read** - For listing/viewing DNS records
- **DNS:Edit** - For creating/updating/deleting records

### 2. Configure Authentication

#### Option A: Save token to config file (recommended)

```bash
cf auth save <your-api-token>
```

This saves the token to `~/.cloudflare/config.yaml`.

#### Option B: Use environment variables

```bash
export CLOUDFLARE_API_TOKEN=your-api-token
# or
export CF_API_TOKEN=your-api-token
```

#### Option C: Use API Key + Email (legacy)

```bash
export CLOUDFLARE_API_KEY=your-api-key
export CLOUDFLARE_API_EMAIL=your-email
```

### 3. Verify authentication

```bash
cf auth verify
```

## Currently Supported Commands

### Authentication
- `cf auth verify` - Verify API credentials
- `cf auth save <token>` - Save API token to config file

### Configuration
- `cf config set <key> <value>` - Set a config value
- `cf config get <key>` - Get a config value
- `cf config list` - List all config values

Available config keys:
- `output_format` - Default output format (`table` or `json`)

### Zone Management
- `cf zones list` - List all zones
- `cf zones get <zone-name-or-id>` - Get zone details

### DNS Record Management
- `cf dns list <zone>` - List DNS records
  - `--type, -t` - Filter by record type (A, AAAA, CNAME, TXT, MX, etc.)
  - `--name, -n` - Filter by record name
  - `--search, -s` - Search in name, content, and comment (case-insensitive)
- `cf dns get <zone> <record-id>` - Get DNS record details
- `cf dns create <zone>` - Create a DNS record
  - `--type, -t` - Record type (required)
  - `--name, -n` - Record name (required)
  - `--content, -c` - Record content (required)
  - `--ttl` - TTL in seconds (1 = auto, default: 1)
  - `--proxied` - Proxy through Cloudflare (true|false)
  - `--priority` - Record priority (for MX, SRV)
  - `--comment` - Comment for the record
- `cf dns update <zone> <record-id>` - Update a DNS record
  - Only specify fields you want to change
  - `--type, -t` - New record type
  - `--name, -n` - New record name
  - `--content, -c` - New record content
  - `--ttl` - TTL in seconds
  - `--proxied` - Set proxy status (true|false)
  - `--priority` - Record priority
  - `--comment` - Comment for the record (use empty string to clear)
- `cf dns delete <zone> <record-id>` - Delete a DNS record
- `cf dns find <zone>` - Find DNS records by name and/or type
  - `--type, -t` - Record type to find
  - `--name, -n` - Record name to find

## Global Flags

All commands support these global flags:

- `--config` - Config file path (default: `~/.cloudflare/config.yaml`)
- `--output, -o` - Output format: `table` (default) or `json`

## Examples

### Zone Operations

```bash
# List all zones
cf zones list

# Get zone details
cf zones get example.com

# Get zone by ID (useful for zone-specific tokens)
cf zones get 023e105f4ecef8ad9ca31a8372d0c353
```

### DNS Record Operations

```bash
# List all DNS records for a zone
cf dns list example.com

# List only A records
cf dns list example.com --type A

# List records matching a name
cf dns list example.com --name www

# Search records by name, content, or comment
cf dns list example.com --search "production"

# Create an A record
cf dns create example.com --name www --type A --content 192.0.2.1

# Create a proxied CNAME record
cf dns create example.com --name blog --type CNAME --content example.com --proxied

# Create an MX record with priority
cf dns create example.com --name mail --type MX --content mail.example.com --priority 10

# Create a record with a comment
cf dns create example.com --name api --type A --content 192.0.2.10 --comment "Production API server"

# Update only the content of a record
cf dns update example.com abc123def456 --content 192.0.2.2

# Enable proxying on an existing record
cf dns update example.com abc123def456 --proxied

# Disable proxying
cf dns update example.com abc123def456 --proxied=false

# Update the comment on a record
cf dns update example.com abc123def456 --comment "Updated comment"

# Clear the comment on a record
cf dns update example.com abc123def456 --comment ""

# Delete a record
cf dns delete example.com abc123def456

# Find record ID by name and type
cf dns find example.com --name www --type A
```

### JSON Output

```bash
# Get JSON output for scripting
cf dns list example.com --output json

# Set JSON as default output format
cf config set output_format json
```

## Permission Quirk (Important!)

If you create an API token scoped to specific zones (not "All zones"), you **cannot** list zones or look up zone IDs by name. The API returns a permission error.

**Workarounds:**

1. Use the zone ID directly instead of the zone name:
   ```bash
   cf dns list 023e105f4ecef8ad9ca31a8372d0c353
   ```

2. Grant your token "All zones" read permission for zone listing

The CLI accepts both zone names and zone IDs for all commands.

## Configuration File

The config file is stored at `~/.cloudflare/config.yaml`:

```yaml
api_token: your-api-token-here
output_format: table
```

Environment variables take precedence over config file values.

## Development

```bash
# Build
go build -o cf .

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Install locally
go install .
```

## Architecture

```
cf/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # CLI setup, global flags
│   ├── auth.go            # auth verify/save commands
│   ├── config.go          # config set/get/list commands
│   ├── zones.go           # zones list/get commands
│   └── dns.go             # dns list/get/create/update/delete/find commands
├── internal/
│   ├── client/
│   │   └── client.go      # Cloudflare API client wrapper
│   ├── config/
│   │   └── config.go      # Configuration management
│   └── output/
│       └── output.go      # Table/JSON output formatting
├── go.mod
└── go.sum
```

## Dependencies

- [cloudflare-go](https://github.com/cloudflare/cloudflare-go) - Official Cloudflare Go library
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration

## Roadmap

This CLI currently focuses on DNS management, but additional Cloudflare API features are planned:

- **DNS** - Record management (current)
- **Zones** - Zone settings, purge cache, development mode
- **SSL/TLS** - Certificate management, SSL settings
- **Firewall** - WAF rules, IP access rules, rate limiting
- **Page Rules** - URL-based settings
- **Workers** - Deploy and manage Workers scripts
- **R2** - Object storage management
- **D1** - Database management
- **KV** - Key-Value storage
- **Pages** - Cloudflare Pages deployments
- **Load Balancing** - Pool and monitor management
- **Access** - Zero Trust access policies

Contributions welcome!

## License

MIT
