# Mock Competition API Server

模拟比赛API服务器，用于调试PentAGI比赛集成功能。

## 快速启动

```bash
go run main.go
```

或指定配置：

```bash
MOCK_API_PORT=8000 MOCK_API_OUTPUT_DIR=./submissions go run main.go
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MOCK_API_PORT` | 服务器端口 | `8000` |
| `MOCK_API_OUTPUT_DIR` | 结果输出目录 | `./mock_api_submissions` |

## API端点

### GET /api/v1/challenges
获取挑战列表

```bash
curl http://localhost:8000/api/v1/challenges
```

### POST /api/v1/answer
提交FLAG答案

```bash
curl -X POST http://localhost:8000/api/v1/answer \
  -H 'Content-Type: application/json' \
  -d '{"challenge_code":"debug_test","answer":"FLAG{test_flag_12345}"}'
```

### GET /submissions
查看所有提交记录

```bash
curl http://localhost:8000/submissions
```

### GET /health
健康检查

```bash
curl http://localhost:8000/health
```

## 自定义正确的FLAG

编辑 `main.go` 中的 `correctFlags`：

```go
correctFlags := map[string]string{
    "debug_test": "FLAG{your_custom_flag}",
    "test2":      "FLAG{another_flag}",
}
```

## 输出

### 控制台

```
=== FLAG SUBMISSION ===
Challenge: debug_test
Answer:    FLAG{test_flag_12345}
Status:    ✅ CORRECT
Points:    100
=====================
```

### 文件

提交记录保存在 `MOCK_API_OUTPUT_DIR` 目录：

```json
{
  "timestamp": "2024-01-15T12:30:45Z",
  "challenge_code": "debug_test",
  "answer": "FLAG{test_flag_12345}",
  "correct": true,
  "earned_points": 100
}
```

## 配合PentAGI使用

1. 启动Mock API：`go run main.go`
2. 配置PentAGI（`.env`）：
   ```bash
   COMPETITION_ENABLED=true
   COMPETITION_DEBUG_MODE=true
   COMPETITION_DEBUG_TARGET_IP=127.0.0.1
   COMPETITION_DEBUG_TARGET_PORTS=8080
   COMPETITION_BASE_URL=http://localhost:8000
   ```
3. 启动PentAGI
4. 观察Mock API接收到的FLAG提交

## 详细文档

完整使用指南请查看：[DEBUG_MODE_GUIDE.md](../../../DEBUG_MODE_GUIDE.md)

