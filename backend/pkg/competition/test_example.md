# Competition Integration 测试说明

## 手动测试步骤

### 1. 环境准备

确保 `.env` 文件中已配置：

```bash
# 数据库配置
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=pentagi
DATABASE_USER=pentagi
DATABASE_PASSWORD=your_password

# LLM Provider (至少配置一个)
# OpenAI
OPENAI_KEY=sk-...

# 或 DeepSeek
LLM_SERVER_URL=https://api.deepseek.com
LLM_SERVER_KEY=sk-...
LLM_SERVER_CONFIG_PATH=/path/to/deepseek.provider.yml

# 比赛配置
COMPETITION_ENABLED=true
COMPETITION_BASE_URL=http://x.x.x.x:8000
COMPETITION_TOKEN=your_api_token
COMPETITION_INTERVAL=60
```

### 2. 启动服务

```bash
cd backend
go run cmd/pentagi/main.go
```

### 3. 观察日志

启动后应该看到类似日志：

```
INFO[0000] Competition service started                   component=competition-service
INFO[0001] fetched challenges from competition API       challenges=2 component=competition stage=debug
INFO[0001] processing challenges                         component=competition-service count=2
INFO[0001] using provider for competition flows          component=competition-service provider_name=openai provider_type=openai
INFO[0001] creating flow for challenge                   challenge_code=debugdemo1 component=competition-service target_ip=10.0.0.200 target_ports=[8080]
INFO[0002] flow created and monitor started              challenge_code=debugdemo1 component=competition-service flow_id=123
INFO[0002] flow monitor started                          challenge_code=debugdemo1 component=flow-monitor flow_id=123
```

### 4. 监控Flow执行

通过前端页面或日志观察Flow执行情况。

### 5. FLAG检测

当AI找到FLAG时，应该看到：

```
INFO[0120] found FLAG in assistant log                   component=flow-monitor flags=["FLAG{abc123}"] log_id=456
INFO[0120] attempting to submit flag                     component=flow-monitor flag=FLAG{abc123}
INFO[0120] submitted answer                              challenge_code=debugdemo1 component=competition correct=true earned_points=270
INFO[0120] successfully found and submitted FLAG!        component=flow-monitor earned_points=270 flag=FLAG{abc123}
INFO[0120] saved result to file                          component=flow-monitor file=./competition_results/20240115_120000_debugdemo1_123.json
INFO[0121] FLAG found, stopping flow                     challenge_code=debugdemo1 component=competition-service flag=FLAG{abc123} flow_id=123
```

### 6. 验证结果文件

检查 `./competition_results/` 目录：

```bash
ls -la ./competition_results/
cat ./competition_results/20240115_120000_debugdemo1_123.json
```

应该看到类似内容：

```json
{
  "timestamp": "2024-01-15T12:00:00Z",
  "challenge_code": "debugdemo1",
  "flow_id": 123,
  "flag": "FLAG{abc123}",
  "correct": true,
  "earned_points": 270,
  "is_solved": false,
  "context": "经过分析，发现目标系统存在SQL注入漏洞...\n成功获取FLAG: FLAG{abc123}"
}
```

## 单元测试

### 测试FLAG提取

创建测试文件 `backend/pkg/competition/monitor_test.go`：

```go
package competition

import (
	"testing"
)

func TestExtractFlags(t *testing.T) {
	monitor := &FlowMonitor{}
	
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple flag",
			input:    "Found the flag: FLAG{abc123}",
			expected: []string{"FLAG{abc123}"},
		},
		{
			name:     "lowercase flag",
			input:    "The flag is flag{xyz789}",
			expected: []string{"FLAG{xyz789}"},
		},
		{
			name:     "multiple flags",
			input:    "FLAG{abc} and FLAG{xyz}",
			expected: []string{"FLAG{abc}", "FLAG{xyz}"},
		},
		{
			name:     "no flag",
			input:    "No flag found",
			expected: nil,
		},
		{
			name:     "uuid flag",
			input:    "FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}",
			expected: []string{"FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.extractFlags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d flags, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected flag %s, got %s", expected, result[i])
				}
			}
		})
	}
}
```

运行测试：

```bash
cd backend/pkg/competition
go test -v
```

## API测试

### 测试获取挑战

```bash
curl -X 'GET' 'http://x.x.x.x:8000/api/v1/challenges' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer your_token'
```

### 测试提交答案

```bash
curl -X 'POST' 'http://x.x.x.x:8000/api/v1/answer' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer your_token' \
  -H 'Content-Type: application/json' \
  -d '{
    "challenge_code": "debugdemo1",
    "answer": "FLAG{test123}"
  }'
```

## 模拟测试

如果没有实际的比赛环境，可以：

1. **Mock API Server**：创建一个简单的HTTP服务器模拟比赛API
2. **手动创建Flow**：通过前端手动创建Flow，在prompt中包含已知的FLAG
3. **测试Monitor**：观察Monitor是否能正确检测和提交FLAG

### Mock API Server示例

```go
// test_server.go
package main

import (
	"encoding/json"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/api/v1/challenges", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"current_stage": "debug",
			"challenges": []map[string]interface{}{
				{
					"challenge_code": "test1",
					"difficulty":     "easy",
					"points":         100,
					"hint_viewed":    false,
					"solved":         false,
					"target_info": map[string]interface{}{
						"ip":   "127.0.0.1",
						"port": []int{8080},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	})
	
	http.HandleFunc("/api/v1/answer", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		
		correct := req["answer"] == "FLAG{test123}"
		response := map[string]interface{}{
			"correct":       correct,
			"earned_points": 100,
			"is_solved":     false,
		}
		json.NewEncoder(w).Encode(response)
	})
	
	log.Println("Mock server started on :8000")
	http.ListenAndServe(":8000", nil)
}
```

运行Mock服务器：

```bash
go run test_server.go
```

## 常见问题

### Q1: Monitor没有检测到FLAG
**A**: 检查以下几点：
- FLAG格式是否为 `FLAG{...}` 或 `flag{...}`
- Monitor是否正常启动（查看日志）
- Flow是否有新的assistant log产生

### Q2: 提交失败
**A**: 
- 检查API地址和Token是否正确
- 检查网络连接
- 查看API返回的错误信息

### Q3: Flow没有自动终止
**A**:
- 检查 `StopFlow` 是否被调用（查看日志）
- 检查Flow状态
- 可能需要手动停止

### Q4: 结果文件没有生成
**A**:
- 检查 `./competition_results/` 目录权限
- 查看Monitor日志中的错误信息
- 检查磁盘空间

## 调试技巧

1. **增加日志级别**：设置 `LOG_LEVEL=debug` 获取更详细的日志
2. **查看数据库**：直接查询assistant_logs表，检查是否有FLAG
3. **断点调试**：在关键函数设置断点，逐步调试
4. **模拟FLAG**：手动在数据库中插入包含FLAG的log，测试Monitor检测功能

## 性能监控

监控以下指标：
- API调用频率（不超过1次/秒）
- Flow执行时间（平均/最大）
- Monitor检查延迟（应该在2秒内）
- 内存使用（特别是monitors map）
- 磁盘使用（结果文件累积）

