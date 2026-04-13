# thrift2x

`thrift2x` is a batch Thrift conversion tool written in Go. It currently supports converting `.thrift` files into TypeScript type definitions.

## Feature Overview

- Recursively scans `.thrift` files under the input directory
- Supports excluding files/directories by path substring
- Parses and generates in parallel, with support for `jobs=auto`
- Uses a target-language plugin mechanism to output corresponding file extensions
- Built-in target: `typescript`

---

## Architecture

### 1) CLI Layer (Command Entry)

- `cmd/thrift2x/main.go`
  - Program entry point, runs `newRootCmd().Execute()`
  - Includes `normalizeLegacySingleDashLongFlags()`: normalizes flags like `-target` to `--target`
- `cmd/thrift2x/root.go`
  - Root command: `thrift2x`
  - Subcommands:
    - `generate`: performs conversion
    - `targets`: lists supported targets
  - Key `generate` flags:
    - `--in` input directory (required)
    - `--out` output directory (required)
    - `--target` output target (required)
    - `--exclude` excluded path substrings (optional, comma-separated)
    - `--jobs` concurrency level (integer or `auto`)
    - `--i64-type` only applies to TypeScript target (`string|number|bigint`)

### 2) Conversion Orchestration Layer (converter)

- `internal/converter/converter.go`
  - `Run(cfg Config)` handles the full workflow:
    1. Validate input arguments
    2. Load target from registry using `cfg.Target`
    3. Create generator (with target options)
    4. Scan Thrift files
    5. Resolve concurrency level
    6. Process files concurrently
    7. Aggregate results and timing stats
- `internal/converter/scan.go`
  - `collectThriftFiles`: recursively collects `.thrift` files
  - `shouldExclude`: filters by path substring
- `internal/converter/worker.go`
  - `runWorkers`: fixed worker pool for parallel processing
  - `resolveJobs/autoJobs`: parses `jobs` or auto-calculates concurrency
- `internal/converter/process.go`
  - `processFile`:
    1. Read file
    2. Parse AST with parser
    3. Generate code via target generator
    4. Compute relative path and write to output directory
    5. Remove existing target file if generated content is empty

### 3) Syntax Parsing Layer (parser + AST)

- `internal/parser/parser.go`
  - Custom lexer + parser
  - Outputs a unified AST structure
- `internal/ast/ast.go`
  - Defines AST nodes: `File/Decl/StructLike/Enum/Typedef/TypeRef`, etc.
  - Provides a unified input model for target generators

### 4) Target Extension Layer

- `internal/targets/target.go`
  - Defines `Target` and `Generator` interfaces
- `internal/targets/registry.go`
  - Global registry: `Register/Get/List`
- `internal/targets/typescript/*`
  - TypeScript target implementation and tests
- `internal/converter/converter.go`
  - Registers built-in targets via blank import (currently only `typescript`)

---

## Execution Flow

1. User runs `thrift2x generate ...`
2. CLI assembles `converter.Config`
3. Converter selects generator by target
4. Scans `.thrift` file list
5. Worker pool parses + generates in parallel
6. Writes output files (`.ts`) using paths relative to input directory
7. Prints summary stats (`total/success/failed/jobs + timing`)

---

## Usage

## 1. List supported targets

```bash
thrift2x targets
```

Current output:

```text
typescript
```

## 2. Generate TypeScript

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript
```

## 3. Customize i64 mapping (TypeScript only)

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --i64-type bigint
```

Available values: `string` (default), `number`, `bigint`

## 4. Exclude specific paths

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --exclude third_party,legacy
```

## 5. Adjust concurrency

```bash
# Auto concurrency (default)
thrift2x generate --in ./idl --out ./gen --target typescript --jobs auto

# Fixed concurrency
thrift2x generate --in ./idl --out ./gen --target typescript --jobs 8
```

---

## Errors and Constraints

- `--target` is required
- `--in` and `--out` are required
- If target does not exist, an error is returned with available targets
- `--i64-type` is ignored for non-TypeScript targets

---

## Development and Testing

```bash
go test ./...
```

To view local command help:

```bash
go run ./cmd/thrift2x --help
go run ./cmd/thrift2x generate --help
```
