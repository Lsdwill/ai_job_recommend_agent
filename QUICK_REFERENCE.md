# 政策向量化功能快速参考

## 🚀 快速开始

```bash
# 1. 编译
go build -o qd-sc.exe ./cmd/server

# 2. 启动
./qd-sc.exe -config config.yaml

# 3. 初始化政策数据
curl -X POST http://localhost:8080/api/policy/update

# 4. 测试搜索
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

## 📋 API 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/policy/update` | POST | 更新政策到向量数据库 |
| `/api/policy/search` | GET | 搜索政策 |
| `/v1/chat/completions` | POST | 对话接口（自动调用政策查询） |

## 🔧 配置要点

```yaml
# Embedding API
embedding:
  base_url: "http://39.98.44.136:6017/emb/embed"
  timeout: 30s

# Milvus 向量数据库
milvus:
  host: "39.98.44.136"
  port: 6012
  collection_name: "policy_vectors"
  dimension: 768
  timeout: 30s

# 政策 API
policy:
  base_url: "https://www.xjksly.cn/sdrc-api/portal/policyInfo/portalList"
  timeout: 60s
```

## 🐛 常见错误及解决

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `missing field 'inputs'` | API 字段错误 | 已修复：使用 `inputs` 字段 |
| `must have less than 512 tokens` | 文本太长 | 已修复：自动截断到 500 tokens |
| `cannot unmarshal array` | 响应格式错误 | 已修复：使用嵌套数组 `[][]float32` |
| `连接 Milvus 失败` | 网络或配置问题 | 检查 host/port 配置 |

## 📊 Embedding API 格式

### 请求
```json
{
  "inputs": "文本内容（< 512 tokens）"
}
```

### 响应
```json
[[0.01, -0.02, 0.03, ..., 0.04]]
```
- 嵌套数组格式
- 外层数组支持批量
- 内层数组是 768 维向量

## 💡 核心策略

### 双内容策略

**向量化内容**（精简，< 512 tokens）：
- 政策名称、级别、来源单位
- 政策说明（截断到 200 字符）
- 适用对象（截断到 150 字符）
- 申请条件（截断到 150 字符）
- 补贴标准（截断到 150 字符）
- 政策标签

**展示内容**（完整）：
- 包含所有字段
- 不截断
- 存储在 Milvus
- 搜索时返回

## 📈 性能指标

| 指标 | 预期值 |
|------|--------|
| 向量化成功率 | > 80% |
| 单条政策处理时间 | ~200ms |
| 50 条政策总时间 | ~5-10 分钟 |
| 搜索响应时间 | < 1 秒 |

## 🔍 测试命令

```bash
# 测试 Embedding API
curl -X POST http://39.98.44.136:6017/emb/embed \
  -H "Content-Type: application/json" \
  -d '{"inputs":"测试文本"}'

# 测试健康检查
curl http://localhost:8080/health

# 更新政策
curl -X POST http://localhost:8080/api/policy/update

# 搜索政策
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"

# 对话测试
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [{"role": "user", "content": "我想了解就业补贴政策"}]
  }'
```

## 📚 文档索引

| 文档 | 说明 |
|------|------|
| `FINAL_TEST_GUIDE.md` | 完整测试指南 |
| `EMBEDDING_FIX.md` | API 格式修复详情 |
| `EMBEDDING_TOKEN_LIMIT.md` | Token 限制解决方案 |
| `POLICY_VECTOR_GUIDE.md` | 详细使用指南 |
| `POLICY_FEATURE.md` | 功能说明 |

## ⚙️ 维护任务

### 每周
- [ ] 更新政策数据：`curl -X POST http://localhost:8080/api/policy/update`
- [ ] 检查成功率日志
- [ ] 验证搜索功能

### 每月
- [ ] 检查 Milvus 存储空间
- [ ] 优化向量化策略
- [ ] 收集用户反馈

### 按需
- [ ] 调整截断长度
- [ ] 优化搜索参数
- [ ] 更新配置

## 🎯 关键代码位置

| 功能 | 文件 |
|------|------|
| Embedding 客户端 | `internal/client/embedding_client.go` |
| Milvus 客户端 | `internal/client/milvus_client.go` |
| 政策服务 | `internal/service/policy_service.go` |
| API 处理器 | `internal/api/handler/policy.go` |
| 数据模型 | `internal/model/policy_vector.go` |
| 配置 | `internal/config/config.go` |

## 🔐 安全注意事项

1. **API 密钥**: 不要在代码中硬编码，使用环境变量
2. **网络访问**: 确保 Milvus 和 Embedding API 的网络安全
3. **数据隐私**: 政策数据可能包含敏感信息，注意访问控制
4. **速率限制**: 避免过快请求 Embedding API

## 📞 支持

遇到问题？检查：
1. 服务日志
2. 相关文档
3. 配置文件
4. 网络连接

---

**版本**: 1.0.0  
**最后更新**: 2026-01-28
