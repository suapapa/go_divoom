# AGENTS.md

This file defines repository-specific guidance for AI/code agents working in this project.

## Project Overview

- Module: `github.com/suapapa/go_divoom`
- Language: Go
- Purpose: Go client for Divoom HTTP/Web API devices (for example Pixoo64)

## Working Rules

- Keep changes focused and minimal.
- Preserve public API compatibility unless explicitly requested.
- Do not modify unrelated files.
- Do not revert existing user changes unless asked.

## Code Style

- Use idiomatic Go.
- Handle all errors explicitly.
- Prefer small helpers over duplicated request/response code.
- Keep exported symbols documented when newly added.

## Validation Checklist

Run these commands after substantive changes:

```bash
gofmt -w *.go _example/basic_controls/main.go _example/play_gif/main.go
go test ./...
go vet ./...
golangci-lint run
```

If lint output includes cache permission warnings in restricted environments, still verify there are no file/line issue reports.

## Repo Layout Notes

- Core package files are at repository root (`package divoom`).
- Runnable samples are under `./_example`.
- Image assets are under `./_img`.

## Documentation Expectations

When behavior changes:

- Update `README.md` quick start or feature notes if needed.
- Keep examples aligned with current API usage.
