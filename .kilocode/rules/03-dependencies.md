# 03 · Dependency Rules

> 本文件约束各层之间的依赖关系和 import 规则，AI 在添加或修改 import 时必须遵守。

## 1. 允许的依赖方向

- Delivery → Application（服务接口）
- Application → Application 中的 Domain 逻辑（实体、验证、规则）
- Application → Ports（在 `internal/app/ports.go` 中定义的 interface）
- Infrastructure → Application 中的 Domain 类型（如 Telegram, SearchFilters） / Ports（接口）

## 2. 禁止的依赖关系（硬禁止）

- **禁止** Application 层中的 Domain 逻辑（如 `telegram.go`, `filters.go`）依赖 Infrastructure 或 Delivery：
  - Domain 逻辑文件只能依赖 Go 标准库。
- **禁止** Delivery handler 直接依赖具体的 Repository 实现、DB client、Search client 或 Cache client：
  - 只能依赖 Application 提供的服务接口。
- **禁止** Application 服务依赖具体驱动类型，例如：
  - `*sql.DB` / `sqlx.DB`
  - 具体的 Redis / Valkey 客户端
  - 具体的 Meilisearch 客户端
- **禁止** Infrastructure 反向依赖 Delivery 层（严禁 infra 调用 handler 或引用 HTTP 层类型）。

当你需要新增 import 时，先对照上面的方向判断是否会造成越层依赖，如果有歧义，优先考虑通过 interface 解耦。

## 3. 依赖注入（DI）约定

- 依赖装配集中在 Application 层（或 `internal/app/container` 一类的位置）。
- 禁止在业务代码中随意 `new` 出具体驱动实例，应由上层构造后注入。
- 禁止使用全局可变变量存放连接和服务实例（包括 DB client、NATS 连接、缓存 client 等）。

## 4. 类型共享规则

- Application 层中的 Domain 类型（如 `Telegram`, `SearchFilters`, `TimeWindow`）可以被 Infrastructure 引用。
- 不要在 Domain 逻辑外传播"技术类型"（数据库行结构、第三方 SDK 返回类型）。
- 跨层共享数据时，优先使用：
  - Application 层中的 Domain 类型（如 `app.Telegram`）；
  - 或为跨层交互专门设计的 DTO 类型。
