package repo

import (
	"context"
	"fmt"

	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	milvusAddr           = "172.17.0.1:19530" // localhost:19530
	milvusDBName         = "agent"
	MilvusCollectionName = "biz"
	defaultDBName        = "default"
)

const (
	MilvusFieldID       = "id"
	MilvusFieldVector   = "vector"
	MilvusFieldContent  = "content"
	MilvusFieldMetadata = "metadata"
)

// MilvusFields milvus字段
var MilvusFields = []*entity.Field{
	{
		Name:     MilvusFieldID,
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "256",
		},
		PrimaryKey: true,
	},
	{
		Name:     MilvusFieldVector,
		DataType: entity.FieldTypeBinaryVector,
		TypeParams: map[string]string{
			"dim": "65536",
		},
	},
	{
		Name:     MilvusFieldContent,
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "18192",
		},
	},
	{
		Name:     MilvusFieldMetadata,
		DataType: entity.FieldTypeJSON,
	},
}

// indexConfig 定义索引配置
type indexConfig struct {
	fieldName   string
	metricType  entity.MetricType
	description string
}

var indexConfigs = []indexConfig{
	{MilvusFieldID, entity.L2, "ID field index"},
	{MilvusFieldContent, entity.L2, "Content field index"},
	{MilvusFieldVector, entity.HAMMING, "Vector field index"},
}

// NewMilvusClient 创建Milvus客户端
func NewMilvusClient(ctx context.Context) (cli.Client, error) {
	// 1. 确保agent数据库存在
	if err := ensureAgentDatabase(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure agent database: %w", err)
	}

	// 2. 连接到agent数据库
	agentClient, err := cli.NewClient(ctx, cli.Config{
		Address: milvusAddr,
		DBName:  milvusDBName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent database: %w", err)
	}

	// 3. 确保biz collection存在
	if err := ensureBizCollection(ctx, agentClient); err != nil {
		agentClient.Close()
		return nil, fmt.Errorf("failed to ensure biz collection: %w", err)
	}

	return agentClient, nil
}

// ensureAgentDatabase 确保agent数据库存在
func ensureAgentDatabase(ctx context.Context) error {
	defaultClient, err := cli.NewClient(ctx, cli.Config{
		Address: milvusAddr,
		DBName:  defaultDBName,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}
	defer defaultClient.Close()

	// 检查agent数据库是否存在
	if exists, err := databaseExists(ctx, defaultClient, milvusDBName); err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	} else if !exists {
		if err := defaultClient.CreateDatabase(ctx, milvusDBName); err != nil {
			return fmt.Errorf("failed to create agent database: %w", err)
		}
	}

	return nil
}

// databaseExists 检查数据库是否存在
func databaseExists(ctx context.Context, client cli.Client, dbName string) (bool, error) {
	databases, err := client.ListDatabases(ctx)
	if err != nil {
		return false, err
	}

	for _, db := range databases {
		if db.Name == dbName {
			return true, nil
		}
	}
	return false, nil
}

// ensureBizCollection 确保biz collection存在
func ensureBizCollection(ctx context.Context, client cli.Client) error {
	// 检查collection是否存在
	if exists, err := collectionExists(ctx, client, MilvusCollectionName); err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	} else if !exists {
		if err := createBizCollection(ctx, client); err != nil {
			return fmt.Errorf("failed to create biz collection: %w", err)
		}
	}

	return nil
}

// collectionExists 检查collection是否存在
func collectionExists(ctx context.Context, client cli.Client, collectionName string) (bool, error) {
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, collection := range collections {
		if collection.Name == collectionName {
			return true, nil
		}
	}
	return false, nil
}

// createBizCollection 创建biz collection及其索引
func createBizCollection(ctx context.Context, client cli.Client) error {
	// 创建collection schema
	schema := &entity.Schema{
		CollectionName: MilvusCollectionName,
		Description:    "Business knowledge collection",
		Fields:         MilvusFields,
	}

	if err := client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// 创建索引
	if err := createIndexes(ctx, client); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes 为collection创建所有必要的索引
func createIndexes(ctx context.Context, client cli.Client) error {
	for _, config := range indexConfigs {
		if err := createIndex(ctx, client, config); err != nil {
			return fmt.Errorf("failed to create %s: %w", config.description, err)
		}
	}
	return nil
}

// createIndex 创建单个索引
func createIndex(ctx context.Context, client cli.Client, config indexConfig) error {
	index, err := entity.NewIndexAUTOINDEX(config.metricType)
	if err != nil {
		return fmt.Errorf("failed to create index config: %w", err)
	}

	if err := client.CreateIndex(ctx, MilvusCollectionName, config.fieldName, index, false); err != nil {
		return fmt.Errorf("failed to create index on field %s: %w", config.fieldName, err)
	}

	return nil
}
