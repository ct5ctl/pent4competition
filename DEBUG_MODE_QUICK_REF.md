# 🐛 调试模式 - 快速参考

## 一键启动（三步骤）

### 1️⃣ 启动Mock API
```bash
cd backend
go run cmd/mock_api/main.go
```

### 2️⃣ 配置 .env
```bash
# 必需配置
COMPETITION_ENABLED=true
COMPETITION_DEBUG_MODE=true
COMPETITION_DEBUG_TARGET_IP=127.0.0.1
COMPETITION_BASE_URL=http://localhost:8000

# 可选配置
COMPETITION_DEBUG_TARGET_PORTS=8080,80,443
COMPETITION_DEBUG_CHALLENGE_CODE=debug_test
```

### 3️⃣ 启动PentAGI
```bash
cd backend
go run cmd/pentagi/main.go
```

## 配置参数速查

| 参数 | 必需 | 说明 | 示例 |
|------|------|------|------|
| `COMPETITION_ENABLED` | ✅ | 启用比赛集成 | `true` |
| `COMPETITION_DEBUG_MODE` | ✅ | 启用调试模式 | `true` |
| `COMPETITION_DEBUG_TARGET_IP` | ✅ | 测试目标IP | `127.0.0.1` |
| `COMPETITION_BASE_URL` | ✅ | Mock API地址 | `http://localhost:8000` |
| `COMPETITION_DEBUG_TARGET_PORTS` | ❌ | 目标端口 | `8080,80,443` |
| `COMPETITION_DEBUG_CHALLENGE_CODE` | ❌ | 挑战代码 | `debug_test` |

## 验证步骤

✅ **1. Mock API运行中？**
```bash
curl http://localhost:8000/health
# 返回: {"status":"ok"}
```

✅ **2. PentAGI日志正常？**
```
INFO Competition service started
INFO running in DEBUG mode, using configured target
```

✅ **3. Flow已创建？**
- 访问前端：`https://localhost:443`
- 查看Flow列表

✅ **4. FLAG已提交？**
```
=== FLAG SUBMISSION ===
Challenge: debug_test
Answer:    FLAG{...}
Status:    ✅ CORRECT
=====================
```

✅ **5. 结果文件存在？**
```bash
ls ./competition_results/
ls ./mock_api_submissions/
```

## 常见问题

### ❌ Mock API启动失败
```bash
# 更换端口
MOCK_API_PORT=9000 go run cmd/mock_api/main.go
# 更新 .env
COMPETITION_BASE_URL=http://localhost:9000
```

### ❌ 调试模式未生效
检查配置：
```bash
COMPETITION_DEBUG_MODE=true  # 必须为true
COMPETITION_DEBUG_TARGET_IP=127.0.0.1  # 必须配置
```

### ❌ FLAG未检测
- 查看Flow执行日志
- 确认AI输出包含FLAG格式
- 检查Monitor日志

## 自定义测试

### 修改正确的FLAG
编辑 `backend/cmd/mock_api/main.go`：
```go
correctFlags := map[string]string{
    "debug_test": "FLAG{your_flag}",
}
```

### 测试多个端口
```bash
COMPETITION_DEBUG_TARGET_PORTS=80,443,8080,3000
```

### 测试局域网目标
```bash
COMPETITION_DEBUG_TARGET_IP=192.168.1.100
COMPETITION_DEBUG_TARGET_PORTS=80
```

## 切换到正式模式

```bash
# .env
COMPETITION_DEBUG_MODE=false  # 关闭调试模式
COMPETITION_BASE_URL=http://real-api:8000  # 真实API
COMPETITION_TOKEN=real_token  # 真实Token
```

## 完整文档

📚 [DEBUG_MODE_GUIDE.md](DEBUG_MODE_GUIDE.md) - 完整使用指南  
📚 [QUICK_START_COMPETITION.md](QUICK_START_COMPETITION.md) - 快速启动指南  
📚 [backend/cmd/mock_api/README.md](backend/cmd/mock_api/README.md) - Mock API说明

## 测试流程图

```
┌─────────────┐
│ 启动Mock API │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  配置 .env   │ ← COMPETITION_DEBUG_MODE=true
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 启动PentAGI  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 自动创建Flow │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  AI开始测试  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  检测FLAG   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 提交到Mock  │ ← Mock API显示FLAG
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  保存结果   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  终止Flow   │
└─────────────┘
```

---

**快速获取帮助**：遇到问题？查看 [DEBUG_MODE_GUIDE.md](DEBUG_MODE_GUIDE.md) 的故障排查章节！

