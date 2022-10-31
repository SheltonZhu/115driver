# 115driver

![Version](https://img.shields.io/badge/release-v1.0.9-brightgreen?style=flat-square) [![Reference](https://img.shields.io/badge/Go-Reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/SheltonZhu/115driver) ![License](https://img.shields.io/:License-MIT-green.svg?style=flat-square)

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

    client = driver.Defalut().ImportCredential(cr)
    if err := driver.LoginCheck(); err != nil {
        log.Fatalf("login error: %s", err)
    }
}

```

More examples can be found in [reference](https://pkg.go.dev/github.com/SheltonZhu/115driver).

## Features

* Login
  * [X] Import credential from cookies
  * [ ] Login via QRCode
  * [ ] Get signed-in user information
* File
  * [X] List
  * [X] Rename
  * [X] Move
  * [X] Copy
  * [X] Delete
  * [X] Make Directory
  * [x] Download
  * [x] Upload SHA1
  * [x] Upload
  * [ ] Search
  * [ ] Get Information by ID
  * [ ] Stat File

## License

[MIT](LICENSE)
