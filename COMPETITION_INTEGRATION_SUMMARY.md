# PentAGI 比赛集成功能 - 实现总结

## 📋 需求回顾

根据腾讯云黑客松智能渗透挑战赛的要求，需要实现：
1. ✅ 通过API自动获取测试目标（IP和端口）
2. ✅ 自动创建Flow进行渗透测试
3. ✅ 检测AI找到的FLAG
4. ✅ 自动提交FLAG到比赛API
5. ✅ 保存结果到本地文件
6. ✅ 自动终止当前Flow，继续下一个测试
7. ✅ 所有功能由 `COMPETITION_ENABLED` 参数统一控制

## 🎯 实现方案

### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                  Competition Service                     │
│  - 定期获取挑战列表                                      │
│  - 创建Flow并启动Monitor                                 │
│  - 管理Flow生命周期                                      │
└────────────┬────────────────────────────────────────────┘
             │
     ┌───────┴────────┬────────────────────┐
     │                │                    │
     ▼                ▼                    ▼
┌─────────┐    ┌──────────────┐    ┌──────────────┐
│ Client  │    │ FlowMonitor  │    │ Flow Worker  │
│         │    │              │    │              │
│ API调用 │    │ FLAG检测     │    │ AI渗透测试   │
│ 挑战列表│    │ 自动提交     │    │              │
│ 提交答案│    │ 结果保存     │    │              │
└─────────┘    └──────────────┘    └──────────────┘
```

### 核心组件

#### 1. Client (client.go)
- **功能**：封装比赛API调用
- **方法**：
  - `GetChallenges()`: 获取挑战列表
  - `GetUnsolvedChallenges()`: 筛选未解决的挑战
  - `SubmitAnswer()`: 提交FLAG答案
  - `BuildPrompt()`: 构建渗透测试prompt

#### 2. FlowMonitor (monitor.go)
- **功能**：监控Flow执行，检测FLAG并自动提交
- **核心机制**：
  - 每2秒检查Flow的assistant logs
  - 使用正则表达式 `(?i)flag\{[^}]+\}` 匹配FLAG
  - 检测到FLAG立即提交到API
  - 提交成功后保存结果到本地JSON文件
- **关键方法**：
  - `Start()`: 启动监控
  - `checkForFlags()`: 检查logs中的FLAG
  - `extractFlags()`: 提取并规范化FLAG格式
  - `trySubmitFlag()`: 提交FLAG并处理响应
  - `saveResult()`: 保存结果到文件

#### 3. Service (service.go)
- **功能**：管理整个自动化流程
- **工作流程**：
  1. 定期从API获取挑战列表
  2. 为每个未解决的挑战创建Flow
  3. 启动FlowMonitor监控
  4. 等待FLAG被找到并提交
  5. 自动终止Flow
  6. 处理下一个挑战
- **关键方法**：
  - `processChallenges()`: 处理挑战列表
  - `processChallenge()`: 为单个挑战创建Flow和Monitor
  - `monitorFlowCompletion()`: 监控Flow完成状态，超时保护

## 📁 文件清单

### 新增文件

```
backend/pkg/competition/
├── client.go              # API客户端，封装HTTP调用
├── service.go             # 比赛服务，管理自动化流程
├── monitor.go             # Flow监控器，检测FLAG并提交
├── monitor_test.go        # 单元测试
├── README.md              # 详细文档
└── test_example.md        # 测试说明
```

### 修改文件

```
backend/
├── pkg/config/config.go   # 添加比赛配置项
└── cmd/pentagi/main.go    # 集成比赛服务启动
```

## ⚙️ 配置说明

### 正式模式配置

在 `.env` 文件中添加以下配置：

```bash
# ============= 比赛集成配置 =============
# 启用比赛集成功能
COMPETITION_ENABLED=true

# 比赛API基础URL（替换为实际地址）
COMPETITION_BASE_URL=http://x.x.x.x:8000

# 比赛API Token
COMPETITION_TOKEN=sk-aj1ok9kyZhpRx08vx31r1hJ26mm8lEjXu7on7WhAabzCFwUE

# 检查挑战列表的间隔（秒），默认60秒
COMPETITION_INTERVAL=60
```

### 调试模式配置（新增）

**在比赛API未公布前，可使用调试模式测试！**

```bash
# ============= 比赛集成配置 =============
COMPETITION_ENABLED=true
COMPETITION_BASE_URL=http://localhost:8000
COMPETITION_TOKEN=debug_token
COMPETITION_INTERVAL=60

# ============= 调试模式配置 =============
# 启用调试模式（跳过API获取挑战，使用配置的目标）
COMPETITION_DEBUG_MODE=true

# 测试目标IP
COMPETITION_DEBUG_TARGET_IP=127.0.0.1

# 测试目标端口（逗号分隔）
COMPETITION_DEBUG_TARGET_PORTS=8080,80,443

# 挑战代码（可选）
COMPETITION_DEBUG_CHALLENGE_CODE=debug_test
```

详细调试模式使用说明请查看：[DEBUG_MODE_GUIDE.md](DEBUG_MODE_GUIDE.md)

## 🔄 工作流程

### 启动流程
```
1. 主服务启动
   ↓
2. 检查 COMPETITION_ENABLED
   ↓
3. 启动 Competition Service
   ↓
4. 定期获取挑战列表（每60秒）
```

### 挑战处理流程
```
1. 从API获取未解决的挑战
   ↓
2. 选择LLM Provider（优先级：OpenAI > Anthropic > Gemini > ...）
   ↓
3. 为每个挑战：
   a. 构建prompt（包含IP、端口、FLAG要求）
   b. 创建Flow
   c. 启动FlowMonitor
   d. AI开始渗透测试
```

### FLAG检测与提交流程
```
1. FlowMonitor每2秒检查assistant logs
   ↓
2. 使用正则匹配FLAG{...}模式
   ↓
3. 发现FLAG后：
   a. 提交到比赛API
   b. 保存结果到JSON文件
   c. 通知Service
   ↓
4. Service终止Flow
   ↓
5. 继续处理下一个挑战
```

## 📊 结果文件

结果保存在 `./competition_results/` 目录：

**文件名格式**：`{timestamp}_{challenge_code}_{flow_id}.json`

**内容示例**：
```json
{
  "timestamp": "2024-01-15T12:30:45Z",
  "challenge_code": "debugdemo1",
  "flow_id": 123,
  "flag": "FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}",
  "correct": true,
  "earned_points": 270,
  "is_solved": false,
  "context": "完整的AI回复内容，包含FLAG的上下文..."
}
```

## 🎨 核心特性

### 1. FLAG检测
- **正则表达式**：`(?i)flag\{[^}]+\}`
- **支持格式**：`FLAG{...}`, `flag{...}`, `Flag{...}` 等
- **自动规范化**：统一转换为 `FLAG{...}` 格式
- **去重机制**：同一FLAG只提交一次

### 2. 自动提交
- 检测到FLAG立即提交，不等待Flow完成
- 支持同一log中多个FLAG的处理
- 错误重试和日志记录

### 3. Flow管理
- 提交成功后自动调用 `StopFlow`
- 30分钟超时保护
- 优雅关闭Monitor

### 4. 结果保存
- JSON格式，易于解析
- 包含完整上下文信息
- 自动创建目录

## 🛡️ 安全与稳定性

### 错误处理
- API调用失败不影响其他挑战
- Flow创建失败记录日志并跳过
- Monitor异常不影响主服务

### 资源管理
- Monitor使用context控制生命周期
- goroutine正确关闭
- 内存中的maps定期清理

### 日志级别
- **INFO**：服务启动/停止、挑战处理、FLAG检测、提交结果
- **DEBUG**：已处理挑战跳过、详细状态
- **ERROR**：API失败、提交失败、Flow操作失败
- **WARN**：Flow超时

## 🧪 测试

### 单元测试
```bash
cd backend/pkg/competition
go test -v
```

测试覆盖：
- FLAG提取功能
- 正则表达式匹配
- 去重和规范化

### 集成测试
参考 `test_example.md` 中的详细说明：
- Mock API Server
- 手动创建Flow测试
- 端到端测试流程

## 📈 性能指标

- **API调用频率**：遵守1次/秒限制
- **Monitor检查间隔**：2秒
- **Flow超时**：30分钟
- **挑战检查间隔**：60秒（可配置）

## 🔍 监控与调试

### 关键日志示例

**服务启动**：
```
INFO Competition service started component=competition-service
```

**发现挑战**：
```
INFO processing challenges component=competition-service count=2
INFO creating flow for challenge challenge_code=debugdemo1 target_ip=10.0.0.200
```

**FLAG检测**：
```
INFO found FLAG in assistant log flags=["FLAG{abc123}"] component=flow-monitor
INFO attempting to submit flag flag=FLAG{abc123}
```

**提交成功**：
```
INFO submitted answer correct=true earned_points=270
INFO successfully found and submitted FLAG! flag=FLAG{abc123}
INFO saved result to file file=./competition_results/...
```

**Flow终止**：
```
INFO FLAG found, stopping flow challenge_code=debugdemo1 flow_id=123
```

## 🚀 使用步骤

1. **配置环境**
   ```bash
   # 编辑 .env 文件
   COMPETITION_ENABLED=true
   COMPETITION_BASE_URL=http://实际地址:8000
   COMPETITION_TOKEN=你的token
   ```

2. **启动服务**
   ```bash
   cd backend
   go run cmd/pentagi/main.go
   ```

3. **观察日志**
   - 查看挑战获取情况
   - 监控Flow创建
   - 观察FLAG检测和提交

4. **查看结果**
   ```bash
   ls -la ./competition_results/
   cat ./competition_results/*.json
   ```

## ⚠️ 注意事项

1. **API限制**
   - 请求频率：1次/秒
   - 每题提交次数：100次
   - 遵守比赛规则

2. **依赖要求**
   - 至少配置一个LLM Provider
   - 数据库中需要用户账号
   - 确保API可访问

3. **资源使用**
   - 每个Flow占用一定资源
   - Monitor会定期检查logs
   - 定期清理结果文件

4. **控制参数**
   - `COMPETITION_ENABLED`: 统一控制开关
   - 关闭后不影响原有功能
   - 可动态调整检查间隔

## 📚 相关文档

- `backend/pkg/competition/README.md` - 详细使用文档
- `backend/pkg/competition/test_example.md` - 测试说明
- `腾讯云黑客松-智能渗透挑战赛API文档.md` - API规范

## ✅ 完成情况

- [x] API客户端实现
- [x] FLAG检测逻辑
- [x] 自动提交功能
- [x] 结果保存
- [x] Flow自动终止
- [x] 配置项管理
- [x] 完整的错误处理
- [x] 详细的日志输出
- [x] 单元测试
- [x] 文档编写

## 🎉 总结

本实现完全满足比赛要求，实现了：
1. ✅ 全自动无人工干预的渗透测试
2. ✅ 自动获取目标、测试、检测、提交、保存结果
3. ✅ 最小化修改，不影响原有功能
4. ✅ 统一的启停控制
5. ✅ 完善的错误处理和日志
6. ✅ 易于测试和调试

系统已准备好参加比赛！🏆

