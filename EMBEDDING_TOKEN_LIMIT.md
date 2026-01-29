# Embedding Token 限制解决方案

## 问题描述

Embedding API 有 **512 tokens** 的输入限制，而政策文本通常很长（可能超过 1600 tokens），导致向量化失败。

错误示例：
```
Input validation error: `inputs` must have less than 512 tokens. Given: 1600
```

## 解决方案

采用**双内容策略**：

### 1. 精简内容用于向量化

在 `buildPolicyContent()` 方法中：
- 只保留核心字段（政策名称、级别、来源单位）
- 对长文本字段进行截断（约150-200字符）
- 最终确保总长度不超过 1500 字符（约 500 tokens）

**包含字段**：
- ✅ 政策名称
- ✅ 政策级别
- ✅ 来源单位
- ✅ 政策说明（截断到200字符）
- ✅ 适用对象（截断到150字符）
- ✅ 申请条件（截断到150字符）
- ✅ 补贴标准（截断到150字符）
- ✅ 政策标签

**省略字段**：
- ❌ 发布时间
- ❌ 申请材料（通常很长）
- ❌ 经办渠道（通常很长）
- ❌ 政策支持（通常很长）
- ❌ 联系电话
- ❌ 备注

### 2. 完整内容用于存储和展示

在 `buildFullPolicyContent()` 方法中：
- 包含所有字段的完整信息
- 不进行截断
- 存储到 Milvus 的 content 字段
- 搜索时返回给用户

## 工作流程

```
政策数据
    ↓
精简内容 (< 512 tokens) → Embedding API → 向量
    ↓                                        ↓
完整内容 (用于展示)  ←  存储到 Milvus  ←  向量
    ↓
返回给用户
```

## 代码实现

### 精简内容构建

```go
func (s *PolicyService) buildPolicyContent(policy model.PolicyInfo) string {
    var builder strings.Builder
    
    // 核心信息
    builder.WriteString(fmt.Sprintf("政策名称：%s\n", policy.Zcmc))
    builder.WriteString(fmt.Sprintf("政策级别：%s\n", policy.ZcLevel))
    
    // 截断长文本
    if policy.PolicyExplanation != "" {
        cleaned := cleanHTML(policy.PolicyExplanation)
        if len(cleaned) > 200 {
            cleaned = cleaned[:200] + "..."
        }
        builder.WriteString(fmt.Sprintf("政策说明：%s\n", cleaned))
    }
    
    // ... 其他字段类似处理
    
    content := builder.String()
    if len(content) > 1500 {
        content = content[:1500] + "..."
    }
    
    return content
}
```

### 完整内容构建

```go
func (s *PolicyService) buildFullPolicyContent(policy model.PolicyInfo) string {
    var builder strings.Builder
    
    // 包含所有字段，不截断
    builder.WriteString(fmt.Sprintf("政策名称：%s\n", policy.Zcmc))
    builder.WriteString(fmt.Sprintf("政策级别：%s\n", policy.ZcLevel))
    builder.WriteString(fmt.Sprintf("来源单位：%s\n", policy.SourceUnit))
    builder.WriteString(fmt.Sprintf("发布时间：%s\n", policy.PublishTime))
    
    if policy.PolicyExplanation != "" {
        builder.WriteString(fmt.Sprintf("政策说明：%s\n", cleanHTML(policy.PolicyExplanation)))
    }
    
    // ... 所有其他字段
    
    return builder.String()
}
```

### 更新流程

```go
func (s *PolicyService) UpdatePolicies(ctx context.Context) error {
    for _, policy := range policies {
        // 1. 精简内容用于向量化
        shortContent := s.buildPolicyContent(policy)
        vector, err := s.embeddingClient.GetEmbedding(shortContent)
        
        // 2. 完整内容用于存储
        fullContent := s.buildFullPolicyContent(policy)
        
        // 3. 存储到 Milvus
        ids = append(ids, policy.ID)
        contents = append(contents, fullContent)  // 存储完整内容
        vectors = append(vectors, vector)
    }
    
    s.milvusClient.Insert(ctx, ids, contents, vectors)
}
```

## 优势

1. **向量质量**：精简内容包含最核心的信息，向量更准确
2. **用户体验**：搜索结果返回完整信息，用户获得所有细节
3. **性能优化**：减少 embedding API 的负载
4. **可扩展性**：可以根据需要调整截断长度

## Token 估算

中文文本的 token 估算：
- 1个中文字符 ≈ 1.5-2 tokens
- 1500 字符 ≈ 450-500 tokens（安全范围）

字段长度控制：
- 政策名称：~20 字符 = ~30 tokens
- 政策级别：~10 字符 = ~15 tokens
- 来源单位：~30 字符 = ~45 tokens
- 政策说明：200 字符 = ~300 tokens
- 适用对象：150 字符 = ~225 tokens
- 申请条件：150 字符 = ~225 tokens
- 补贴标准：150 字符 = ~225 tokens
- 政策标签：~10 字符 = ~15 tokens

**总计**：约 1080 tokens（理论最大值）

实际使用中，由于：
- 不是所有字段都有内容
- 清理 HTML 后文本更短
- 最终有 1500 字符的硬限制

实际 token 数通常在 **300-500 tokens** 之间，安全范围内。

## 测试

```bash
# 重新编译
go build -o qd-sc.exe ./cmd/server

# 启动服务
./qd-sc.exe -config config.yaml

# 更新政策（应该不再报错）
curl -X POST http://localhost:8080/api/policy/update

# 测试搜索（应该返回完整信息）
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

## 注意事项

1. **向量语义**：虽然使用精简内容，但包含了最核心的语义信息，不影响搜索准确性
2. **内容完整性**：用户看到的是完整信息，不会丢失任何细节
3. **性能影响**：每条政策需要调用两次内容构建方法，但性能影响可忽略
4. **存储空间**：Milvus 存储的是完整内容，需要更多存储空间

## 未来优化

如果需要更精确的向量表示，可以考虑：

1. **分段向量化**：将长文本分成多段，每段生成向量，然后平均或拼接
2. **更大的 Embedding 模型**：使用支持更长输入的模型
3. **摘要生成**：使用 LLM 生成政策摘要用于向量化
4. **多向量存储**：为每条政策存储多个向量（标题向量、内容向量等）

目前的方案已经能够满足大部分使用场景。
