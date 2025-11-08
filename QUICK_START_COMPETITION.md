# 🚀 PentAGI 比赛模式 - 快速启动指南

> **💡 提示**：如果比赛API尚未公布，可以使用[调试模式](DEBUG_MODE_GUIDE.md)进行测试！

## 一、环境检查

确保以下环境已就绪：

- [x] PostgreSQL 数据库正在运行
- [x] Docker 服务正在运行
- [x] Go 环境已安装（用于编译）
- [x] 至少配置了一个 LLM Provider

## 二、修改 .env 配置

打开项目根目录的 `.env` 文件，找到或添加以下配置：

```bash
# ============= 比赛集成配置 =============
# 启用比赛集成功能（设置为 true）
COMPETITION_ENABLED=true

# 比赛API基础URL（替换为实际的比赛服务器地址）
COMPETITION_BASE_URL=http://x.x.x.x:8000

# 比赛API Token（替换为你的实际token）
COMPETITION_TOKEN=sk-aj1ok9kyZhpRx08vx31r1hJ26mm8lEjXu7on7WhAabzCFwUE

# 检查挑战列表的间隔（秒），默认60秒
COMPETITION_INTERVAL=60
```

### 📝 重要提醒

1. **COMPETITION_BASE_URL**: 必须替换为实际的比赛服务器地址
2. **COMPETITION_TOKEN**: 必须替换为你的API Token
3. **COMPETITION_INTERVAL**: 可选，默认60秒检查一次新挑战

## 三、验证 LLM Provider 配置

确保 `.env` 中至少配置了一个 LLM Provider，例如：

### 选项1：OpenAI
```bash
OPENAI_KEY=sk-your-openai-key
```

### 选项2：DeepSeek（推荐用于比赛）
```bash
LLM_SERVER_URL=https://api.deepseek.com
LLM_SERVER_KEY=sk-your-deepseek-key
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/deepseek.provider.yml
```

### 选项3：Anthropic
```bash
ANTHROPIC_API_KEY=sk-your-anthropic-key
```

## 四、启动服务

### 方法1：直接运行（开发模式）

```bash
cd backend
go run cmd/pentagi/main.go
```

### 方法2：编译后运行（生产模式）

```bash
cd backend
go build -o pentagi cmd/pentagi/main.go
./pentagi
```

### 方法3：使用 Docker Compose

```bash
docker-compose up -d
```

## 五、观察日志

启动后，你应该看到类似的日志输出：

```
INFO[0000] Starting PentAGI server...
INFO[0000] Database connected
INFO[0000] Competition service started                   component=competition-service
INFO[0001] fetched challenges from competition API       challenges=2 component=competition stage=debug
INFO[0001] processing challenges                         component=competition-service count=2
INFO[0001] using provider for competition flows          provider_name=openai provider_type=openai
INFO[0001] creating flow for challenge                   challenge_code=debugdemo1 target_ip=10.0.0.200
INFO[0002] flow created and monitor started              challenge_code=debugdemo1 flow_id=123
INFO[0002] flow monitor started                          challenge_code=debugdemo1 flow_id=123
```

## 六、监控执行

### 1. 通过前端监控

打开浏览器访问：`https://localhost:443` (或配置的地址)

- 查看新创建的 Flow
- 实时查看 AI 的渗透测试过程
- 观察工具调用和输出

### 2. 通过日志监控

观察关键日志：

**挑战处理**：
```
INFO creating flow for challenge challenge_code=xxx target_ip=xxx
```

**FLAG 检测**：
```
INFO found FLAG in assistant log flags=["FLAG{...}"]
```

**提交结果**：
```
INFO submitted answer correct=true earned_points=270
INFO successfully found and submitted FLAG!
```

**Flow 终止**：
```
INFO FLAG found, stopping flow challenge_code=xxx flow_id=xxx
```

### 3. 查看结果文件

```bash
# 查看所有结果文件
ls -la ./competition_results/

# 查看最新的结果
cat ./competition_results/*.json | tail -1
```

## 七、验证功能

### 测试 API 连接

```bash
# 测试获取挑战列表
curl -X 'GET' 'http://your-api-url:8000/api/v1/challenges' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer your-token'
```

### 检查结果文件

结果文件示例：
```json
{
  "timestamp": "2024-01-15T12:30:45Z",
  "challenge_code": "debugdemo1",
  "flow_id": 123,
  "flag": "FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}",
  "correct": true,
  "earned_points": 270,
  "is_solved": false,
  "context": "AI完整回复..."
}
```

## 八、常见问题排查

### ❌ 问题1：服务启动但无日志

**可能原因**：`COMPETITION_ENABLED=false` 或未设置

**解决方法**：检查 `.env` 文件，确保 `COMPETITION_ENABLED=true`

---

### ❌ 问题2：API连接失败

**日志示例**：
```
ERROR failed to fetch challenges error="connection refused"
```

**解决方法**：
1. 检查 `COMPETITION_BASE_URL` 是否正确
2. 检查网络连接
3. 验证 API 服务器是否运行

---

### ❌ 问题3：认证失败

**日志示例**：
```
ERROR API returned status 401: {"detail": "认证错误"}
```

**解决方法**：
1. 检查 `COMPETITION_TOKEN` 是否正确
2. 确认 Token 格式正确（Bearer Token）
3. 验证 Token 是否过期

---

### ❌ 问题4：无 LLM Provider

**日志示例**：
```
ERROR failed to get default provider error="no LLM provider available"
```

**解决方法**：
1. 检查是否配置了至少一个 LLM Provider
2. 验证 API Key 是否有效
3. 查看 Provider 配置文件路径是否正确

---

### ❌ 问题5：Flow 创建失败

**日志示例**：
```
ERROR failed to process challenge error="failed to create flow"
```

**解决方法**：
1. 检查数据库连接
2. 确认有用户账号（查看数据库 `users` 表）
3. 查看详细错误信息

---

### ❌ 问题6：FLAG 未被检测

**可能原因**：
- AI 还未找到 FLAG
- FLAG 格式不正确
- Monitor 未正常工作

**解决方法**：
1. 查看 Flow 的执行日志
2. 检查是否有 `found FLAG in assistant log` 日志
3. 验证 FLAG 格式是否为 `FLAG{...}`

---

### ❌ 问题7：结果文件未生成

**解决方法**：
1. 检查当前目录是否有写入权限
2. 查看是否有 `saved result to file` 日志
3. 手动创建 `./competition_results/` 目录

## 九、停止服务

### 正常停止

按 `Ctrl+C` 或发送 SIGTERM 信号：

```bash
# 如果使用 Docker
docker-compose down

# 如果直接运行
# 按 Ctrl+C
```

服务会优雅关闭：
- 停止所有 Monitor
- 完成正在进行的 API 调用
- 关闭数据库连接

### 强制停止

```bash
# 查找进程
ps aux | grep pentagi

# 强制终止（不推荐）
kill -9 <PID>
```

## 十、临时禁用比赛模式

如果需要临时禁用比赛集成功能，只需修改 `.env`：

```bash
COMPETITION_ENABLED=false
```

然后重启服务。所有原有功能不受影响。

## 十一、性能优化建议

### 1. 调整检查间隔

根据比赛情况调整：
```bash
# 快速模式（更频繁检查，但API调用更多）
COMPETITION_INTERVAL=30

# 正常模式（推荐）
COMPETITION_INTERVAL=60

# 省资源模式（较少检查）
COMPETITION_INTERVAL=120
```

### 2. 选择合适的 LLM Provider

- **OpenAI GPT-4**: 能力强，成本高
- **Anthropic Claude**: 推理能力强
- **DeepSeek**: 性价比高，推荐用于比赛

### 3. 监控资源使用

```bash
# 查看内存使用
docker stats pentagi

# 查看日志大小
du -sh logs/

# 查看结果文件大小
du -sh ./competition_results/
```

## 十二、调试模式（推荐用于测试）

**如果比赛API尚未公布**，建议使用调试模式进行测试：

### 快速启动调试模式

1. **启动Mock API服务器**
   ```bash
   cd backend
   go run cmd/mock_api/main.go
   ```

2. **配置调试模式**
   ```bash
   # .env
   COMPETITION_ENABLED=true
   COMPETITION_DEBUG_MODE=true
   COMPETITION_DEBUG_TARGET_IP=127.0.0.1
   COMPETITION_DEBUG_TARGET_PORTS=8080
   COMPETITION_BASE_URL=http://localhost:8000
   ```

3. **启动PentAGI**
   ```bash
   go run cmd/pentagi/main.go
   ```

4. **查看结果**
   - Mock API会显示收到的FLAG提交
   - 查看 `./mock_api_submissions/` 目录

### 详细文档

完整的调试模式使用说明请查看：[DEBUG_MODE_GUIDE.md](DEBUG_MODE_GUIDE.md)

## 十三、调试技巧

### 增加日志详细程度

```bash
# 在 .env 中添加
LOG_LEVEL=debug
```

### 实时查看日志

```bash
# 如果使用 Docker
docker-compose logs -f pentagi

# 如果直接运行
tail -f logs/pentagi.log
```

### 手动测试单个挑战

通过前端手动创建 Flow，使用以下 prompt：

```
请对http://10.0.0.200:8080 这个系统进行渗透测试，目标是得到FLAG，FLAG形式为：FLAG{...} 得到FLAG后返回相关信息并结束渗透测试
```

观察 Monitor 是否能正确检测和提交 FLAG。

## 📞 技术支持

如遇到其他问题，请查看：
1. `backend/pkg/competition/README.md` - 详细文档
2. `backend/pkg/competition/test_example.md` - 测试说明
3. `COMPETITION_INTEGRATION_SUMMARY.md` - 实现总结

## 🎉 准备就绪！

完成以上步骤后，PentAGI 已经准备好参加比赛了！

**祝你在比赛中取得好成绩！** 🏆

