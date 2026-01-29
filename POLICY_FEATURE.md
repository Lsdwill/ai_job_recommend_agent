# 政策向量化查询功能

## 新增功能

系统现已集成**政策智能查询**功能，支持基于语义相似度的政策检索。

### 核心特性

✅ **自动向量化**: 将政策文本转换为向量存储  
✅ **语义搜索**: 基于语义相似度而非关键词匹配  
✅ **智能对话**: AI助手自动识别政策查询意图  
✅ **实时更新**: 支持定期更新最新政策数据  

## 快速开始

### 1. 配置

在 `config.yaml` 中添加以下配置：

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

### 2. 初始化政策数据

首次使用前，需要初始化政策向量数据库：

```bash
# 启动服务
./qd-sc.exe -config config.yaml

# 初始化政策数据（需要几分钟）
curl -X POST http://localhost:8080/api/policy/update
```

### 3. 使用方式

#### 方式一：直接搜索API

```bash
curl "http://localhost:8080/api/policy/search?query=就业补贴&topK=3"
```

#### 方式二：对话接口（推荐）

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "我想了解就业见习补贴政策"}
    ]
  }'
```

AI助手会自动：
1. 识别政策查询意图
2. 搜索相关政策
3. 整理并返回结果

## 支持的政策类型

- 就业补贴政策
- 创业扶持政策
- 见习补贴政策
- 职业培训政策
- 社保医保政策
- 人才引进政策

## 示例对话

**用户**: "有哪些就业补贴政策？"

**AI助手**: 
```
为您找到 3 条相关政策：

【政策 1】
政策名称：就业见习补贴
政策级别：自治区级
来源单位：新疆维吾尔自治区人力资源和社会保障厅
发布时间：2021-09-24
适用对象：离校2年内未就业高校毕业生和16-24岁新疆籍失业青年
补贴标准：按当地最低工资标准给予见习补贴...
申请材料：《就业见习补贴申请表》、参加就业见习人员名单...
经办渠道：新疆人社公共服务平台 https://rsggfw.rst.xinjiang.gov.cn

【政策 2】
...
```

## API接口

### 更新政策

- **接口**: `POST /api/policy/update`
- **功能**: 从政策API获取最新政策并更新向量数据库
- **建议**: 每周执行一次

### 搜索政策

- **接口**: `GET /api/policy/search`
- **参数**: 
  - `query`: 查询关键词（必填）
  - `topK`: 返回结果数量（可选，默认5）

## 技术实现

```
政策API → Embedding模型 → Milvus向量数据库 → 语义搜索
```

1. **数据获取**: 从政策API获取结构化政策数据
2. **文本处理**: 清理HTML标签，组合关键字段
3. **向量化**: 通过Embedding模型转换为768维向量
4. **存储**: 存储到Milvus向量数据库
5. **检索**: 基于L2距离计算相似度

## 维护

### 定期更新

建议设置定时任务每周更新一次：

```bash
# Linux/Mac
0 2 * * 0 curl -X POST http://localhost:8080/api/policy/update

# Windows任务计划程序
schtasks /create /tn "UpdatePolicy" /tr "curl -X POST http://localhost:8080/api/policy/update" /sc weekly /d SUN /st 02:00
```

### 监控

查看日志确认更新状态：

```bash
# 查看最近的日志
tail -f logs/app.log | grep policy
```

## 故障排查

### 更新失败

1. 检查政策API是否可访问
2. 检查Embedding服务是否正常
3. 检查Milvus连接状态

### 搜索结果不准确

1. 使用更具体的关键词
2. 增加topK数量
3. 重新更新政策数据

## 更多信息

详细文档请参考：[POLICY_VECTOR_GUIDE.md](./POLICY_VECTOR_GUIDE.md)
