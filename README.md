# ytc — YouTrack CLI

A fast, cross-platform command-line client for [JetBrains YouTrack](https://www.jetbrains.com/youtrack/). Built in Go for instant startup (~5ms), optimized for both human developers and AI agents.

## Features

- **Issue Management** — Create, view, update, search, and manage issues
- **Project Management** — List and inspect projects
- **Custom Fields** — Update state, priority, assignee, type with simple commands
- **AI Agent Optimized** — `--json`, `--fields`, `--quiet` flags for minimal token usage
- **Cross-Platform** — Single binary for macOS, Linux, and Windows
- **Fast** — ~5ms startup time (vs ~400ms for Python-based tools)
- **Zero Dependencies** — One binary, no runtime needed

> **Note:** Currently tested on macOS (Apple Silicon) only. Windows and Linux builds are supported via cross-compilation but have not been tested on actual devices. Contributions and feedback welcome!

## Installation

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/dustinle98/youtrack-cli/releases):

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `ytc_*_darwin_arm64.tar.gz` |
| macOS (Intel) | `ytc_*_darwin_amd64.tar.gz` |
| Linux (x86_64) | `ytc_*_linux_amd64.tar.gz` |
| Linux (ARM64) | `ytc_*_linux_arm64.tar.gz` |
| Windows (x86_64) | `ytc_*_windows_amd64.zip` |
| Windows (ARM64) | `ytc_*_windows_arm64.zip` |

**macOS / Linux:**
```bash
# Download and extract (example for macOS Apple Silicon)
tar -xzf ytc_*_darwin_arm64.tar.gz
sudo mv ytc /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
# Extract the zip, then add to PATH or move to a directory in PATH
Expand-Archive ytc_*_windows_amd64.zip -DestinationPath C:\Tools
# Add C:\Tools to your system PATH
```

### From Source (requires Go 1.22+)

```bash
git clone https://github.com/dustinle98/youtrack-cli.git
cd youtrack-cli
go build -o ytc .

# Install to PATH
sudo cp ytc /usr/local/bin/    # macOS/Linux
# or copy ytc.exe to PATH      # Windows
```

### Cross-Compile

```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o ytc-mac-arm .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o ytc-mac-intel .

# Linux
GOOS=linux GOARCH=amd64 go build -o ytc-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o ytc.exe .
```

## Quick Start

```bash
# 1. Authenticate
ytc auth login

# 2. List projects
ytc project list

# 3. List issues
ytc issue list -p MYPROJECT

# 4. View an issue
ytc issue view MYPROJECT-123

# 5. Search
ytc search "bug in login"
```

## Configuration

`ytc` stores configuration in `~/.config/ytc/config.yml`. You can also use environment variables:

| Env Variable | Description |
|---|---|
| `YOUTRACK_URL` | YouTrack instance URL (e.g. `https://myteam.youtrack.cloud`) |
| `YOUTRACK_API_TOKEN` | API token from YouTrack → Profile → Authentication |
| `YOUTRACK_VERIFY_SSL` | Set to `false` to skip SSL verification |

Environment variables take precedence over the config file.

## Commands

### Authentication

```bash
ytc auth login       # Interactive setup (URL + token)
ytc auth status      # Show connection status
```

### Issues

```bash
# List & Search
ytc issue list -p PROJECT              # List unresolved issues
ytc issue list -p PROJECT -s "Open"    # Filter by state
ytc issue list -p PROJECT -n 50        # Limit results

# View
ytc issue view PROJECT-123             # Human-readable details

# Create
ytc issue create -p PROJECT -s "Bug: login fails" -d "Steps to reproduce..."

# Update
ytc issue update PROJECT-123 -s "New title"
ytc issue update PROJECT-123 -d "New description"

# Custom Field Updates
ytc issue state PROJECT-123 "In Progress"
ytc issue assign PROJECT-123 john.doe
ytc issue priority PROJECT-123 "Critical"
ytc issue type PROJECT-123 "Bug"

# Comments
ytc issue comment PROJECT-123 "Fixed in commit abc123"
```

### Projects

```bash
ytc project list              # List all projects
ytc project view PROJECT      # View project details
```

### Search

```bash
ytc search "bug in login"                              # Free text search
ytc search -p PROJECT -s "Open"                        # Structured filters
ytc search -p PROJECT -a john.doe -s "In Progress"     # Multiple filters
ytc search "project: PROJECT #Unresolved"              # YouTrack query syntax
```

### Aliases

| Full Command | Alias |
|---|---|
| `ytc issue` | `ytc i` |
| `ytc issue list` | `ytc i ls` |
| `ytc project` | `ytc p` |
| `ytc project list` | `ytc p ls` |
| `ytc search` | `ytc s` or `ytc find` |

## Output Formats

### Table (default)

```
$ ytc issue list -p PROJECT -n 3

ID            STATE           ASSIGNEE   SUMMARY
────────────  ──────────────  ─────────  ─────────────────────────────────
PROJECT-101   Ready for Work  alice      Fix authentication timeout issue
PROJECT-98    In Development  bob        Add dark mode support
PROJECT-95    Open            charlie    Update user onboarding flow
```

### JSON (`--json`)

```bash
$ ytc issue view PROJECT-123 --json
{
  "id": "2-12345",
  "idReadable": "PROJECT-123",
  "summary": "Fix authentication timeout",
  "customFields": [...]
}
```

### Filtered JSON (`--json --fields`)

```bash
$ ytc issue list -p PROJECT -n 2 --json --fields idReadable,summary
[
  { "idReadable": "PROJECT-101", "summary": "Fix authentication timeout issue" },
  { "idReadable": "PROJECT-98", "summary": "Add dark mode support" }
]
```

### Quiet (`--quiet`)

```bash
$ ytc issue list -p PROJECT -n 3 --quiet
PROJECT-101
PROJECT-98
PROJECT-95
```

## AI Agent Integration

`ytc` is designed for AI agent usage with minimal token consumption.

### Setup for AI Agents

Add this to your agent's system prompt or tool instructions:

```
You have access to `ytc` CLI for YouTrack issue tracking. Key commands:
- ytc issue list -p <PROJECT> --json
- ytc issue view <ID> --json --fields idReadable,summary,state
- ytc issue create -p <PROJECT> -s "summary" -d "description" --json
- ytc issue state <ID> "<state>"
- ytc issue assign <ID> <login>
- ytc issue comment <ID> "text"
- ytc search "query" --json
- ytc project list --json
All commands support: --json, --fields <comma-sep>, --quiet
```

### Why CLI over MCP for Agents?

| | MCP (30+ tools) | CLI (`ytc`) |
|---|---|---|
| Context overhead | ~3000-5000 tokens (tool schemas) | ~200 tokens (prompt instruction) |
| Per-call input | ~50 tokens (JSON-RPC) | ~10 tokens (command string) |
| Per-call output | Unstructured string | JSON with `--fields` filter |
| Startup time | ~400ms (Python) | ~5ms (Go binary) |

### Example Agent Workflow

```bash
# Agent reads issue
ytc issue view PROJECT-123 --json --fields idReadable,summary,state
# → {"idReadable":"PROJECT-123","summary":"Fix login bug","state":"Open"}

# Agent updates state
ytc issue state PROJECT-123 "In Progress"
# → ✓ PROJECT-123 → State: In Progress

# Agent adds comment
ytc issue comment PROJECT-123 "Started working on fix"
# → ✓ Comment added to PROJECT-123
```

## Project Structure

```
youtrack-cli/
├── main.go                     # Entry point
├── go.mod / go.sum
├── Makefile
├── cmd/
│   ├── root.go                 # Root command + global flags
│   ├── auth.go                 # ytc auth login/status
│   ├── issue.go                # ytc issue subcommands
│   ├── project.go              # ytc project subcommands
│   └── search.go               # ytc search
└── internal/
    ├── config/config.go        # Config (YAML + env vars)
    ├── api/
    │   ├── client.go           # HTTP client (auth, retry, backoff)
    │   ├── issues.go           # Issues API
    │   └── projects.go         # Projects & Users API
    └── output/formatter.go     # Output formatting
```

## Development

```bash
# Build
make build

# Run tests
make test

# Clean
make clean
```

## Acknowledgments

Inspired by [glab](https://gitlab.com/gitlab-org/cli) (GitLab CLI) and [youtrack-mcp](https://github.com/tonyzorin/youtrack-mcp) (YouTrack MCP Server).

## License

MIT
