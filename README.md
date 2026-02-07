# 115driver

A comprehensive Go library for interacting with 115 cloud storage. This package provides a full-featured driver for 115.com's API, supporting login, file operations, upload/download, and more.

[![Go Report Card](https://goreportcard.com/badge/github.com/SheltonZhu/115driver)](https://goreportcard.com/report/github.com/SheltonZhu/115driver)
[![Release](https://img.shields.io/github/release/SheltonZhu/115driver)](https://github.com/SheltonZhu/115driver/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/SheltonZhu/115driver/v4.svg)](https://pkg.go.dev/github.com/SheltonZhu/115driver)
[![License](https://img.shields.io/:License-MIT-orange.svg)](https://raw.githubusercontent.com/SheltonZhu/115driver/main/LICENSE)


## Installation

```bash
go get github.com/SheltonZhu/115driver
```

## Features

### Authentication
- [X] Import credentials from cookies
- [X] QR code login
- [X] Get signed-in user information

### File Operations
- [X] List files and directories
- [X] Rename files and directories
- [X] Move files and directories
- [X] Copy files and directories
- [X] Delete files and directories
- [X] Create directories
- [X] Download files
- [X] Upload files
- [X] Rapid upload (similar to Google Drive's rapid upload)
- [X] Search files
- [X] Get file information by ID
- [X] Get file statistics
- [X] Download files via share code
- [X] Offline download

### Recycle Bin
- [X] List deleted items
- [X] Restore deleted items
- [X] Clean recycle bin

### Additional Features
- [X] Share files via share code
- [X] File tagging
- [X] Batch operations

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    // Create credentials from cookie string
    cr, err := driver.CredentialFromCookie("your_cookie_string")
    if err != nil {
        log.Fatalf("Failed to create credential: %v", err)
    }

    // Or create manually
    cr := &driver.Credential{
        UID:  "your_uid",
        CID:  "your_cid",
        SEID: "your_seid",
        KID:  "your_kid",
    }

    // Create client with credentials
    client := driver.Default().ImportCredential(cr)

    // Check login status
    if err := client.LoginCheck(); err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    log.Println("Successfully logged in!")
}
```

### Download a File

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "io"
    "log"
    "os"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // Download a file using pickcode
    pickCode := "abc123"
    downloadInfo, err := client.Download(pickCode)
    if err != nil {
        log.Fatalf("Download failed: %v", err)
    }

    // Save the file
    outFile, err := os.Create("/path/to/save/file.zip")
    if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer outFile.Close()

    fileReader, err := downloadInfo.Get()
    if err != nil {
        log.Fatalf("Failed to get file reader: %v", err)
    }
    defer fileReader.Close()

    if _, err := io.Copy(outFile, fileReader); err != nil {
        log.Fatalf("Failed to save file: %v", err)
    }

    log.Println("Download completed!")
}
```

### Upload a File

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
    "os"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // Open the file
    file, err := os.Open("/path/to/local/file.zip")
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    defer file.Close()

    // Get file info
    fileInfo, err := file.Stat()
    if err != nil {
        log.Fatalf("Failed to get file info: %v", err)
    }

    // Rapid upload (fast upload using file hash)
    uploadID, err := client.RapidUploadOrByOSS(
        "0", // parent directory ID (0 for root)
        fileInfo.Name(),
        fileInfo.Size(),
        file,
    )
    if err != nil {
        log.Fatalf("Upload failed: %v", err)
    }

    log.Printf("Upload started, init response: %+v", uploadID)
}
```

### Rapid Upload

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "io"
    "log"
    "os"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // Open the file
    file, err := os.Open("/path/to/local/file.zip")
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    defer file.Close()

    // Get file info
    fileInfo, err := file.Stat()
    if err != nil {
        log.Fatalf("Failed to get file info: %v", err)
    }

    // Rapid upload using file hash
    uploadID, err := client.RapidUploadOrByOSS(
        "0", // parent directory ID (0 for root)
        fileInfo.Name(),
        fileInfo.Size(),
        file,
    )
    if err != nil {
        log.Fatalf("Rapid upload failed: %v", err)
    }

    log.Printf("Rapid upload started, init response: %+v", uploadID)
}
```

### List Files in a Directory

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // List files in root directory
    files, err := client.List("0")
    if err != nil {
        log.Fatalf("List failed: %v", err)
    }

    for _, file := range files {
        log.Printf("File: %s, Size: %d, Type: %s", file.Name, file.Size, file.Type)
    }
}
```

### Search Files

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // Search for files
    keyword := "document"
    results, err := client.Search(&driver.SearchOption{
        SearchValue: keyword,
        Limit:       100,
    })
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    log.Printf("Found %d results", results.Count)
    for _, result := range results.Files {
        log.Printf("File: %s, Size: %d", result.Name, result.Size)
    }
}
```

### Offline Download

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    client := driver.Default()

    // Create credentials and login
    cr, _ := driver.CredentialFromCookie("your_cookie")
    client = client.ImportCredential(cr)
    client.LoginCheck()

    // Add offline download task
    url := "https://example.com/file.zip"
    taskIDs, err := client.AddOfflineTaskURIs([]string{url}, "0") // "0" for root directory
    if err != nil {
        log.Fatalf("Offline download failed: %v", err)
    }

    log.Printf("Offline download task created with hash: %s", taskIDs[0])
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
115driver/
├── pkg/
│   ├── driver/          # Core driver implementation
│   │   ├── client.go    # Client interface
│   │   ├── login.go     # Authentication
│   │   ├── file.go      # File operations
│   │   ├── upload.go    # Upload functionality
│   │   ├── download.go  # Download functionality
│   │   ├── search.go    # Search
│   │   ├── share.go     # Share files
│   │   ├── offline.go   # Offline download
│   │   └── ...          # Other modules
│   └── crypto/          # Cryptography utilities
└── mcp/                 # MCP server implementation
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
