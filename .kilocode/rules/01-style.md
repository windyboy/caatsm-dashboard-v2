# 01 · Code Style & Project Conventions

本规则适用于 caatsm-dashboard-v2 的后端 Go 代码和前端 Deno/Svelte 代码。

## 1. 基础约定

- 源码统一使用 UTF-8。
- 缩进统一使用 4 个空格。
- 禁止引入与现有文件完全不同的风格（import 顺序、命名、注释风格等）。

## 2. Go 代码格式化

- 所有 Go 文件必须通过：
  - `gofmt`
  - `goimports`
- import 分组顺序：
  1. 标准库
  2. 第三方依赖
  3. 本项目内部包

## 3. 命名规则（后端）

- 避免 `util`, `helper`, `common` 这类无语义命名。
- 包名应体现业务或层次，如 `dashboard`, `search`, `sync`, `export`, `ws`。
- 结构体与函数命名：
  - 导出标识符使用 PascalCase。
  - 非导出标识符使用 camelCase。
  - 避免 `tmp`, `foo`, `bar`, `data`，改用 `msg`, `window`, `filter`, `opts` 等有含义的名字。

## 4. 前端（Deno/Svelte）代码风格

- 使用 Deno 自带格式化和 lint 工具（`deno fmt`, `deno lint`）。
- Svelte 组件命名与文件名一致，例如：
  - `Dashboard.svelte`
  - `MessagesTable.svelte`

## 5. 注释风格

- 对于导出类型/函数：必须提供 GoDoc 注释。
- 对于复杂逻辑（特别是查询逻辑、时间窗口逻辑、流式处理逻辑）：
  - 注释中说明“为什么要这样做”，而不仅仅是“做了什么”。
