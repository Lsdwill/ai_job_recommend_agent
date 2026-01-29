package client

import (
	"context"
	"fmt"
	"qd-sc/internal/config"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// MilvusClient Milvus客户端
type MilvusClient struct {
	client         client.Client
	collectionName string
	dimension      int
}

// NewMilvusClient 创建Milvus客户端
func NewMilvusClient(cfg *config.MilvusConfig) (*MilvusClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	c, err := client.NewGrpcClient(ctx, fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("连接Milvus失败: %w", err)
	}

	mc := &MilvusClient{
		client:         c,
		collectionName: cfg.CollectionName,
		dimension:      cfg.Dimension,
	}

	// 初始化集合
	if err := mc.initCollection(ctx); err != nil {
		return nil, err
	}

	return mc, nil
}

// initCollection 初始化集合
func (m *MilvusClient) initCollection(ctx context.Context) error {
	// 检查集合是否存在
	has, err := m.client.HasCollection(ctx, m.collectionName)
	if err != nil {
		return fmt.Errorf("检查集合失败: %w", err)
	}

	if !has {
		// 创建集合
		schema := &entity.Schema{
			CollectionName: m.collectionName,
			Description:    "政策向量存储",
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeVarChar,
					PrimaryKey: true,
					AutoID:     false,
					TypeParams: map[string]string{
						"max_length": "256",
					},
				},
				{
					Name:     "content",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "65535",
					},
				},
				{
					Name:     "vector",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": fmt.Sprintf("%d", m.dimension),
					},
				},
			},
		}

		if err := m.client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}

		// 创建索引
		idx, err := entity.NewIndexHNSW(entity.L2, 8, 200)
		if err != nil {
			return fmt.Errorf("创建索引配置失败: %w", err)
		}

		if err := m.client.CreateIndex(ctx, m.collectionName, "vector", idx, false); err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	// 加载集合
	if err := m.client.LoadCollection(ctx, m.collectionName, false); err != nil {
		return fmt.Errorf("加载集合失败: %w", err)
	}

	return nil
}

// Insert 插入向量
func (m *MilvusClient) Insert(ctx context.Context, ids []string, contents []string, vectors [][]float32) error {
	idColumn := entity.NewColumnVarChar("id", ids)
	contentColumn := entity.NewColumnVarChar("content", contents)
	vectorColumn := entity.NewColumnFloatVector("vector", m.dimension, vectors)

	if _, err := m.client.Insert(ctx, m.collectionName, "", idColumn, contentColumn, vectorColumn); err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	// 刷新以确保数据持久化
	if err := m.client.Flush(ctx, m.collectionName, false); err != nil {
		return fmt.Errorf("刷新数据失败: %w", err)
	}

	return nil
}

// Search 搜索相似向量
func (m *MilvusClient) Search(ctx context.Context, vector []float32, topK int) ([]SearchResult, error) {
	sp, _ := entity.NewIndexHNSWSearchParam(16)

	searchResult, err := m.client.Search(
		ctx,
		m.collectionName,
		[]string{},
		"",
		[]string{"id", "content"},
		[]entity.Vector{entity.FloatVector(vector)},
		"vector",
		entity.L2,
		topK,
		sp,
	)

	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	if len(searchResult) == 0 {
		return []SearchResult{}, nil
	}

	results := make([]SearchResult, 0)
	for i := 0; i < searchResult[0].ResultCount; i++ {
		id, _ := searchResult[0].IDs.GetAsString(i)
		content := ""
		if contentField := searchResult[0].Fields.GetColumn("content"); contentField != nil {
			if contentCol, ok := contentField.(*entity.ColumnVarChar); ok {
				content, _ = contentCol.ValueByIdx(i)
			}
		}

		results = append(results, SearchResult{
			ID:       id,
			Content:  content,
			Distance: searchResult[0].Scores[i],
		})
	}

	return results, nil
}

// Delete 删除向量
func (m *MilvusClient) Delete(ctx context.Context, ids []string) error {
	expr := fmt.Sprintf("id in %v", ids)
	if err := m.client.Delete(ctx, m.collectionName, "", expr); err != nil {
		return fmt.Errorf("删除数据失败: %w", err)
	}
	return nil
}

// Close 关闭客户端
func (m *MilvusClient) Close() error {
	return m.client.Close()
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       string
	Content  string
	Distance float32
}
