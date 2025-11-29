# 02 · Architecture Goals & Layers

本项目采用简化后的 Clean Architecture，目标是：

1. Delivery / Application / Domain / Infrastructure 分层清晰。
2. Domain 层成为业务规则的单一真相来源。
3. Application 负责 orchestrate，不直接操作驱动。
4. Infrastructure 按业务模块归类，而不是纯按技术拆分。

## 1. 逻辑分层

- **Delivery**  
  - HTTP / WebSocket / API 层  
  - 请求解析、基础校验、响应序列化  

- **Application**  
  - 用例（Use Case）与服务（Service）  
  - 组合 Domain 与接口（ports）  
  - 控制业务流程与事务边界  

- **Domain**  
  - 业务实体（Message, Flight, TimeRange 等）  
  - 时间窗口规则、过滤规则、统计规则  
  - 不依赖任何外部技术  

- **Infrastructure**  
  - DB / Search / Cache / EventBus 的具体实现  
  - Repository / Adapter / Client 封装  

## 2. 模块化方向

- 新模块优先按照业务模块划分，而不是按技术划分：
  - `dashboard`（看板相关）
  - `search`（全文检索）
  - `sync`（数据同步处理）
  - `stats`（统计与指标）
  - `export`（导出）

## 3. 架构约束

- 不在 handler 写业务逻辑。
- 不在 infra 写业务决策，只写“如何访问外部系统”。
- 尽量将“规则”与“流程”放在 application + domain。
