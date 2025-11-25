# 测试指南

本文档描述如何测试 Deno + Svelte 前端改造后的功能。

## 前置条件

1. 确保所有依赖服务正在运行：
   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

2. 启动 Go 后端：
   ```bash
   make dev
   # 或
   ./bin/caatsm -config config/config.local.toml
   ```

3. 启动前端开发服务器：
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## WebSocket 测试

### 手动测试步骤

1. **打开 Dashboard 页面**
   - 访问 `http://localhost:5173`
   - 应该看到 "Live Stream" 组件和统计卡片

2. **检查 WebSocket 连接**
   - 打开浏览器开发者工具 (F12)
   - 切换到 Network 标签
   - 过滤 "WS" (WebSocket)
   - 应该看到到 `ws://localhost:3002/ws` 的连接

3. **验证实时消息接收**
   - 如果有消息通过 NATS 流入系统
   - 应该能在 Live Stream 中看到新消息出现
   - 统计数字应该实时更新

4. **测试连接重连**
   - 在开发者工具的 Console 中，应该看到 "WebSocket connected" 日志
   - 如果断开连接，应该看到重连尝试的日志

### 自动化测试

运行 Playwright 测试：

```bash
cd frontend
npm run test
```

## REST API 测试

### 手动测试步骤

1. **测试搜索功能**
   - 访问 `http://localhost:5173/search`
   - 输入搜索关键词
   - 点击 "Search" 按钮
   - 应该看到搜索结果

2. **测试自动补全**
   - 在搜索框中输入至少 2 个字符
   - 应该看到自动补全建议下拉菜单

3. **测试统计接口**
   - 访问 Dashboard 页面
   - 统计卡片应该显示数据
   - 可以通过浏览器开发者工具的 Network 标签查看 API 请求

4. **测试导出功能**
   - 在搜索页面执行搜索
   - 应该有导出选项（如果实现了）

### API 端点测试

使用 curl 或 Postman 测试 API：

```bash
# 测试搜索
curl "http://localhost:3002/api/search?query=test"

# 测试统计总数
curl "http://localhost:3002/api/stats/total"

# 测试优先级统计
curl "http://localhost:3002/api/stats/priority"

# 测试类型统计
curl "http://localhost:3002/api/stats/type"

# 测试自动补全
curl "http://localhost:3002/api/autocomplete?term=test&size=5"
```

所有 API 应该返回 JSON 格式的响应。

## WebSocket 消息格式测试

### 消息类型

WebSocket 消息应该遵循以下格式：

```json
{
  "type": "message|stats-total|stats-priority|stats-type",
  "data": { ... }
}
```

### 测试消息接收

在浏览器 Console 中运行：

```javascript
// 连接到 WebSocket
const ws = new WebSocket('ws://localhost:3002/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

## 集成测试

### 完整流程测试

1. **启动所有服务**
   ```bash
   # 终端 1: 启动依赖服务
   docker compose -f docker-compose.dev.yml up -d
   
   # 终端 2: 启动 Go 后端
   make dev
   
   # 终端 3: 启动前端
   cd frontend && npm run dev
   ```

2. **测试实时数据流**
   - 如果有消息发布工具，发布一些测试消息到 NATS
   - 观察 Dashboard 是否实时更新
   - 检查统计数字是否正确更新

3. **测试搜索和实时更新的组合**
   - 在搜索页面执行搜索
   - 同时观察 Dashboard 的实时更新
   - 两者应该互不干扰

## 性能测试

### WebSocket 连接数

测试多个客户端同时连接：

```bash
# 使用多个浏览器标签或窗口
# 每个标签都应该能独立接收消息
```

### 消息处理性能

- 观察大量消息流入时的性能
- 检查前端是否限制消息列表长度（应该最多 50 条）
- 检查内存使用情况

## 错误处理测试

### 网络断开

1. 断开网络连接
2. WebSocket 应该尝试重连
3. 恢复网络后应该自动重连成功

### 后端服务停止

1. 停止 Go 后端服务
2. WebSocket 应该检测到断开
3. 应该显示重连尝试
4. 重启后端后应该自动重连

## 浏览器兼容性

测试以下浏览器：
- Chrome/Edge (Chromium)
- Firefox
- Safari

## 已知问题

- 如果遇到类型错误，可能是 TypeScript 配置问题，不影响运行时功能
- WebSocket 重连可能需要几秒钟

## 故障排除

### WebSocket 连接失败

1. 检查 Go 后端是否运行在 `localhost:3002`
2. 检查防火墙设置
3. 检查浏览器控制台的错误信息

### API 请求失败

1. 检查 Vite 代理配置
2. 检查 CORS 设置
3. 检查后端日志

### 前端构建失败

1. 运行 `npm install` 重新安装依赖
2. 运行 `npm run check` 检查类型错误
3. 清除 `.svelte-kit` 目录后重试

## WebSocket Handler Tests

### Test File Location
`internal/handlers/websocket_test.go`

### Test Cases Covered
- Valid telegram_processed event handling (data field format)
- Valid telegram_processed event handling (top-level telegram format)
- Invalid JSON event handling
- Event missing type field
- Non-telegram_processed event (stats_update)
- Stats service error handling
- WebSocketMessage JSON serialization

### Running Tests
```bash
go test ./internal/handlers/... -v
```

## EventBroadcaster Tests

### Test File Location
`internal/handlers/broadcaster_test.go`

### Test Cases Covered
- Subscribe/Unsubscribe functionality
- Broadcast to multiple clients
- Channel full handling (capacity 10)
- Close functionality (all channels closed)
- PublishStatsUpdate with nil Redis client
- JSON validation

### Running Tests
```bash
go test ./internal/handlers/... -v
```
