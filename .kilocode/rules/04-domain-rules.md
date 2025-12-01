# 04 · Domain Rules

> Domain 逻辑（业务规则和约束）位于 `internal/app/` 包中，是所有业务规则的单一真相来源。AI 在变更业务逻辑时，优先在这里落笔。

**重要说明**：本项目采用简化的架构，Domain 逻辑直接集成在 Application 层（`internal/app/`）中，而不是维护独立的 `internal/domain/` 包。Domain 相关的文件包括：`telegram.go`, `query.go`, `filters.go`, `events.go`, `errors.go` 等。

## 1. Domain 的职责（在 `internal/app/` 中）

- 定义与航空报文 / dashboard 相关的业务模型：
  - 报文（`Telegram` 在 `telegram.go`）
  - 查询和过滤（`SearchFilters`, `TimeWindow` 在 `query.go`, `filters.go`）
  - 领域事件（`TelegramReceived`, `TelegramValidated` 等在 `events.go`）
- 定义业务规则：
  - 查询时间窗口限制（例如最多 90 天）
  - 报表统计逻辑（如按时间粒度汇总）
  - 过滤与搜索条件组合规则
- 提供便于测试的纯函数或轻量服务：
  - 例如：`Telegram.Validate()`, `TimeWindow.ValidateMaxWindow()`

当你需要修改业务规则（时间、过滤、统计逻辑）时，优先在 `internal/app/` 中的 Domain 相关文件（如 `telegram.go`, `filters.go`）中新增或修改类型或函数，再让 Application 服务或 Delivery 调用。

## 2. Domain 逻辑文件中的禁止内容

Domain 逻辑文件（如 `telegram.go`, `filters.go`, `query.go`）中：

- 不允许出现任何外部依赖：
  - 数据库 / SQL / ORM
  - HTTP 调用
  - Meilisearch / Redis / Valkey / NATS 客户端
- 不允许在 Domain 逻辑中做 I/O：
  - 不读写文件
  - 不访问网络
- 日志：
  - 可以记录基础错误信息，但不要依赖 HTTP 状态码等上层概念。
- Time handling:
  - Accept `time.Time` as parameter instead of calling `time.Now()` directly
  - Example: `func (e *Entity) Validate(now time.Time) error`
  - Use zero value check for backward compatibility: `if now.IsZero() { now = time.Now() }`
  - Enables deterministic testing with fixed time values

**注意**：虽然 Domain 逻辑在 `internal/app/` 包中，但 Domain 相关的文件（如 `telegram.go`, `filters.go`）应保持纯业务逻辑，只依赖 Go 标准库。Application 服务文件（如 `dashboard.go`, `search.go`）可以依赖端口接口和基础设施。

## 3. 时间窗口与限制示例（本项目相关）

- 时间窗口限制（例如"最多查询 90 天"）应在 `internal/app/` 中的 Domain 逻辑文件（如 `query.go` 或 `filters.go`）以类型或方法表现：
  - 如：`TimeWindow.ValidateMaxWindow(maxDays int)`。
- Application 服务 / Delivery 若需要判断时间窗口是否合法，应调用 Domain 逻辑提供的方法，而不是在各处硬编码天数。

## 4. 业务错误与校验

- `internal/app/` 中的 Domain 逻辑应返回能表达业务语义的错误（定义在 `errors.go`），例如：
  - `ErrInvalidTelegram` - "报文格式不合法"
  - `ErrInvalidFilter` - "过滤条件冲突"
  - 时间范围相关的错误
- Application 服务 / Delivery 层负责将这些错误转换成：
  - HTTP 状态码；
  - 或前端可展示的错误提示文案。

如果你发现类似逻辑散落在 handler 或 Infrastructure 中（例如重复的时间窗口校验），优先建议提取到 `internal/app/` 中的 Domain 逻辑文件并复用。
