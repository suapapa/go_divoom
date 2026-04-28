# go_divoom

![example](_img/go_divoom.jpg)

Go client package for the Divoom Web API (for example Pixoo64).

- Package docs: <https://pkg.go.dev/github.com/suapapa/go_divoom>

## Install

```bash
go get github.com/suapapa/go_divoom
```

## Requirements

- Device and host must be on the same LAN
- Device must support Divoom HTTP/Web API

## Quick Start

```go
package main

import (
	"log"

	divoom "github.com/suapapa/go_divoom"
)

func main() {
	devs, err := divoom.FindDevice()
	if err != nil {
		log.Fatal(err)
	}
	if len(devs) == 0 {
		log.Fatal("no divoom device found on LAN")
	}

	c := divoom.NewClient(devs[0])

	if err := c.SetBrightness(50); err != nil {
		log.Fatal(err)
	}
	if err := c.SelectChannel(divoom.ChannelFaces); err != nil {
		log.Fatal(err)
	}
}
```

## Examples

- `./_example/basic_controls`
- `./_example/play_gif`

## Development

Run these checks before committing:

```bash
gofmt -w *.go _example/basic_controls/main.go _example/play_gif/main.go
go test ./...
go vet ./...
golangci-lint run
```

## References

- Divoom API docs: <http://doc.divoom-gz.com/web/#/12?page_id=143>
