# 03 · Dependency Rules

本文件约束各层之间的依赖关系和导入规则。

## 1. 允许的依赖方向

- Delivery → Application
- Application → Domain
- Application → Ports（interface）
- Infrastructure → Domain（类型） / Ports（接口）

## 2. 禁止的依赖关系

- Domain 依赖 Application / Delivery / Infrastructure。
- Handler 直接依赖具体的 Repository 实现、DB client、Search client 或 Cache client。
- Application 依赖具体驱动类型，如：
  - `*sql.DB` / `sqlx.DB`
  - 具体的 Redis/Valkey 客户端
  - 具体的 Meilisearch 客户端
- Infrastructure 依赖 Delivery 层（严禁 infra 调用 handler）。

## 3. 依赖注入（DI）约定

- 依赖装配集中在 Application 层（或 `internal/app/container` 一类的位置）。
- 禁止在业务代码中随意 `new` 出具体驱动实例，应由上层构造传入。
- 禁止使用全局可变变量存放连接和服务实例。

## 4. 类型共享规则

- Domain 层的类型可以被 Application 和 Infrastructure 引用。
- 不要在 Domain 外传播“技术类型”（数据库行结构、第三方 SDK 返回类型）。
- 跨层共享数据时，优先使用 Domain 类型或专门的 DTO，而不是技术实体。
