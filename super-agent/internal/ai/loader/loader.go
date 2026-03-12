package loader

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

func NewFileLoader(ctx context.Context) (document.Loader, error) {
	config := &file.FileLoaderConfig{}
	loader, err := file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}
	return loader, nil
}
