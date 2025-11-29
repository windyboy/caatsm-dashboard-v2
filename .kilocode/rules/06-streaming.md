# 06 · Streaming & Realtime Rules

本项目包含实时 WebSocket 以及大体量 CSV 流式导出，这些都是一等公民。

## 1. 统一抽象思想

Streaming 相关代码应尽量围绕几个统一概念设计：

- 数据源 (StreamSource)：封装从 DB / Search / 其他接口按批读取数据。
- 数据块 (StreamChunk)：单次推送/写出的一批数据行。
- 写出器 (StreamWriter)：封装向 WebSocket / HTTP Response 写入数据块的逻辑。

## 2. Delivery 层职责

Delivery 只做：

- 管理连接（WebSocket session / HTTP Response）
- basic backpressure 控制（例如限制缓冲区大小，遇到慢客户端及时中断）
- 调用 Application 层的 streaming 用例
- 序列化/反序列化（JSON / CSV 格式）

Delivery 不做：

- 决定 chunk 大小与读取策略
- 决定业务过滤和排序
- 决定导出字段和业务含义

这些属于 Application + Domain。

## 3. Application 层职责（流式用例）

- 定义 streaming 用例：
  - 如 `StreamDashboardMessages`
  - 如 `ExportMessagesAsCSV`
- 负责：
  - 根据 Domain 规则计算时间窗口、过滤条件等
  - 驱动 StreamSource 按批读取
  - 将数据转换为适合输出的 DTO

## 4. WebSocket 特别规则

- 每个 WebSocket 连接应有：
  - 上限 buffer（防止内存爆炸）
  - 合理的超时/cancel 机制（context）
- 当客户端明显过慢时：
  - 优先保证服务器稳定，必要时主动关闭连接。

## 5. CSV Streaming 导出

- 优先使用流式写出：
  - 不将完整结果集加载到内存
  - 一批一批写出行
- 当数据量非常大时：
  - 明确限制单次导出最大时间范围或行数
  - Application 层进行分页或时间分片控制
