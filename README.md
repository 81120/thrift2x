# thrift2x

`thrift2x` 是一个用 Go 编写的 Thrift 文件批量转换工具，当前支持将 `.thrift` 转成 TypeScript 类型定义。

## 功能概览

- 递归扫描输入目录下的 `.thrift` 文件
- 支持按路径子串排除文件/目录
- 并行解析与生成，支持 `jobs=auto`
- 基于目标语言（target）插件机制输出对应文件扩展名
- 当前内置 target：`typescript`

---

## 架构设计

### 1) CLI 层（命令入口）

- `cmd/thrift2x/main.go`
  - 程序入口，执行 `newRootCmd().Execute()`
  - 包含 `normalizeLegacySingleDashLongFlags()`：把形如 `-target` 归一为 `--target`
- `cmd/thrift2x/root.go`
  - 根命令 `thrift2x`
  - 子命令：
    - `generate`：执行转换
    - `targets`：列出支持的 target
  - `generate` 关键参数：
    - `--in` 输入目录（必填）
    - `--out` 输出目录（必填）
    - `--target` 输出目标（必填）
    - `--exclude` 排除路径子串（可选，逗号分隔）
    - `--jobs` 并行度（整数或 `auto`）
    - `--i64-type` 仅 TypeScript target 生效（`string|number|bigint`）

### 2) 转换编排层（converter）

- `internal/converter/converter.go`
  - `Run(cfg Config)` 负责整体流程：
    1. 校验输入参数
    2. 根据 `cfg.Target` 从 registry 获取 target
    3. 创建 generator（携带 target options）
    4. 扫描 thrift 文件
    5. 解析并行度
    6. 并发处理每个文件
    7. 聚合结果与耗时统计
- `internal/converter/scan.go`
  - `collectThriftFiles`：递归遍历目录，收集 `.thrift`
  - `shouldExclude`：基于路径子串过滤
- `internal/converter/worker.go`
  - `runWorkers`：固定 worker 池并行处理文件
  - `resolveJobs/autoJobs`：解析 `jobs` 或自动计算并发
- `internal/converter/process.go`
  - `processFile`：
    1. 读文件
    2. 调用 parser 解析 AST
    3. 调用 target generator 生成代码
    4. 计算相对路径并写入输出目录
    5. 若生成内容为空则删除已有目标文件

### 3) 语法解析层（parser + AST）

- `internal/parser/parser.go`
  - 自定义 lexer + parser
  - 输出统一 AST 结构
- `internal/ast/ast.go`
  - 定义 AST 节点：`File/Decl/StructLike/Enum/Typedef/TypeRef` 等
  - 为 target generator 提供统一输入模型

### 4) Target 扩展层

- `internal/targets/target.go`
  - 定义 `Target` 与 `Generator` 接口
- `internal/targets/registry.go`
  - 全局注册中心：`Register/Get/List`
- `internal/targets/typescript/*`
  - TypeScript target 实现与测试
- `internal/converter/converter.go`
  - 通过空白导入注册内置 target（当前仅 `typescript`）

---

## 执行流程

1. 用户执行 `thrift2x generate ...`
2. CLI 组装 `converter.Config`
3. Converter 按 target 选择 generator
4. 扫描 `.thrift` 文件列表
5. Worker 池并行解析 + 生成
6. 按输入目录相对路径写出目标文件（`.ts`）
7. 输出统计信息（total/success/failed/jobs + timing）

---

## 使用说明

## 1. 查看支持的 target

```bash
thrift2x targets
```

当前输出：

```text
typescript
```

## 2. 生成 TypeScript

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript
```

## 3. 自定义 i64 映射（仅 TypeScript 生效）

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --i64-type bigint
```

可选值：`string`（默认）、`number`、`bigint`

## 4. 排除部分路径

```bash
thrift2x generate \
  --in ./idl \
  --out ./gen \
  --target typescript \
  --exclude third_party,legacy
```

## 5. 调整并行度

```bash
# 自动并行（默认）
thrift2x generate --in ./idl --out ./gen --target typescript --jobs auto

# 固定并行
thrift2x generate --in ./idl --out ./gen --target typescript --jobs 8
```

---

## 错误与约束

- `--target` 必填
- `--in` 与 `--out` 必填
- target 不存在时会报错并列出可用 target
- `--i64-type` 在非 TypeScript target 下会被忽略

---

## 开发与测试

```bash
go test ./...
```

如果需要本地查看命令帮助：

```bash
go run ./cmd/thrift2x --help
go run ./cmd/thrift2x generate --help
```
