# thrift2x

`thrift2x` is a Go-based CLI tool for batch-converting Thrift IDL files.

Today it supports one target:

- `typescript` → generates `.ts` type definitions from `.thrift`

## Why thrift2x

- Recursive conversion for entire IDL directories
- Parallel parsing/generation (`--jobs auto` or fixed workers)
- Path-based exclude rules
- Pluggable target system for future language backends

## Installation

### Install via `go install`

```bash
go install github.com/81120/thrift2x/cmd/thrift2x@latest
```

### Build locally

```bash
git clone <your-fork-or-this-repo>
cd thrift2x
go build -o thrift2x ./cmd/thrift2x
```

## Quick Start

Given a project structure like:

```text
./idl
├── user.thrift
└── order.thrift
```

Run:

```bash
thrift2x generate --in ./idl --out ./gen --target typescript
```

Generated files will be written to `./gen` with the same relative directory layout.

## CLI

### List supported targets

```bash
thrift2x targets
```

Current output:

```text
typescript
```

### Generate code

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript
```

## Configuration

### Required flags

- `--in`: input directory containing `.thrift` files
- `--out`: output directory
- `--target`: generation target

### Optional flags

- `--exclude`: comma-separated path substrings to skip
  - Example: `--exclude third_party,legacy`
- `--jobs`: worker count (`auto` by default)
  - Example: `--jobs 8`
- `--i64-type`: TypeScript-only i64 mapping
  - Values: `string` (default), `number`, `bigint`

## Examples

### TypeScript with bigint for i64

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --i64-type bigint
```

### Exclude some paths

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --exclude third_party,legacy
```

### Set fixed concurrency

```bash
thrift2x generate --in ./idl --out ./gen --target typescript --jobs 8
```

## Error Behavior

- Missing required flags (`--in`, `--out`, `--target`) returns an error
- Unknown target returns an error and shows available targets
- `--i64-type` is ignored for non-TypeScript targets

## Development

Run tests:

```bash
go test ./...
```

Run locally without install:

```bash
go run ./cmd/thrift2x --help
go run ./cmd/thrift2x generate --help
```

## Architecture Overview

### CLI layer

- `cmd/thrift2x/main.go`
  - entry point
  - normalizes legacy single-dash long flags (`-target` → `--target`)
- `cmd/thrift2x/root.go`
  - root command and subcommands (`generate`, `targets`)

### Conversion orchestration

- `internal/converter/converter.go`
  - validates config, resolves target, runs full pipeline
- `internal/converter/scan.go`
  - file discovery + exclude filtering
- `internal/converter/worker.go`
  - worker pool + jobs resolution
- `internal/converter/process.go`
  - parse, generate, and write each file

### Parser and AST

- `internal/parser/parser.go`
  - custom lexer/parser
- `internal/ast/ast.go`
  - shared AST model used by targets

### Target system

- `internal/targets/target.go`
  - target/generator interfaces
- `internal/targets/registry.go`
  - register/get/list target registry
- `internal/targets/typescript/*`
  - built-in TypeScript target
