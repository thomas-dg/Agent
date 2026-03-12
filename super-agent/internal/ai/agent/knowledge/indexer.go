package knowledge

import (
	"context"
	indexer2 "super-agent/internal/ai/indexer"

	"github.com/cloudwego/eino/components/indexer"
)

// newIndexer component initialization function of node 'MilvusIndexer' in graph 'knowledge'
func newIndexer(ctx context.Context) (idr indexer.Indexer, err error) {
	// TODO Modify component configuration here.
	//config := &redis.IndexerConfig{}
	//embeddingIns11, err := newEmbedding(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//config.Embedding = embeddingIns11
	//idr, err = redis.NewIndexer(ctx, config)
	//if err != nil {
	//	return nil, err
	//}
	//return idr, nil

	return indexer2.NewMilvusIndexer(ctx)
}
