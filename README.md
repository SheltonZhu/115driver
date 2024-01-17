# 115driver

🖴 A 115 cloud driver package.

[![Goreport](https://goreportcard.com/badge/github.com/SheltonZhu/115driver)](https://goreportcard.com/report/github.com/SheltonZhu/115driver) [![Version](https://img.shields.io/badge/release-v1.0.23-blue)](https://github.com/SheltonZhu/115driver/tags) [![Reference](https://img.shields.io/badge/Go-Reference-red.svg)](https://pkg.go.dev/github.com/SheltonZhu/115driver) [![License](https://img.shields.io/:License-MIT-orange.svg)](https://raw.githubusercontent.com/SheltonZhu/115driver/main/LICENSE)

---

## Features

* Login
  * [X] Import credential from cookies
  * [x] Login via QRCode
  * [X] Get signed-in user information
* File
  * [X] List
  * [X] Rename
  * [X] Move
  * [X] Copy
  * [X] Delete
  * [X] Make Directory
  * [X] Download
  * [X] Upload
  * [X] Rapid Upload
  * [ ] Search
  * [X] Get Information by ID
  * [X] Stat File
  * [x] Download by share code

## Example

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    cr := &driver.Credential{
        UID: "xxx",
        CID: "xxx",
        SEID: "xxx",
    }
    // or err := cr.FromCookie(cookieStr)

    client := driver.Defalut().ImportCredential(cr)
    if err := driver.LoginCheck(); err != nil {
        log.Fatalf("login error: %s", err)
    }
}

```

More examples can be found in [reference](https://pkg.go.dev/github.com/SheltonZhu/115driver).

## License

[MIT](LICENSE)
