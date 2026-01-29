# 政策向量化查询功能使用指南

## 功能概述

本系统集成了政策向量化存储和智能检索功能，可以将政策文本转换为向量并存储在 Milvus 向量数据库中，支持基于语义相似度的政策查询。

## 系统架构

```
政策API → Embedding模型 → Milvus向量数据库 → 对话系统
```

### 组件说明

1. **政策数据源**: `https://www.xjksly.cn/sdrc-api/portal/policyInfo/portalList`
2. **Embedding服务**: `http://39.98.44.136:6017/emb/embed`
3. **Milvus数据库**: `39.98.44.136:6012`
4. **向量维度**: 768（根据embedding模型调整）

## 配置说明

在 `config.yaml` 中配置以下参数：

```yaml
# 政策API配置
policy:
  base_url: "https://www.xjksly.cn/sdrc-api/portal/policyInfo/portalList"
  timeout: 60s

# Embedding配置
embedding:
  base_url: "http://39.98.44.136:6017/emb/embed"
  timeout: 30s

# Milvus向量数据库配置
milvus:
  host: "39.98.44.136"
  port: 6012
  collection_name: "policy_vectors"
  dimension: 768
  timeout: 30s
```

## API接口

### 1. 更新政策到向量数据库

**接口**: `POST /api/policy/update`

**功能**: 从政策API获取最新政策，转换为向量并存储到Milvus

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/policy/update
```

**响应示例**:
```json
{
  "message": "政策更新成功"
}
```

**注意事项**:
- 首次使用前必须调用此接口初始化政策数据
- 建议定期调用以更新最新政策
- 更新过程可能需要几分钟，取决于政策数量

### 2. 搜索政策

**接口**: `GET /api/policy/search`

**参数**:
- `query` (必填): 查询关键词
- `topK` (可选): 返回结果数量，默认5

**请求示例**:
```bash
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

**响应示例**:
```json
{
  "query": "就业补贴",
  "results": [
    {
      "ID": "1988473569041494018",
      "Content": "政策名称：就业见习补贴\n政策级别：自治区级\n...",
      "Distance": 0.15
    }
  ]
}
```

### 3. 对话中查询政策

在对话接口中，AI助手会自动调用政策查询工具。

**请求示例**:
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

**AI助手会自动**:
1. 识别用户的政策查询意图
2. 调用 `queryPolicy` 工具搜索相关政策
3. 将搜索结果整理后返回给用户

## 使用流程

### 初始化流程

1. **启动服务**
```bash
./qd-sc.exe -config config.yaml
```

2. **初始化政策数据**
```bash
curl -X POST http://localhost:8080/api/policy/update
```

等待更新完成（可能需要几分钟）

3. **测试查询**
```bash
curl "http://localhost:8080/api/policy/search?query=创业扶持"
```

### 日常使用

用户可以通过对话接口直接询问政策相关问题：

- "有哪些就业补贴政策？"
- "创业扶持政策是什么？"
- "大学生就业见习补贴怎么申请？"

AI助手会自动搜索相关政策并提供详细信息。

## 政策数据结构

每条政策包含以下信息：

- **政策名称** (zcmc)
- **政策级别** (zcLevel): 如"自治区级"
- **来源单位** (sourceUnit)
- **发布时间** (publishTime)
- **政策说明** (policyExplanation)
- **适用对象** (applicableObjects)
- **申请条件** (applyCondition)
- **补贴标准** (btbz)
- **申请材料** (sqcl)
- **经办渠道** (jbqd)
- **政策支持** (zczc)
- **政策标签** (jyzcbq)

## 向量化处理

系统会将政策的关键信息组合成文本，然后通过Embedding模型转换为768维向量：

```
政策名称：就业见习补贴
政策级别：自治区级
来源单位：新疆维吾尔自治区人力资源和社会保障厅
发布时间：2021-09-24
政策说明：...
适用对象：...
申请条件：...
补贴标准：...
申请材料：...
经办渠道：...
```

## 相似度计算

- 使用 **L2距离** 计算向量相似度
- Distance 越小表示越相似
- 相似度评分 = 1.0 - Distance

## 维护建议

1. **定期更新**: 建议每周调用一次 `/api/policy/update` 接口更新政策数据
2. **监控日志**: 关注政策更新和查询的日志，及时发现问题
3. **性能优化**: 如果政策数量很大，可以考虑：
   - 增加 Milvus 索引参数
   - 调整批量插入大小
   - 使用异步更新

## 故障排查

### 问题1: 更新政策失败

**可能原因**:
- 政策API不可访问
- Embedding服务不可用
- Milvus连接失败

**解决方法**:
```bash
# 测试政策API
curl https://www.xjksly.cn/sdrc-api/portal/policyInfo/portalList

# 测试Embedding服务
curl -X POST http://39.98.44.136:6017/emb/embed \
  -H "Content-Type: application/json" \
  -d '{"input": "测试文本"}'

# 检查Milvus连接
# 查看服务日志
```

### 问题2: 搜索结果不准确

**可能原因**:
- 查询关键词不够精确
- 向量维度配置错误
- 政策数据未更新

**解决方法**:
- 使用更具体的关键词
- 检查 `dimension` 配置是否与embedding模型匹配
- 重新调用更新接口

### 问题3: 查询速度慢

**可能原因**:
- Milvus索引未优化
- 网络延迟
- 数据量过大

**解决方法**:
- 调整Milvus索引参数（HNSW的M和efConstruction）
- 减少topK数量
- 考虑使用缓存

## 扩展功能

### 自定义向量维度

如果使用不同的embedding模型，需要修改配置：

```yaml
milvus:
  dimension: 1024  # 根据实际模型调整
```

### 批量更新优化

对于大量政策，可以在 `policy_service.go` 中调整批量大小：

```go
// 分批处理，每批100条
batchSize := 100
for i := 0; i < len(policies); i += batchSize {
    end := i + batchSize
    if end > len(policies) {
        end = len(policies)
    }
    batch := policies[i:end]
    // 处理批次
}
```

## 技术栈

- **Go**: 后端服务
- **Milvus**: 向量数据库
- **Embedding模型**: 文本向量化（768维）
- **Gin**: Web框架
- **gRPC**: Milvus客户端通信

### Embedding API 格式

**请求**:
```json
{
  "inputs": "文本内容"
}
```

**响应**:
```json
[[0.010686361, -0.011039364, -0.015814532, ...]]
```
返回一个嵌套数组，外层数组包含一个内层数组，内层数组是 768 维的浮点数向量。

## 相关文件

- `internal/service/policy_service.go`: 政策服务实现
- `internal/client/embedding_client.go`: Embedding客户端
- `internal/client/milvus_client.go`: Milvus客户端
- `internal/api/handler/policy.go`: API处理器
- `internal/model/policy_vector.go`: 数据模型
- `config.yaml`: 配置文件
