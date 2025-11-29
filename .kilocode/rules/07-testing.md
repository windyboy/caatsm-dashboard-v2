# 07 · Testing Strategy

测试重点是保证业务规则和流式处理的稳定性，而不是追求 100% 覆盖率。

## 1. 优先级排序

1. Domain 规则测试（最高优先级）
2. Application 层用例测试（包含典型 happy path 与关键边界）
3. Infrastructure 关键路径集成测试（主查询、写入、索引）
4. Delivery 层 handler 基本行为测试（路由/绑定/状态码）

## 2. Domain 测试

- 对于时间窗口、过滤、统计等逻辑：
  - 使用纯函数或轻量服务进行单元测试
  - 覆盖至少：
    - 合法输入
    - 边界值（如恰好 90 天）
    - 明显非法输入

## 3. Application 测试

- 使用 mocks/stubs 替代具体 Repository/Adapter。
- 重点验证：
  - 正确调用 infra 接口的顺序
  - 正确处理 Domain 返回的错误
  - 正确组合多个 infra 结果（如 DB + Search + Cache）

## 4. Infrastructure / Integration 测试

- 对主 Repository 和 Search 适配器进行集成测试：
  - 使用测试数据库 / 索引
  - 验证关键查询的行为（过滤/排序/分页）

## 5. Streaming & WebSocket 测试

- 对 streaming 用例：
  - 测试分批读取逻辑是否按预期终止
  - 测试边界条件（无数据、小批量、大批量）
- 对 WebSocket：
  - 测试客户端断开、超时、慢消费等场景下的资源释放。

## 6. 测试实践约定

- 避免在测试中使用 `time.Sleep`，优先使用：
  - `context.Context`
  - `sync.WaitGroup`
  - 明确的 mock 行为
- 测试用例名称应表达具体场景：
  - `TestTimeRange_Validate_MaxWindowExceeded`
  - `TestDashboardService_StreamMessages_SlowClientCancelled`
