# 02 · Architecture Goals & Layers

> 本文件告诉 AI：**在 caatsm-dashboard-v2 中，功能应该落在哪一层，以及各层能做什么、不能做什么。**

## 1. 总体目标

1. Delivery / Application / Infrastructure 分层清晰。
2. Domain 逻辑（业务实体、规则、验证）集成在 Application 层（`internal/app/`），成为业务规则的单一真相来源。
3. Application 负责 orchestrate，不直接操作具体驱动。
4. Infrastructure 按业务模块归类，而不是纯按技术堆在一起。

**注意**：本项目采用简化的 Clean Architecture，Domain 逻辑直接集成在 `internal/app/` 包中，而不是维护独立的 `internal/domain/` 包。这样简化结构的同时保持了清晰的关注点分离。

## 2. 逻辑分层

- **Delivery** (`internal/delivery/`)
  - HTTP / WebSocket / API 层。
  - 负责：请求解析、基础校验、路由、响应序列化。
  - 不负责：业务规则、数据库访问细节。

- **Application** (`internal/app/`)
  - 用例（Use Case）与服务（Service）。
  - **Domain 逻辑**：业务实体（Telegram, SearchFilters, TimeWindow）、业务规则、验证逻辑。
  - 端口接口（Ports）：定义 Repository, Cache, SearchIndex 等接口。
  - 负责：组合 Domain 逻辑与端口接口，控制业务流程与事务边界。
  - 不负责：SQL 细节、外部服务调用细节。

- **Infrastructure** (`internal/infrastructure/`)
  - DB / Search / Cache / EventBus 的具体实现。
  - Repository / Adapter / Client 封装。
  - 直接实现 Application 层定义的端口接口。
  - 只负责"如何访问外部系统"。

## 3. 新功能落点指引（给 AI）

当你实现一个新需求时，请按下面顺序判断：

- 只涉及路由、请求/响应体变化 → 修改 **Delivery**。
- 需要调整业务流程（先查 A 再查 B、增加步骤） → 修改或新增 **Application 服务方法**（在 `internal/app/` 中）。
- 需要新增或修改业务规则（时间窗口、过滤逻辑、统计口径） → 修改或新增 **Application 层中的 Domain 类型或函数**（如 `telegram.go`, `filters.go`, `query.go`）。
- 只变更持久化方式或外部系统用法（例如切换索引字段） → 修改 **Infrastructure**。

如果发现原有代码把业务逻辑写在 Delivery 或 Infrastructure 里，可以在用户允许的情况下，**优先帮忙往 Application 层收拢**，并保持改动局部化。

## 4. 架构约束（硬规则）

- 不在 handler 写真正的业务决策逻辑：
  - handler 最多做参数校验加调用 Application。
- 不在 Infrastructure 决定业务规则（例如：时间窗口上限、报表统计口径）。
- Application 层中的 Domain 逻辑（实体、验证、规则）不依赖 Infrastructure 或 Delivery（详见 `03-dependencies.md`）。
- 新增模块优先按照业务模块划分，而不是按技术：
  - 例如：`dashboard`、`search`、`sync`、`stats`、`export`。
- Domain 逻辑文件（如 `telegram.go`, `filters.go`）应保持纯业务逻辑，只依赖 Go 标准库。

## 5. Service Boundaries

- Split services when:
  - Public methods exceed 10 → split by functionality
  - Service handles multiple distinct domains → split by domain
- Service naming: Use `*Service` suffix (e.g., `UserService`, `OrderService`)
