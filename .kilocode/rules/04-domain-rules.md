# 04 · Domain Rules

Domain 层是所有业务规则和约束的真相来源。

## 1. Domain 的职责

- 定义与航空报文/dashboard 相关的业务模型：
  - 报文（Message/Telegram）
  - 航班 / 流量实体（Flight, Sector 等）
  - 时间范围（TimeRange）
- 定义业务规则：
  - 查询时间窗口限制（例如最多 90 天）
  - 报表统计逻辑（如按时间粒度汇总）
  - 过滤与搜索条件的组合规则
- 提供便于测试的纯函数或轻量服务：
  - 例如：`func NewTimeRange(start, end time.Time) (TimeRange, error)`

## 2. Domain 中禁止的内容

- 不允许出现任何外部依赖：
  - 数据库 / SQL / ORM
  - HTTP 调用
  - Meilisearch / Redis / Valkey / NATS 客户端
- 不允许在 Domain 中做 I/O:
  - 不读写文件
  - 不访问网络
  - 不直接 log 高级信息（如 HTTP 状态）

## 3. 时间窗口与限制示例（本项目相关）

- 时间窗口限制（例如“最多查询 90 天”）应在 Domain 层以类型或函数表现：
  - 如 `TimeRange.ValidateMaxWindow(maxDays int)`。
- Application/Delivery 若需要判断时间窗口，应调用 Domain 提供的方法，而不是自行硬编码。

## 4. 业务错误与校验

- Domain 层应返回能表达业务语义的错误（例如“时间范围不合法”、“过滤条件冲突”）。
- Application/Delivery 层负责将这些错误转换成 HTTP 状态码或前端提示。
