# 115driver

A comprehensive Go library, CLI tool, and MCP server for [115 cloud storage](https://115.com). It provides a full-featured driver for 115.com's API, supporting login, file operations, upload/download, offline downloads, and more.

[![Go Report Card](https://goreportcard.com/badge/github.com/SheltonZhu/115driver)](https://goreportcard.com/report/github.com/SheltonZhu/115driver)
[![Release](https://img.shields.io/github/release/SheltonZhu/115driver)](https://github.com/SheltonZhu/115driver/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/SheltonZhu/115driver/v4.svg)](https://pkg.go.dev/github.com/SheltonZhu/115driver)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/:License-MIT-orange.svg)](https://raw.githubusercontent.com/SheltonZhu/115driver/main/LICENSE)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI](#cli)
- [MCP Server](#mcp-server)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Features

**Authentication** — Cookie-based login, QR code login, and user identity verification.

**File Operations** — List, rename, move, copy, delete, download, upload (with rapid upload via SHA1 deduplication and multipart upload via Aliyun OSS), search with filters, and get file info/statistics.

**Offline Downloads** — Add HTTP, ED2K, and magnet link download tasks; list, delete, and clear tasks.

**Share** — Create share links and download files via share code.

**Recycle Bin** — List, restore, and permanently delete items.

**CLI** — Full-featured command-line interface with colored table output, JSON mode for scripts, shell completions, and multiple profile support.

**MCP Server** — [Model Context Protocol](https://modelcontextprotocol.io/) server for AI application integration (Claude Desktop, Cursor, etc.).

## Installation

```bash
go get github.com/SheltonZhu/115driver
```

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    // Option 1: Import credentials from cookie string
    cr, err := driver.CredentialFromCookie("your_cookie_string")
    if err != nil {
        log.Fatalf("Failed to create credential: %v", err)
    }

    // Option 2: Create credentials manually
    // cr := &driver.Credential{
    //     UID:  "your_uid",
    //     CID:  "your_cid",
    //     SEID: "your_seid",
    //     KID:  "your_kid",
    // }

    // Create client with credentials
    client := driver.Default().ImportCredential(cr)

    // Check login status
    if err := client.LoginCheck(); err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    log.Println("Successfully logged in!")
}
```

### Common Operations

The examples below assume you have an authenticated `client` (see Basic Usage above).

```go
// Download a file using pickcode
downloadInfo, err := client.Download("pickcode_here")
if err != nil { /* handle error */ }
fileReader, _ := downloadInfo.Get()
defer fileReader.Close()
// write fileReader to file...
```

```go
// Upload a file (auto-selects rapid upload or multipart via OSS)
file, _ := os.Open("/path/to/local/file.zip")
defer file.Close()
fileInfo, _ := file.Stat()
uploadID, err := client.RapidUploadOrByOSS(
    "0",            // parent directory ID ("0" for root)
    fileInfo.Name(),
    fileInfo.Size(),
    file,
)
```

```go
// List files in root directory
files, err := client.List("0")
for _, f := range files {
    log.Printf("File: %s, Size: %d, Type: %s", f.Name, f.Size, f.Type)
}
```

```go
// Search for files
results, err := client.Search(&driver.SearchOption{
    SearchValue: "document",
    Limit:       100,
})
for _, r := range results.Files {
    log.Printf("File: %s, Size: %d", r.Name, r.Size)
}
```

```go
// Add offline download task
taskIDs, err := client.AddOfflineTaskURIs(
    []string{"https://example.com/file.zip"},
    "0", // "0" for root directory
)
```

## CLI

115driver includes a CLI tool for interacting with 115 cloud storage from the command line, designed for both human use (colored table output) and AI agent consumption (`--json` flag).

### Install

```bash
go install github.com/SheltonZhu/115driver/cmd/115driver@latest
```

### Authentication

```bash
# QR code login (interactive)
115driver login

# Cookie login
115driver login --cookie "UID=xxx;CID=xxx;SEID=xxx;KID=xxx"

# Verify identity
115driver whoami

# Account and storage info
115driver info
```

Credentials are stored in `~/.115driver/config.toml` and support multiple profiles.

### Authentication Priority

1. `--cookie` flag
2. `DRIVER115_COOKIE` environment variable
3. Config file (`~/.115driver/config.toml`)

Additional env vars: `DRIVER115_CONFIG` (config path), `DRIVER115_PROFILE` (profile name).

### Commands

```bash
# List files
115driver ls /path/to/dir
115driver ls -l /path/to/dir          # detailed view

# File info
115driver stat /path/to/file

# Account and storage info
115driver info

# Create directories
115driver mkdir /new/dir
115driver mkdir -p /deep/nested/dir   # create parents

# Move / Copy / Rename / Delete
115driver mv /source/file /dest/dir
115driver cp /source/file /dest/dir
115driver rename /path/to/file new_name
115driver rm /path/to/file

# Upload & Download
115driver upload /local/file /remote/dir
115driver download /remote/file /local/dir

# Search
115driver search keyword
115driver search keyword -t video     # filter by type
115driver search keyword --sort size  # sort results

# Offline downloads (HTTP/ED2K/magnet)
115driver offline add <url>
115driver offline add <url> -d /save/dir
115driver offline list
115driver offline rm <hash>
```

### JSON Output

All commands support `--json` for machine-readable output:

```bash
115driver --json ls /path/to/dir
115driver --json stat /path/to/file
115driver --json info
```

### Shell Completion

```bash
# Bash
echo 'source <(115driver completion bash)' >> ~/.bashrc

# Zsh
echo 'source <(115driver completion zsh)' >> ~/.zshrc

# Fish
115driver completion fish > ~/.config/fish/completions/115driver.fish
```

## MCP Server

115driver includes an MCP (Model Context Protocol) server for AI application integration (Claude Desktop, Cursor, etc.).

### Install

**Option 1: go install**

```bash
go install github.com/SheltonZhu/115driver/mcp@latest
```

**Option 2: build from source**

```bash
git clone https://github.com/SheltonZhu/115driver.git
cd 115driver
go build -o 115driver-mcp-server ./mcp/
```

### Usage

```bash
# If installed via go install:
mcp --cookie="UID=xxx;CID=xxx;SEID=xxx;KID=xxx"

# If built from source:
./115driver-mcp-server --cookie="UID=xxx;CID=xxx;SEID=xxx;KID=xxx"
```

### Available Tools

| Category | Tools |
|----------|-------|
| **Account** | `getAccountInfo` |
| **Directory** | `listDirectory` |
| **File** | `stat`, `mkdir`, `delete`, `rename`, `move`, `copy`, `upload_from_url`, `upload_from_local`, `download_file`, `get_download_info` |
| **Search** | `search` |
| **Offline** | `listOfflineTasks`, `addOfflineTaskURIs`, `deleteOfflineTasks`, `clearOfflineTasks` |
| **Share** | `getShareSnap` |
| **Recycle** | `listRecycleBin`, `revertRecycleBin`, `cleanRecycleBin` |

### Configure with Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "115driver": {
      "command": "mcp",
      "args": ["--cookie=UID=xxx;CID=xxx;SEID=xxx;KID=xxx"]
    }
  }
}
```

## API Reference

For detailed API documentation, visit [pkg.go.dev](https://pkg.go.dev/github.com/SheltonZhu/115driver).

## Troubleshooting

### Login Issues

If you encounter login issues:
1. Make sure your cookie is valid and not expired
2. Check that all required fields (UID, CID, SEID, KID) are present
3. Try logging in through the web interface first to obtain a fresh cookie

### Upload/Download Issues

If upload or download fails:
1. Verify file paths are correct
2. Check your internet connection
3. Ensure you have sufficient storage space
4. Check the returned error message for specific details

### Rate Limiting

The 115 API may have rate limits. If you encounter rate limiting errors:
1. Add delays between operations
2. Implement retry logic with exponential backoff
3. Consider using a proxy if needed

## Project Structure

```
115driver/                    # Go 1.23+
├── cmd/
│   └── 115driver/            # CLI entry point (go install binary)
├── cli/                      # CLI implementation
│   ├── cmd/                  # Cobra commands
│   └── internal/             # Internal packages (auth, output, resolver)
├── internal/                 # Shared app-level helpers
├── pkg/
│   ├── driver/               # Core driver (client, login, file, upload, download, search, share, offline)
│   └── crypto/               # Cryptography utilities (ECDH, AES, RSA)
└── mcp/                      # MCP server (stdin/stdout JSON-RPC 2.0)
    ├── main.go               # Entry point
    └── server/tools/         # Tool implementations (account, dir, file, search, offline, share, recycle)
```

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=sheltonzhu/115driver&type=date&legend=top-left)](https://www.star-history.com/#sheltonzhu/115driver&type=date&legend=top-left)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Contributors

<!-- readme: contributors -start -->
<table>
<tr>
    <td align="center">
        <a href="https://github.com/SheltonZhu">
            <img src="https://avatars.githubusercontent.com/u/26734784?v=4" width="100;" alt="SheltonZhu"/>
            <br />
            <sub><b>SheltonZhu</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/power721">
            <img src="https://avatars.githubusercontent.com/u/2384040?v=4" width="100;" alt="power721"/>
            <br />
            <sub><b>power721</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/xhofe">
            <img src="https://avatars.githubusercontent.com/u/36558727?v=4" width="100;" alt="xhofe"/>
            <br />
            <sub><b>xhofe</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/Ovear">
            <img src="https://avatars.githubusercontent.com/u/1362137?v=4" width="100;" alt="Ovear"/>
            <br />
            <sub><b>Ovear</b></sub>
        </a>
    </td></tr>
</table>
<!-- readme: contributors -end -->

## License

[MIT](LICENSE)
