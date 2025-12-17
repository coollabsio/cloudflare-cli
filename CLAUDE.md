# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI tool for managing Cloudflare DNS records, built with Go using the Cobra framework. The CLI allows users to manage zones and DNS records through the Cloudflare API.

### API Specification
This CLI is a client for the Cloudflare API using the official `cloudflare-go` library (v0.x branch).
- **Library**: github.com/cloudflare/cloudflare-go
- **Authentication**: API Token (recommended) or API Key + Email (legacy)
- **Documentation**: https://developers.cloudflare.com/api/

## Architecture

### Command Structure
The codebase follows Cobra's command pattern with a root command and subcommands:
- Entry point: `main.go` calls `cmd.Execute()`
- Root command: `cmd/root.go` - contains global flags, config loading, output format handling
- Subcommands: Each command group is in its own file in `cmd/`:
  - `auth.go` - authentication (verify, save token)
  - `config.go` - configuration management (set, get, list)
  - `zones.go` - zone management (list, get) + helper functions
  - `dns.go` - DNS record CRUD (list, get, create, update, delete, find)

### Configuration Management
- Config file location: `~/.cloudflare/config.yaml`
- Environment variables override config file:
  - `CLOUDFLARE_API_TOKEN` or `CF_API_TOKEN`
  - `CLOUDFLARE_API_KEY` or `CF_API_KEY`
  - `CLOUDFLARE_API_EMAIL` or `CF_API_EMAIL`
- Config struct in `internal/config/config.go`

### API Client
Core API wrapper in `internal/client/client.go`:
- Wraps `cloudflare-go` library
- Handles both API token and API key+email authentication
- Provides zone ID resolution (name or ID)
- DNS record CRUD operations
- Helpful error messages for permission issues

### Output Formatting
Output layer in `internal/output/output.go`:
- `FormatTable` - aligned table output (default)
- `FormatJSON` - JSON output for scripting
- Helper functions: `FormatTTL()`, `FormatBool()`

## Development Commands

### Build
```bash
go build -o cf .
```

### Run locally
```bash
go run . [command]
```

### Test a command
```bash
go run . auth verify
go run . dns list example.com --type A
```

### Install locally
```bash
go install .
```

### Run tests
```bash
go test ./...
go test -cover ./...
```

## Key Patterns

### Adding a New Command
1. Create or modify file in `cmd/` (e.g., `cmd/newfeature.go`)
2. Define command struct with `cobra.Command`
3. Implement `RunE` function with:
   - Create client: `c, err := client.New(cfg)`
   - Resolve zone if needed: `zoneID, err := resolveZone(c, ctx, args[0])`
   - Call API method
   - Format output (check `outputFormat` for JSON)
4. Register flags in `init()` function
5. Add command to parent: `parentCmd.AddCommand(yourCmd)`

### Zone Resolution
The CLI accepts both zone names and zone IDs. Use `resolveZone()` helper:
```go
zoneID, err := resolveZone(c, ctx, args[0])
if err != nil {
    return err
}
```

### Output Format Handling
```go
if outputFormat == "json" {
    return out.WriteJSON(data)
}
// Table output
headers := []string{"ID", "Name", "Value"}
rows := [][]string{{...}}
return out.WriteTable(headers, rows)
```

### Flag with Optional Value (NoOptDefVal)
For boolean-like flags that should default to "true" when specified without value:
```go
cmd.Flags().StringVar(&flagVar, "flag", "", "description")
cmd.Flags().Lookup("flag").NoOptDefVal = "true"
```
This allows: `--flag` (means true), `--flag=false` (means false)

## Permission Quirk

**Important**: Zone-specific API tokens cannot list zones by name. The `zones:list` permission is required.

The client handles this by:
1. First checking if input looks like a zone ID (32 hex chars)
2. If not, attempting to look up by name
3. Providing helpful error message suggesting zone ID usage

## Future Improvements

### High Priority
- [ ] Add `--zone-id` flag as alternative to positional zone argument
- [ ] Add confirmation prompt for destructive operations (delete)
- [ ] Add `--dry-run` flag for create/update/delete operations
- [ ] Batch operations (create/update/delete multiple records)
- [ ] Import/export DNS records (JSON/BIND format)

### Medium Priority
- [ ] Shell completion (bash, zsh, fish)
- [ ] Self-update command
- [ ] Support for more DNS record types (SRV, CAA, CERT, etc.)
- [ ] Colored output for terminal
- [ ] Pagination for large record sets
- [ ] Record comparison/diff before update
- [ ] Template support for common record patterns

### Low Priority
- [ ] Interactive mode / TUI
- [ ] Multiple account/context support (like coolify-cli)
- [ ] DNS record history/audit log viewing
- [ ] DNSSEC management
- [ ] Zone transfer/clone between accounts
- [ ] Webhook notifications on changes
- [ ] Rate limit handling with backoff

### Code Quality
- [ ] Add comprehensive unit tests
- [ ] Add integration tests with mock server
- [ ] Add golangci-lint configuration
- [ ] Add GitHub Actions CI/CD
- [ ] Add GoReleaser for multi-platform builds
- [ ] Add install script

### Documentation
- [ ] Add man page generation
- [ ] Add usage examples for common workflows
- [ ] Add troubleshooting guide
- [ ] Add contributing guide

## Testing Requirements

When adding new features:
1. Test command parsing and flag handling
2. Test output formatting (table, json)
3. Test error handling (API errors, validation)
4. Use mock HTTP server for API tests (never call real APIs)

### Example Test Structure
```go
func TestDNSCreate(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
        {
            name:    "missing required flags",
            args:    []string{"example.com"},
            wantErr: true,
        },
        {
            name:    "valid create",
            args:    []string{"example.com", "--name", "www", "--type", "A", "--content", "1.2.3.4"},
            wantErr: false,
        },
    }
    // ...
}
```

## Code Style

- Use Go 1.21+ idioms
- Prefer standard library over external dependencies
- Handle errors explicitly, provide helpful messages
- Use context for cancellation
- Follow Cloudflare API naming conventions for consistency
