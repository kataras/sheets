# Sheets

[![build status](https://img.shields.io/travis/kataras/sheets/master.svg?style=for-the-badge&logo=travis)](https://travis-ci.org/kataras/sheets) [![report card](https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=for-the-badge)](https://goreportcard.com/report/github.com/kataras/sheets) [![godocs](https://img.shields.io/badge/go-%20docs-488AC7.svg?style=for-the-badge)](https://godoc.org/github.com/kataras/sheets)

A [Google Sheets](https://docs.google.com/spreadsheets) Client written in Go.

This package is under active development and a **work-in-progress** project. You should NOT use it on production. Please consider using the official [Google's Sheets client for Go](https://developers.google.com/sheets/api/quickstart/go) instead.

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl).

```sh
$ go get -u github.com/kataras/sheets
```

## Getting Started

First of all, navigate to <https://developers.google.com/sheets/api> and enable the Sheets API Service in your [Google Console](https://console.cloud.google.com/). Place the secret client service account or token file as `client_secret.json` near the executable example.

Example Code:

```go
package main

import (
    "context"
    "time"

    "github.com/kataras/sheets"
)

func main() {
    ctx := context.TODO()
    client := sheets.NewClient(sheets.ServiceAccount(ctx, "client_secret.json"))

    var (
        spreadsheetID := "1Ku0YXrcy8Nqmji7ABS8AmLAyxP5duQIRwmaAJAqyMYY"
        dataRange := "NamedRange or selectors like A1:E4 or *"
        records []struct{
            Timestamp time.Time
            Email     string
            Username  string
            IgnoredMe string `sheets:"-"`
        }{}
    )

    err := client.ReadSpreadsheet(ctx, &records, spreadsheetID, dataRange)
    if err != nil {
        panic(err)
    }

    // [...]

    err = client.UpdateSpreadsheet(ctx, spreadsheetID, "A2:Z", [][]interface{}{
        {"updated record value: 1.1", "updated record value: 1.2"},
        {"updated record value: 2.1", "updated record value: 2.2"},
    })

    // [...]
}
```

## License

This software is licensed under the [MIT License](LICENSE).
