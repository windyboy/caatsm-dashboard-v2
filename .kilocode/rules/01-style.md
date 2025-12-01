# 01 · Code Style & Project Conventions

> 面向 Kilo / 其他 AI 编程助手：**当你在 caatsm-dashboard-v2 中写代码时，必须遵守下面的风格约定。**

## 1. 基础约定

- 源码统一使用 UTF-8。
- 非 Go 代码（前端、脚本、文档中的示例）统一使用 4 个空格缩进。
- Go 代码缩进以 `gofmt` 默认为准（tab 缩进），不要在示例中强行用空格替代。
- 修改文件时：
  - 尽量只改与当前任务相关的行，避免为了“变好看”整体重排。
  - 除非用户明确要求，不要大范围重命名或重构。

## 2. Go 代码格式化

- 所有 Go 文件必须符合：
  - `gofmt`
  - `goimports`
- import 分组顺序（从上到下）：
  1. 标准库
  2. 第三方依赖
  3. 本项目内部包
- 生成代码时，直接按上述分组生成 import，不要引入未使用的依赖。

## 3. 命名规则（后端）

- 避免 `util`、`helper`、`common` 这类无语义包名。
- 包名应体现业务或层次，例如：
  - `dashboard`、`search`、`sync`、`export`、`ws`。
- 结构体与函数命名：
  - 导出标识符使用 PascalCase。
  - 非导出标识符使用 camelCase。
- 变量命名：
  - 避免 `tmp`、`foo`、`bar`、`data`。
  - 使用有实际含义的名字，例如：`msg`、`window`、`filter`、`opts`、`limit`、`offset`。
  - `ctx` 专门保留给 `context.Context`。

## 4. 前端 / 脚本（如 Deno / Svelte）代码风格

- 使用 Deno 自带工具保证风格一致：
  - `deno fmt`
  - `deno lint`
- Svelte 组件：文件名和组件名保持一致，例如：
  - `Dashboard.svelte`
  - `MessagesTable.svelte`
- 组件职责尽量单一：
  - 一个组件不要同时承担页面布局和复杂业务逻辑。
  - 复杂逻辑拆到独立的 store 或工具模块。

## 5. 注释风格

- 导出类型、接口和函数：必须提供简短 GoDoc 注释，说明用途和关键行为。
- 对于复杂逻辑（尤其是查询、时间窗口、流式处理）：
  - 注释说明"为什么要这样做"（业务原因、边界条件），而不是机械重复"做了什么"。
- 临时调试代码使用完后应删除，不要残留到最终提交。

## 6. Error Handling

- Wrap errors with `%w` verb: `fmt.Errorf("operation failed: %w", err)`
- Domain errors return as-is, do not wrap
- Infrastructure errors must be wrapped with operation context
- Application services wrap infrastructure errors with business context

Examples:
```go
// Domain error - return directly
return ErrInvalidInput{Field: "id", Reason: "cannot be empty"}

// Infrastructure error - wrap with operation context
if err := db.Query(ctx, query); err != nil {
    return fmt.Errorf("database query failed: %w", err)
}
```
