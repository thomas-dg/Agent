package indexer

import (
	"context"
	"super-agent/internal/ai/embedder"
	"super-agent/repo"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
)

func NewMilvusIndexer(ctx context.Context) (*milvus.Indexer, error) {
	cli, err := repo.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}

	eb, err := embedder.NewQianwenEmbedder(ctx)
	if err != nil {
		return nil, err
	}
	config := &milvus.IndexerConfig{
		Client:     cli,
		Collection: repo.MilvusCollectionName,
		Fields:     repo.MilvusFields,
		Embedding:  eb,
	}
	indexer, err := milvus.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}
	return indexer, nil
}
