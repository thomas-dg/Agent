package embedder

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/gogf/gf/v2/frame/g"
)

var embeddingDim = 2048

func NewQianwenEmbedder(ctx context.Context) (embedding.Embedder, error) {
	model, err := g.Cfg().Get(ctx, "embedding_model.model")
	if err != nil {
		return nil, err
	}
	log.Printf("embedding model: %s\n", model.String())
	api_key, err := g.Cfg().Get(ctx, "embedding_model.api_key")
	if err != nil {
		return nil, err
	}
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		Model:      model.String(),
		APIKey:     api_key.String(),
		Dimensions: &embeddingDim,
	})
	if err != nil {
		log.Printf("new embedder error: %v\n", err)
		return nil, err
	}
	return embedder, nil
}
