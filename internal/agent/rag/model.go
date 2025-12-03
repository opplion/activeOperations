package rag

import (
	"activeOperations/config"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var (
	Schema       *entity.Schema
	IndexOptions []milvusclient.CreateIndexOption
)

// 生成 Schema
func NewSchema() {
	cfg := config.GetConfig()
	Schema = entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("id").WithIsAutoID(false).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(int64(*cfg.Embedding.Dimensions))).
		WithField(entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(10000)).
		WithField(entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON).WithMaxLength(2048))
}

// 生成 IndexOptions
func NewIndexOptions() {
	cfg := config.GetConfig()
	IndexOptions = []milvusclient.CreateIndexOption{
		milvusclient.NewCreateIndexOption(cfg.Milvus.CollectionName, "vector", index.NewAutoIndex(entity.COSINE)),
		milvusclient.NewCreateIndexOption(cfg.Milvus.CollectionName, "id", index.NewAutoIndex(entity.COSINE)),
	}
}
