# 05 · Infrastructure Rules

Infrastructure 层负责“如何访问外部系统”，而不负责“业务规则”。

## 1. 模块划分原则

- 优先按业务模块划分，而不是按技术：
  - `infrastructure/dashboard`
  - `infrastructure/sync`
  - `infrastructure/stats`
  - `infrastructure/search`
  - `infrastructure/export`
- 对于多个业务模块共享的底层能力（例如 PostgreSQL 会话管理），可以抽成更底层的内部包，但不应把所有东西都丢进一个技术命名包。

## 2. 职责边界

基础设施实现主要负责：

- 数据持久化（PostgreSQL/TimescaleDB）
- 搜索索引（Meilisearch）
- 缓存（Valkey/Redis）
- 消息队列 / 事件总线（NATS / JetStream）
- 与外部系统的 API 调用

不负责：

- 决定业务流程和规则（时间窗口、过滤策略、统计口径）
- 决定用户能看到哪些数据（这是 Application/Domain 的事情）

## 3. Repository / Adapter 约定

- Repository / Adapter 必须实现 ports 接口。
- 实现名称应明确对应的业务模块，例如：
  - `DashboardRepository`
  - `SearchRepository`
  - `SyncEventStore`
- Repository 方法应避免泄露底层实现细节（例如返回 DB 专用结构体）。

## 4. 事务和资源管理

- Infrastructure 层负责：
  - 事务开启/提交/回滚
  - 连接池配置
  - 超时/cancel 与 context 处理
- Application 层负责：
  - 定义一个 use case 是否需要事务（由 Application 调用合适的 infra API）。
