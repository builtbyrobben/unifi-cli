# unifi-cli

Command-line interface for the UniFi Site Manager cloud API. Manage hosts, sites, devices, and ISP metrics from your terminal.

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/builtbyrobben/unifi-cli/releases).

### Build from Source

```bash
git clone https://github.com/builtbyrobben/unifi-cli.git
cd unifi-cli
make build
```

## Configuration

unifi-cli authenticates via a UniFi Site Manager API key. You can provide it in two ways:

**Environment variable (recommended for CI/scripts):**

```bash
export UNIFI_API_KEY="your-api-key"
```

**Keyring storage (recommended for interactive use):**

```bash
# Interactive prompt (secure)
unifi-cli auth set-key --stdin

# Pipe from environment
echo "$UNIFI_API_KEY" | unifi-cli auth set-key --stdin
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `UNIFI_API_KEY` | API key (overrides keyring) |
| `UNIFI_CLI_COLOR` | Color output: `auto`, `always`, `never` |
| `UNIFI_CLI_OUTPUT` | Default output mode: `json`, `plain` |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output JSON to stdout (best for scripting) |
| `--plain` | Output stable, parseable text (TSV; no colors) |
| `--color` | Color output: `auto`, `always`, `never` |
| `--verbose` | Enable verbose logging |
| `--force` | Skip confirmations for destructive commands |
| `--no-input` | Never prompt; fail instead (useful for CI) |

## Commands

### auth

Manage authentication credentials.

```bash
# Store API key in system keyring (interactive prompt)
unifi-cli auth set-key --stdin

# Check authentication status
unifi-cli auth status

# Remove stored credentials
unifi-cli auth remove
```

### hosts

Manage UniFi console hosts.

```bash
# List all hosts
unifi-cli hosts list

# List with pagination
unifi-cli hosts list --page-size 10

# Get a specific host by ID
unifi-cli hosts get abc123

# Get host details as JSON
unifi-cli hosts get abc123 --json
```

### sites

Manage UniFi sites.

```bash
# List all sites
unifi-cli sites list

# List with pagination
unifi-cli sites list --page-size 10

# Output as JSON
unifi-cli sites list --json
```

### devices

Manage UniFi network devices.

```bash
# List all devices across all hosts
unifi-cli devices list

# Filter devices by host ID
unifi-cli devices list --host abc123

# List with pagination
unifi-cli devices list --page-size 20

# Output as JSON
unifi-cli devices list --json
```

### isp-metrics

View ISP performance metrics (latency, bandwidth, packet loss).

```bash
# Get 5-minute interval metrics
unifi-cli isp-metrics get 5m

# Get 1-hour interval metrics
unifi-cli isp-metrics get 1h

# Specify duration (24h, 7d, or 30d)
unifi-cli isp-metrics get 5m --duration 24h
unifi-cli isp-metrics get 1h --duration 7d
unifi-cli isp-metrics get 1h --duration 30d

# Custom time range (epoch timestamps)
unifi-cli isp-metrics get 5m --begin 1700000000 --end 1700100000

# Output as JSON
unifi-cli isp-metrics get 1h --duration 24h --json
```

### version

Print version information.

```bash
unifi-cli version
```

## License

MIT
