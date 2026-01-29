# 政策向量化功能最终测试指南

## 修复历史

### 问题 1: API 字段错误
- **错误**: `missing field 'inputs'`
- **原因**: 使用了 `input` 而不是 `inputs`
- **修复**: 将请求字段改为 `inputs`

### 问题 2: Token 限制
- **错误**: `inputs must have less than 512 tokens. Given: 1600`
- **原因**: 政策文本太长，超过 512 tokens 限制
- **修复**: 实现双内容策略，精简内容用于向量化，完整内容用于展示

### 问题 3: 响应格式错误
- **错误**: `cannot unmarshal array into Go value of type float32`
- **原因**: API 返回嵌套数组 `[[...]]` 而不是单层数组 `[...]`
- **修复**: 将响应类型改为 `[][]float32` 并取第一个元素

## 最终正确的 API 格式

### 请求
```json
{
  "inputs": "政策文本内容（< 512 tokens）"
}
```

### 响应
```json
[[0.010686361, -0.011039364, ..., 0.00836296]]
```

注意：
- 外层是数组（支持批量请求）
- 内层是 768 维向量
- 我们取 `response[0]` 作为结果

## 测试步骤

### 1. 编译项目
```bash
go build -o qd-sc.exe ./cmd/server
```

### 2. 启动服务
```bash
./qd-sc.exe -config config.yaml
```

预期输出：
```
2026/01/28 17:45:00 服务器启动在 0.0.0.0:8080
2026/01/28 17:45:00 OpenAI兼容端点: POST http://0.0.0.0:8080/v1/chat/completions
```

### 3. 测试健康检查
```bash
curl http://localhost:8080/health
```

预期输出：
```json
{"status":"ok"}
```

### 4. 更新政策到向量数据库
```bash
curl -X POST http://localhost:8080/api/policy/update
```

预期日志输出：
```
正在处理政策: 就业见习补贴
向量化成功，维度: 768
正在处理政策: 就业创业服务补贴
向量化成功，维度: 768
...
成功更新 50 条政策到向量数据库
```

**注意**：
- 这个过程可能需要 5-10 分钟（取决于政策数量）
- 每条政策之间有 100ms 延迟，避免请求过快
- 如果某些政策失败，会显示警告但继续处理其他政策

### 5. 测试政策搜索
```bash
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

预期输出：
```json
{
  "query": "就业补贴",
  "results": [
    {
      "ID": "1988473569041494018",
      "Content": "政策名称：就业见习补贴\n政策级别：自治区级\n...",
      "Distance": 0.15
    },
    {
      "ID": "1988503056181407745",
      "Content": "政策名称：就业创业服务补贴\n...",
      "Distance": 0.23
    },
    ...
  ]
}
```

### 6. 测试对话接口
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {
        "role": "user",
        "content": "我想了解就业见习补贴政策"
      }
    ]
  }'
```

预期：AI 助手会自动调用 `queryPolicy` 工具并返回相关政策信息。

## 验证清单

- [ ] 服务成功启动
- [ ] 健康检查返回 OK
- [ ] 政策更新成功（无 token 限制错误）
- [ ] 至少成功向量化 80% 的政策
- [ ] 搜索返回相关结果
- [ ] 对话接口能够识别政策查询意图
- [ ] 返回的政策内容完整（包含所有字段）

## 常见问题

### Q1: 更新时仍然报 token 限制错误
**A**: 检查 `buildPolicyContent()` 方法是否正确截断了长文本。可以添加日志查看实际内容长度：
```go
fmt.Printf("政策 %s 内容长度: %d 字符\n", policy.Zcmc, len(content))
```

### Q2: 搜索结果不准确
**A**: 可能原因：
1. 向量化的内容太精简，丢失了关键信息
2. 查询关键词不够精确
3. topK 设置太小

解决方法：
- 调整 `buildPolicyContent()` 中的截断长度
- 使用更具体的查询关键词
- 增加 topK 数量

### Q3: Milvus 连接失败
**A**: 检查：
1. Milvus 服务是否运行：`telnet 39.98.44.136 6012`
2. 网络是否可达
3. 配置文件中的地址和端口是否正确

### Q4: 部分政策向量化失败
**A**: 这是正常的，可能原因：
1. 某些政策内容特殊字符导致 API 解析失败
2. 网络临时波动
3. API 限流

只要成功率 > 80% 就可以接受。

## 性能优化建议

### 1. 批量处理
如果 embedding API 支持批量请求，可以修改为批量处理：
```go
// 每 10 条政策一批
batchSize := 10
for i := 0; i < len(policies); i += batchSize {
    batch := policies[i:min(i+batchSize, len(policies))]
    vectors := s.embeddingClient.GetEmbeddingBatch(batch)
    // ...
}
```

### 2. 并发处理
使用 goroutine 并发调用（注意控制并发数）：
```go
const workers = 5
semaphore := make(chan struct{}, workers)
var wg sync.WaitGroup

for _, policy := range policies {
    wg.Add(1)
    semaphore <- struct{}{}
    
    go func(p PolicyInfo) {
        defer wg.Done()
        defer func() { <-semaphore }()
        
        content := s.buildPolicyContent(p)
        vector, _ := s.embeddingClient.GetEmbedding(content)
        // 存储结果
    }(policy)
}

wg.Wait()
```

### 3. 增量更新
只更新新增或修改的政策：
```go
// 检查政策是否已存在
exists := s.milvusClient.Exists(ctx, policy.ID)
if exists {
    // 检查是否需要更新（比较更新时间）
    continue
}
// 插入新政策
```

## 监控建议

### 1. 添加日志
```go
log.Printf("开始更新政策，总数: %d", len(policies))
log.Printf("成功向量化: %d, 失败: %d", successCount, failCount)
log.Printf("插入 Milvus 耗时: %v", duration)
```

### 2. 添加指标
```go
// 记录成功率
successRate := float64(successCount) / float64(len(policies)) * 100
log.Printf("向量化成功率: %.2f%%", successRate)

// 记录平均耗时
avgTime := totalTime / time.Duration(successCount)
log.Printf("平均每条政策耗时: %v", avgTime)
```

### 3. 错误收集
```go
var failedPolicies []string
for _, policy := range policies {
    if err := process(policy); err != nil {
        failedPolicies = append(failedPolicies, policy.Zcmc)
    }
}
log.Printf("失败的政策: %v", failedPolicies)
```

## 下一步

1. **定期更新**: 设置定时任务每周更新一次政策
2. **监控告警**: 当成功率 < 80% 时发送告警
3. **用户反馈**: 收集用户对搜索结果的反馈，优化向量化策略
4. **A/B 测试**: 测试不同的内容截断策略，找到最优方案

## 相关文档

- `EMBEDDING_FIX.md`: API 格式修复详情
- `EMBEDDING_TOKEN_LIMIT.md`: Token 限制解决方案
- `POLICY_VECTOR_GUIDE.md`: 完整使用指南
- `POLICY_FEATURE.md`: 功能说明
