# 🐛 PentAGI 调试模式使用指南

## 📋 概述

调试模式允许你在比赛API未公布前测试PentAGI的比赛集成功能。在此模式下：
- ✅ 不需要真实的比赛API
- ✅ 可以自定义测试目标（IP和端口）
- ✅ FLAG仍会提交到指定的URL（通常是本机Mock服务器）
- ✅ 完整测试FLAG检测、提交、保存等功能

## 🚀 快速开始

### 第一步：启动Mock API服务器

Mock API服务器用于接收PentAGI提交的FLAG。

#### 1. 启动服务器

```bash
cd backend
go run cmd/mock_api/main.go
```

或指定端口和输出目录：

```bash
MOCK_API_PORT=8000 MOCK_API_OUTPUT_DIR=./mock_submissions go run cmd/mock_api/main.go
```

#### 2. 验证服务器运行

你应该看到类似输出：

```
============================================================
🚀 Mock Competition API Server
============================================================
Server:        http://localhost:8000
Output Dir:    ./mock_api_submissions

Endpoints:
  GET  /api/v1/challenges  - Get challenges
  POST /api/v1/answer      - Submit answer
  GET  /submissions        - View all submissions
  GET  /health             - Health check

Correct Flags:
  debug_test: FLAG{test_flag_12345}
============================================================

⏳ Waiting for FLAG submissions...
```

#### 3. 测试服务器

```bash
# 测试健康检查
curl http://localhost:8000/health

# 测试获取挑战
curl http://localhost:8000/api/v1/challenges

# 测试提交FLAG
curl -X POST http://localhost:8000/api/v1/answer \
  -H 'Content-Type: application/json' \
  -d '{"challenge_code":"debug_test","answer":"FLAG{test_flag_12345}"}'
```

### 第二步：配置PentAGI调试模式

编辑 `.env` 文件，添加以下配置：

```bash
# ============= 比赛集成配置 =============
# 启用比赛集成
COMPETITION_ENABLED=true

# Mock API地址（本机）
COMPETITION_BASE_URL=http://localhost:8000

# Token（调试模式可以随意设置）
COMPETITION_TOKEN=debug_token

# 检查间隔（秒）
COMPETITION_INTERVAL=60

# ============= 调试模式配置 =============
# 启用调试模式
COMPETITION_DEBUG_MODE=true

# 测试目标IP（你的测试环境）
COMPETITION_DEBUG_TARGET_IP=127.0.0.1

# 测试目标端口（逗号分隔，支持多个）
COMPETITION_DEBUG_TARGET_PORTS=8080,80,443

# 挑战代码（可选，默认为debug_test）
COMPETITION_DEBUG_CHALLENGE_CODE=debug_test
```

### 第三步：启动PentAGI

```bash
cd backend
go run cmd/pentagi/main.go
```

### 第四步：观察测试过程

#### PentAGI日志

你应该看到类似输出：

```
INFO[0000] Competition service started
INFO[0001] running in DEBUG mode, using configured target
INFO[0001] created debug challenge challenge_code=debug_test target_ip=127.0.0.1 target_ports=[8080 80 443]
INFO[0001] processing challenges count=1
INFO[0002] creating flow for challenge challenge_code=debug_test target_ip=127.0.0.1
INFO[0002] flow created and monitor started challenge_code=debug_test flow_id=123
INFO[0002] flow monitor started
```

#### Mock API日志

当PentAGI找到FLAG并提交时，Mock API会显示：

```
=== FLAG SUBMISSION ===
Challenge: debug_test
Answer:    FLAG{test_flag_12345}
Status:    ✅ CORRECT
Points:    100
=====================
```

#### 结果文件

查看以下位置的结果文件：

1. **PentAGI结果**：`./competition_results/*.json`
2. **Mock API记录**：`./mock_api_submissions/*.json`

## 📝 配置详解

### 必需配置

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `COMPETITION_ENABLED` | 启用比赛集成 | `true` |
| `COMPETITION_BASE_URL` | Mock API地址 | `http://localhost:8000` |
| `COMPETITION_DEBUG_MODE` | 启用调试模式 | `true` |
| `COMPETITION_DEBUG_TARGET_IP` | 测试目标IP | `127.0.0.1` 或 `192.168.1.100` |

### 可选配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `COMPETITION_TOKEN` | API Token | 无（调试模式可忽略） |
| `COMPETITION_INTERVAL` | 检查间隔（秒） | `60` |
| `COMPETITION_DEBUG_TARGET_PORTS` | 目标端口 | `80` |
| `COMPETITION_DEBUG_CHALLENGE_CODE` | 挑战代码 | `debug_test` |

### 端口配置说明

支持多种格式：

```bash
# 单个端口
COMPETITION_DEBUG_TARGET_PORTS=8080

# 多个端口（逗号分隔）
COMPETITION_DEBUG_TARGET_PORTS=8080,80,443

# 带空格也可以
COMPETITION_DEBUG_TARGET_PORTS=8080, 80, 443
```

## 🎯 测试场景

### 场景1：测试本地应用

测试本地运行的Web应用：

```bash
# .env配置
COMPETITION_DEBUG_MODE=true
COMPETITION_DEBUG_TARGET_IP=127.0.0.1
COMPETITION_DEBUG_TARGET_PORTS=3000
```

### 场景2：测试局域网服务器

测试局域网内的靶机：

```bash
# .env配置
COMPETITION_DEBUG_MODE=true
COMPETITION_DEBUG_TARGET_IP=192.168.1.100
COMPETITION_DEBUG_TARGET_PORTS=80,443,8080
```

### 场景3：测试Docker容器

测试Docker容器中的应用：

```bash
# .env配置
COMPETITION_DEBUG_MODE=true
COMPETITION_DEBUG_TARGET_IP=172.17.0.2
COMPETITION_DEBUG_TARGET_PORTS=8000
```

## 🔍 验证流程

### 1. 验证Mock API运行

```bash
curl http://localhost:8000/health
# 应该返回：{"status":"ok"}
```

### 2. 验证PentAGI启动

查看日志中是否有：
```
INFO Competition service started
INFO running in DEBUG mode, using configured target
```

### 3. 验证Flow创建

通过前端查看是否创建了新的Flow：
- 访问 `https://localhost:443`
- 查看Flow列表
- 应该看到一个自动创建的Flow

### 4. 验证FLAG检测

查看PentAGI日志：
```
INFO found FLAG in assistant log flags=["FLAG{...}"]
INFO attempting to submit flag
```

### 5. 验证FLAG提交

查看Mock API日志：
```
=== FLAG SUBMISSION ===
Challenge: debug_test
Answer:    FLAG{...}
Status:    ✅ CORRECT / ❌ INCORRECT
=====================
```

### 6. 验证结果保存

检查文件：
```bash
ls -la ./competition_results/
ls -la ./mock_api_submissions/
```

## 📊 Mock API端点

### GET /api/v1/challenges

获取挑战列表（返回mock数据）

```bash
curl http://localhost:8000/api/v1/challenges
```

响应：
```json
{
  "current_stage": "debug",
  "challenges": [
    {
      "challenge_code": "debug_test",
      "difficulty": "debug",
      "points": 100,
      "hint_viewed": false,
      "solved": false,
      "target_info": {
        "ip": "127.0.0.1",
        "port": [8080]
      }
    }
  ]
}
```

### POST /api/v1/answer

提交FLAG答案

```bash
curl -X POST http://localhost:8000/api/v1/answer \
  -H 'Content-Type: application/json' \
  -d '{
    "challenge_code": "debug_test",
    "answer": "FLAG{test_flag_12345}"
  }'
```

响应：
```json
{
  "correct": true,
  "earned_points": 100,
  "is_solved": false
}
```

### GET /submissions

查看所有提交记录

```bash
curl http://localhost:8000/submissions
```

响应：
```json
{
  "total": 2,
  "submissions": [
    {
      "timestamp": "2024-01-15T12:30:45Z",
      "challenge_code": "debug_test",
      "answer": "FLAG{test_flag_12345}",
      "correct": true,
      "earned_points": 100
    }
  ]
}
```

## 🎨 自定义Mock API

### 修改正确的FLAG

编辑 `backend/cmd/mock_api/main.go`：

```go
// Define correct flags for testing
correctFlags := map[string]string{
    "debug_test": "FLAG{your_custom_flag}",
    "test2":      "FLAG{another_flag}",
}
```

### 修改挑战信息

编辑 `GetChallenges` 方法：

```go
Challenges: []Challenge{
    {
        ChallengeCode: "your_code",
        Difficulty:    "medium",
        Points:        200,
        HintViewed:    false,
        Solved:        false,
        TargetInfo: TargetInfo{
            IP:   "192.168.1.100",
            Port: []int{80, 443},
        },
    },
}
```

### 修改端口

```bash
MOCK_API_PORT=9000 go run cmd/mock_api/main.go
```

然后更新 `.env`：
```bash
COMPETITION_BASE_URL=http://localhost:9000
```

## 🐛 故障排查

### ❌ Mock API无法启动

**错误**：`bind: address already in use`

**解决**：
```bash
# 查找占用端口的进程
lsof -i :8000  # macOS/Linux
netstat -ano | findstr :8000  # Windows

# 更换端口
MOCK_API_PORT=9000 go run cmd/mock_api/main.go
```

---

### ❌ PentAGI无法连接Mock API

**日志**：`connection refused`

**解决**：
1. 确认Mock API正在运行
2. 检查 `COMPETITION_BASE_URL` 配置
3. 测试连接：`curl http://localhost:8000/health`

---

### ❌ FLAG未被检测

**可能原因**：
- AI还未找到FLAG
- 测试目标不存在或无漏洞

**解决**：
1. 通过前端查看Flow执行情况
2. 查看AI的输出是否包含FLAG
3. 可以手动在Flow中输入包含FLAG的消息测试

---

### ❌ 调试模式未生效

**日志**：没有看到 "running in DEBUG mode"

**解决**：
1. 检查 `COMPETITION_DEBUG_MODE=true`
2. 检查 `COMPETITION_DEBUG_TARGET_IP` 是否配置
3. 重启服务

## 📈 性能测试

### 基准测试

1. **单次测试**：设置 `COMPETITION_INTERVAL` 较大，观察单个Flow
2. **循环测试**：设置 `COMPETITION_INTERVAL=10`，观察多次执行
3. **压力测试**：同时运行多个测试目标

### 监控指标

- Flow创建时间
- FLAG检测延迟
- 提交响应时间
- 资源使用情况

## 🔄 切换到正式模式

当比赛API公布后，只需修改配置：

```bash
# .env
# 关闭调试模式
COMPETITION_DEBUG_MODE=false

# 使用真实API
COMPETITION_BASE_URL=http://real-api-url:8000
COMPETITION_TOKEN=your_real_token
```

重启PentAGI即可，无需修改代码。

## 📚 相关文档

- `backend/cmd/mock_api/main.go` - Mock API源代码
- `backend/pkg/competition/README.md` - 比赛集成详细文档
- `QUICK_START_COMPETITION.md` - 快速启动指南

## 💡 最佳实践

1. **先测试Mock API**：确保Mock API正常工作
2. **查看日志**：密切关注PentAGI和Mock API的日志输出
3. **验证结果文件**：检查两个目录的JSON文件
4. **逐步调试**：从简单的测试目标开始
5. **保存配置**：为调试和正式模式分别保存配置文件

## 🎉 完成

现在你可以在比赛API公布前完整测试PentAGI的比赛集成功能了！

祝测试顺利！🚀

