# divoom-mcp

`cmd/divoom-mcp` is an MCP server for `go_divoom`.

It supports both `stdio` and `sse` transports, and exposes Divoom features as MCP tools.

## Run

### stdio mode

```bash
go run ./cmd/divoom-mcp -mode stdio
```

### sse mode

```bash
go run ./cmd/divoom-mcp -mode sse -addr 127.0.0.1:8080
```

## Usage Notes

- Call `connect_device` before using device control tools.
- Discovery/query tools (`find_device`, `dial_type`, `dial_list`, `get_font_list`) work without a connected device.
