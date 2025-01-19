# 115driver

🖴 A 115 cloud driver package.

[![Goreport](https://goreportcard.com/badge/github.com/SheltonZhu/115driver)](https://goreportcard.com/report/github.com/SheltonZhu/115driver) [![Release](https://img.shields.io/github/release/SheltonZhu/115driver)](https://github.com/SheltonZhu/115driver/releases) [![Reference](https://img.shields.io/badge/Go-Reference-red.svg)](https://pkg.go.dev/github.com/SheltonZhu/115driver) [![License](https://img.shields.io/:License-MIT-orange.svg)](https://raw.githubusercontent.com/SheltonZhu/115driver/main/LICENSE) [![Downloads](https://img.shields.io/github/downloads/SheltonZhu/115driver/total?color=%239F7AEA&logo=github)](https://github.com/SheltonZhu/115driver/releases)

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
  * [x] Offline Download
* Recycle Bin
  * [x] List
  * [x] Revert
  * [x] Clean

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
        KID: "xxx",
    }
    // or err := cr.FromCookie(cookieStr)

    client := driver.Defalut().ImportCredential(cr)
    if err := client.LoginCheck(); err != nil {
        log.Fatalf("login error: %s", err)
    }
}

```

More examples can be found in [reference](https://pkg.go.dev/github.com/SheltonZhu/115driver).

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
    </td>
    <td align="center">
        <a href="https://github.com/power721">
            <img src="https://avatars.githubusercontent.com/u/2384040?v=4" width="100;" alt="power721"/>
            <br />
            <sub><b>power721</b></sub>
        </a>
    </td></tr>
</table>
<!-- readme: contributors -end -->

## License

[MIT](LICENSE)
