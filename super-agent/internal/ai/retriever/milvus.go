package retriever

import (
	"context"
	"super-agent/internal/ai/embedder"
	"super-agent/repo"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
)

func NewMilvusRetriever(ctx context.Context, topK int) (rtr retriever.Retriever, err error) {
	cli, err := repo.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder.NewQianwenEmbedder(ctx)
	if err != nil {
		return nil, err
	}
	r, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      cli,
		Collection:  repo.MilvusCollectionName,
		VectorField: repo.MilvusFieldVector,
		OutputFields: []string{
			repo.MilvusFieldID,
			repo.MilvusFieldContent,
			repo.MilvusFieldMetadata,
		},
		TopK:      topK,
		Embedding: eb,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
