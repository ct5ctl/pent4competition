# Competition Integration

这个包实现了PentAGI与腾讯云黑客松智能渗透挑战赛的自动集成功能。

## 功能说明

- **自动获取挑战**：从比赛API自动获取挑战列表
- **自动创建Flow**：为每个未解决的挑战自动创建Flow并开始渗透测试
- **顺序处理**：按顺序处理挑战，避免并行测试
- **智能提示词**：使用优化的prompt模板，明确告知AI目标和FLAG格式
- **实时监控**：监控Flow执行过程，自动检测AI回复中的FLAG模式
- **自动提交**：检测到FLAG后自动提交到比赛API验证
- **自动终止**：提交成功后自动终止当前Flow，继续处理下一个挑战
- **结果保存**：所有提交结果（包括FLAG、得分、上下文）自动保存到本地JSON文件

## 配置说明

在 `.env` 文件中添加以下配置项：

```bash
# 启用比赛集成功能
COMPETITION_ENABLED=true

# 比赛API基础URL（需要替换为实际的比赛服务器地址）
COMPETITION_BASE_URL=http://x.x.x.x:8000

# 比赛API Token
COMPETITION_TOKEN=your_api_token_here

# 检查挑战列表的间隔（秒），默认60秒
COMPETITION_INTERVAL=60
```

## 工作流程

### 1. 服务启动
- 如果 `COMPETITION_ENABLED=true`，比赛服务会自动启动
- 服务会定期（根据 `COMPETITION_INTERVAL` 配置）从比赛API获取挑战列表

### 2. 挑战处理
对于每个未解决的挑战（`solved: false`），系统会：
1. 构建包含目标IP和端口的prompt
2. 使用第一个可用的LLM Provider创建Flow
3. 启动FlowMonitor监控Flow执行
4. AI开始自动渗透测试

### 3. FLAG检测与提交
FlowMonitor会：
1. 每2秒检查一次Flow的assistant logs
2. 使用正则表达式匹配 `FLAG{...}` 或 `flag{...}` 模式
3. 检测到FLAG后立即提交到比赛API
4. 保存提交结果到本地文件（JSON格式）

### 4. Flow终止
提交成功后：
1. 自动调用 `StopFlow` 终止当前Flow
2. 停止FlowMonitor
3. 继续处理下一个挑战

### 5. 超时保护
- 每个Flow最多运行30分钟
- 超时后自动终止，避免资源浪费

## 文件说明

### client.go
- `Client`: HTTP客户端，封装比赛API调用
- `GetChallenges()`: 获取挑战列表
- `GetUnsolvedChallenges()`: 获取未解决的挑战
- `SubmitAnswer()`: 提交答案
- `BuildPrompt()`: 构建prompt模板

### monitor.go
- `FlowMonitor`: Flow监控器，检测FLAG并自动提交
- `Start()`: 启动监控
- `checkForFlags()`: 检查assistant logs中的FLAG
- `extractFlags()`: 使用正则提取FLAG
- `trySubmitFlag()`: 提交FLAG并保存结果
- `saveResult()`: 保存结果到本地JSON文件

### service.go
- `Service`: 比赛服务，管理整个自动化流程
- `Start()`: 启动服务
- `processChallenges()`: 处理挑战列表
- `processChallenge()`: 为单个挑战创建Flow和Monitor
- `monitorFlowCompletion()`: 监控Flow完成状态

## 结果文件

所有提交结果保存在 `./competition_results/` 目录下，格式如下：

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "challenge_code": "debugdemo1",
  "flow_id": 123,
  "flag": "FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}",
  "correct": true,
  "earned_points": 270,
  "is_solved": false,
  "context": "AI回复的完整内容，包含FLAG的上下文"
}
```

文件命名格式：`{timestamp}_{challenge_code}_{flow_id}.json`

## 注意事项

1. **LLM Provider配置**：确保至少配置了一个LLM Provider（OpenAI、Anthropic、Gemini等）
2. **用户账号**：确保数据库中有至少一个用户账号
3. **API可访问性**：比赛API必须可访问，且Token有效
4. **顺序执行**：服务按顺序处理挑战，不会并行执行
5. **去重机制**：每个challenge在本次运行中只会处理一次（通过内存中的processed map追踪）
6. **重启行为**：如果服务重启，会重新检查所有挑战
7. **已解决挑战**：如果挑战在API中标记为已解决（`solved: true`），将不会被处理
8. **结果目录**：确保有权限创建 `./competition_results/` 目录

## API接口

### 1. 获取挑战列表

```bash
GET http://x.x.x.x:8000/api/v1/challenges
Authorization: Bearer {COMPETITION_TOKEN}
```

响应格式：
```json
{
  "current_stage": "debug",
  "challenges": [
    {
      "challenge_code": "debugdemo1",
      "difficulty": "medium",
      "points": 300,
      "hint_viewed": false,
      "solved": false,
      "target_info": {
        "ip": "10.0.0.200",
        "port": [8080]
      }
    }
  ]
}
```

### 2. 提交答案

```bash
POST http://x.x.x.x:8000/api/v1/answer
Authorization: Bearer {COMPETITION_TOKEN}
Content-Type: application/json

{
  "challenge_code": "debugdemo1",
  "answer": "FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}"
}
```

响应格式：
```json
{
  "correct": true,
  "earned_points": 270,
  "is_solved": false
}
```

## 日志

比赛服务的日志会包含以下信息：

### 服务级别
- `INFO`: 服务启动/停止、挑战处理开始/结束
- `DEBUG`: 无新挑战、已处理挑战跳过
- `ERROR`: API调用失败、Flow创建失败

### Monitor级别
- `INFO`: Monitor启动/停止、FLAG检测、提交成功/失败、结果保存
- `ERROR`: 日志获取失败、提交失败、文件保存失败

### Flow管理级别
- `INFO`: Flow创建、Flow终止、超时处理
- `WARN`: Flow超时
- `ERROR`: Flow操作失败

## 故障排查

### 1. 服务未启动
- 检查 `COMPETITION_ENABLED=true`
- 检查 `COMPETITION_BASE_URL` 和 `COMPETITION_TOKEN` 是否配置

### 2. 无法获取挑战
- 检查API地址是否正确
- 检查Token是否有效
- 检查网络连接

### 3. Flow创建失败
- 检查是否配置了LLM Provider
- 检查Provider配置是否正确
- 检查数据库中是否有用户

### 4. FLAG未被检测
- 检查AI是否真的找到了FLAG
- 检查FLAG格式是否为 `FLAG{...}` 或 `flag{...}`
- 查看Monitor日志

### 5. 提交失败
- 检查API是否可访问
- 检查提交的FLAG格式是否正确
- 检查是否超过提交次数限制（100次/题）

## 性能优化建议

1. **调整检查间隔**：根据API限制调整 `COMPETITION_INTERVAL`，避免频繁请求
2. **Monitor检查频率**：目前每2秒检查一次logs，可根据需要调整
3. **超时时间**：默认30分钟，可根据挑战难度调整
4. **结果目录**：定期清理旧的结果文件，避免占用过多磁盘空间
