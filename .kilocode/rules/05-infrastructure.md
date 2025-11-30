# 05 · Infrastructure Rules

> Infrastructure 层只负责“如何访问外部系统”，不负责决定“业务规则”。

## 1. 模块划分原则

- 当前 Infrastructure 按技术栈划分（符合实际代码结构）：
  - `infrastructure/persistence/` - PostgreSQL 实现
  - `infrastructure/search/` - Meilisearch 实现
  - `infrastructure/cache/` - Valkey/Redis 实现
  - `infrastructure/streaming/` - NATS 实现
  - `infrastructure/event/` - 事件总线实现
  - `infrastructure/ws/` - WebSocket hub 实现
- 对于多个业务模块共享的底层能力，可以抽成更底层内部包：
  - 例如 `infrastructure/persistence/internal/session`
  - 但不要建一个什么都往里扔的"大杂烩"包。

当你新增 Repository / Adapter 时，优先选择对应的技术栈目录，确保直接实现 `internal/app/ports.go` 中定义的接口。

## 2. 职责边界

基础设施实现主要负责：

- 数据持久化（PostgreSQL / TimescaleDB）
- 搜索索引（Meilisearch）
- 缓存（Valkey / Redis）
- 消息队列 / 事件总线（NATS / JetStream）
- 与外部系统的 API 调用

不负责：

- 决定业务流程和规则（时间窗口、过滤策略、统计口径）
- 决定用户能看到哪些数据（这是 Application / Domain 的责任）

## 3. Repository / Adapter 约定

- Repository / Adapter 必须实现 ports 接口（在 `internal/app/ports.go` 中定义）。
- 实现名称应明确对应的业务模块，例如：
  - `PostgresStore`（实现 `app.Repository`）
  - `MeilisearchIndex`（实现 `app.SearchIndex`）
  - `ValkeyStore`（实现 `app.Cache`）
- Repository 方法应避免泄露底层实现细节：
  - 不要直接返回数据库行结构体或第三方 SDK 类型；
  - 返回 Application 层中的 Domain 类型（如 `app.Telegram`）或专门定义的 DTO。

## 4. 事务和资源管理

- Infrastructure 层负责：
  - 事务开启 / 提交 / 回滚；
  - 连接池配置；
  - 与连接/会话相关的超时 / cancel 处理。
- Application 层负责：
  - 定义某个 use case 是否需要事务；
  - 决定在一个 use case 中需要执行哪些仓储操作，由 infra 提供相应 API。
