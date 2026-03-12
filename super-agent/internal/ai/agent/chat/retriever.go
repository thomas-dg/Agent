package chat

import (
	"context"
	myretriever "super-agent/internal/ai/retriever"

	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever component initialization function of node 'MilvusRetriever' in graph 'chat'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	// TODO Modify component configuration here.
	return myretriever.NewMilvusRetriever(ctx, 3)
}
