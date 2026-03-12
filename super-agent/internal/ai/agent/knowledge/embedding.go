package knowledge

import (
	"context"
	"super-agent/internal/ai/embedder"

	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	// TODO Modify component configuration here.
	//config := &openai.EmbeddingConfig{}
	//eb, err = openai.NewEmbedder(ctx, config)
	//if err != nil {
	//	return nil, err
	//}
	//return eb, nil

	return embedder.NewQianwenEmbedder(ctx)
}
