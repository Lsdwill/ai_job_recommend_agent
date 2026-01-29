# Embedding API 修复说明

## 问题描述

初始实现时，embedding API 的请求和响应格式不正确，导致向量化失败：

```
警告：政策 就业见习补贴 向量化失败: 重试3次后仍失败: 请求失败，状态码: 422, 
响应: Failed to deserialize the JSON body into the target type: missing field `inputs`
```

## 问题原因

1. **请求字段错误**: 使用了 `input` 而不是 `inputs`
2. **响应格式错误**: 期望的是包含 `data` 字段的对象，实际返回的是向量数组

## 修复方案

### 1. 修复请求格式

**修改前** (`internal/model/policy_vector.go`):
```go
type EmbeddingRequest struct {
    Input string `json:"input"`  // ❌ 错误
}
```

**修改后**:
```go
type EmbeddingRequest struct {
    Inputs string `json:"inputs"`  // ✅ 正确
}
```

### 2. 修复响应格式

**修改前** (`internal/model/policy_vector.go`):
```go
type EmbeddingResponse []float32  // 单层数组
```

**修改后**:
```go
type EmbeddingResponse [][]float32  // 嵌套数组
```

### 3. 更新解析逻辑

**修改前** (`internal/client/embedding_client.go`):
```go
if len(embResp) == 0 {
    return nil, fmt.Errorf("返回的向量为空")
}
return embResp, nil
```

**修改后**:
```go
if len(embResp) == 0 || len(embResp[0]) == 0 {
    return nil, fmt.Errorf("返回的向量为空")
}
return embResp[0], nil  // 取第一个向量
```

## 实际 API 格式

### 请求示例
```bash
curl -X POST http://39.98.44.136:6017/emb/embed \
  -H "Content-Type: application/json" \
  -d '{"inputs":"测试文本"}'
```

### 响应示例
```json
[[
  0.010686361,
  -0.011039364,
  -0.015814532,
  0.01756029,
  ...
  (共768个浮点数)
]]
```

注意：响应是一个嵌套数组 `[[...]]`，外层数组包含一个内层数组。

## 验证方法

### 方法1: 使用 PowerShell
```powershell
$body = '{"inputs":"就业补贴政策测试"}'
Invoke-WebRequest -Uri "http://39.98.44.136:6017/emb/embed" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body `
  -UseBasicParsing | Select-Object -ExpandProperty Content
```

### 方法2: 使用测试脚本
```bash
# 运行测试脚本
test_embedding.bat
```

### 方法3: 测试完整流程
```bash
# 1. 启动服务
./qd-sc.exe -config config.yaml

# 2. 更新政策（应该成功）
curl -X POST http://localhost:8080/api/policy/update

# 3. 搜索政策
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

## 修复后的预期结果

更新政策时应该看到类似的日志：
```
2026/01/28 17:45:00 正在处理政策: 就业见习补贴
2026/01/28 17:45:01 向量化成功，维度: 768
2026/01/28 17:45:01 正在处理政策: 就业创业服务补贴
2026/01/28 17:45:02 向量化成功，维度: 768
...
2026/01/28 17:45:30 成功插入 50 条政策到向量数据库
```

## 相关文件

- `internal/model/policy_vector.go`: 数据模型定义
- `internal/client/embedding_client.go`: Embedding 客户端
- `internal/service/policy_service.go`: 政策服务
- `POLICY_VECTOR_GUIDE.md`: 完整使用指南
- `test_embedding.bat`: Embedding API 测试脚本

## 注意事项

1. **向量维度**: 确保 `config.yaml` 中的 `milvus.dimension` 设置为 768
2. **API 地址**: 确认 embedding API 地址正确且可访问
3. **超时设置**: 如果政策数量很多，可能需要增加超时时间
4. **批量处理**: 当前实现是逐条处理，每条之间间隔 100ms 避免请求过快

## 性能优化建议

如果需要处理大量政策，可以考虑：

1. **批量请求**: 修改 embedding API 支持批量输入
2. **并发处理**: 使用 goroutine 并发调用（注意控制并发数）
3. **缓存机制**: 对已处理的政策进行缓存，避免重复向量化

示例并发处理代码：
```go
// 使用 worker pool 并发处理
const workers = 5
semaphore := make(chan struct{}, workers)

for _, policy := range policies {
    semaphore <- struct{}{}
    go func(p PolicyInfo) {
        defer func() { <-semaphore }()
        // 处理政策
    }(policy)
}
```
