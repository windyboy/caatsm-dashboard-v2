# 06 · Streaming & Realtime Rules

> 本项目包含实时 WebSocket 以及大体量 CSV 流式导出，这些都是一等公民。AI 在实现相关功能时必须格外注意资源和边界。

## 1. Streaming Pattern

Use Go channels for streaming operations:

- Repository methods return channels: `Stream(ctx, filters) (<-chan Entity, <-chan error)`
- Producer closes channels to signal completion
- Consumer must concurrently consume both channels
- Use `select` to handle channels and context cancellation

Example:
```go
items, errors := repo.Stream(ctx, filters)
for {
    select {
    case item, ok := <-items:
        if !ok { return }
        // process item
    case err, ok := <-errors:
        if !ok { return }
        if err != nil { return err }
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## 2. Delivery 层职责

Delivery 只做：

- 管理连接（WebSocket session / HTTP Response）。
- 基础 backpressure 控制（例如限制缓冲区大小，遇到慢客户端及时中断）。
- 调用 Application 层的 streaming 用例。
- JSON / CSV 等格式的序列化与反序列化。

Delivery 不做：

- 决定 chunk 大小与读取策略。
- 决定业务过滤和排序。
- 决定导出字段和业务含义。

这些属于 Application + Domain 的职责。

## 3. Application 层职责（流式用例）

- 定义 streaming 用例（在 `internal/app/` 的服务中），例如：
  - `DashboardService.StreamMessages()`
  - `ExportService.ExportAsCSV()`
- 负责：
  - 根据 Domain 规则（在 `internal/app/` 中，如 `filters.go`, `query.go`）计算时间窗口、过滤条件等；
  - 驱动 `StreamSource` 按批读取；
  - 将数据转换为适合输出的 DTO；
  - 确保所有 streaming 入口都接受 `context.Context` 并在取消时及时退出。

在新增 streaming 用例时，优先把"读取 + 组装 DTO"放在 Application 服务中，用简单清晰的循环结构，不要在 handler 里直接写 SQL 加写 Response。

## 4. WebSocket 特别规则

- 每个 WebSocket 连接应有：
  - 上限 buffer（防止内存爆炸）；
  - 合理的超时与 cancel 机制 (`context.Context`)。
- 当客户端明显过慢时：
  - 优先保证服务器稳定，必要时主动关闭连接；
  - 记录足够的日志方便排查，但不要刷屏。

## 5. CSV Streaming 导出

- 优先使用流式写出：
  - 不将完整结果集一次性加载到内存；
  - 一批一批写出行。
- 当数据量非常大时：
  - 明确限制单次导出最大时间范围或行数；
  - Application 层进行分页或时间分片控制；
  - 需要时在 `internal/app/` 中的 Domain 逻辑文件（如 `query.go` 或 `filters.go`）中定义"导出上限"规则，而不是在 handler 里硬编码天数。
