package main

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"super-agent/internal/ai/agent/knowledge"
	loader2 "super-agent/internal/ai/loader"
	"super-agent/repo"
	"super-agent/utils/callback"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func main() {
	ctx := context.Background()
	r, err := knowledge.BuildKnowledge(ctx)
	if err != nil {
		panic(err)
	}
	err = filepath.WalkDir("./internal/cmd/knowledge/docs", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// 跳过无法访问的文件，不中断整个遍历
			fmt.Printf("[warn] access error, skipping %s: %v\n", path, walkErr)
			return nil
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			fmt.Printf("[skip] not a markdown file: %s\n", path)
			return nil
		}

		fmt.Printf("[start] indexing file: %s\n", path)
		// 删除biz数据metadata中_source一样的数据
		loader, err := loader2.NewFileLoader(ctx)
		if err != nil {
			return err
		}
		docs, err := loader.Load(ctx, document.Source{URI: path})
		if err != nil {
			return err
		}
		cli, err := repo.NewMilvusClient(ctx)
		if err != nil {
			return err
		}
		// 查询所有metadata中_source一样的数据并删除
		expr := fmt.Sprintf(`metadata["_source"] == "%s"`, docs[0].MetaData["_source"])
		queryResult, err := cli.Query(ctx, repo.MilvusCollectionName, []string{}, expr, []string{"id"})
		if err != nil {
			return err
		} else if len(queryResult) > 0 {
			// 提取所有需要删除的id
			var idsToDelete []string
			for _, column := range queryResult {
				if column.Name() == "id" {
					for i := 0; i < column.Len(); i++ {
						id, err := column.GetAsString(i)
						if err == nil {
							idsToDelete = append(idsToDelete, id)
						}
					}
				}
			}
			// 删除这些数据
			if len(idsToDelete) > 0 {
				deleteExpr := fmt.Sprintf(`id in ["%s"]`, strings.Join(idsToDelete, `","`))
				err = cli.Delete(ctx, repo.MilvusCollectionName, "", deleteExpr)
				if err != nil {
					fmt.Printf("[warn] delete existing data failed: %v\n", err)
				} else {
					fmt.Printf("[info] deleted %d existing records with _source: %s\n", len(idsToDelete), docs[0].MetaData["_source"])
				}
			}
		}
		// 重新构建
		ids, err := r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(callback.LogCallback(nil)))
		if err != nil {
			return fmt.Errorf("[debug]invoke index graph failed: %w", err)
		}
		fmt.Printf("[done] indexing file: %s, len of parts: %d，%s\n", path, len(ids), ids)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("walk docs dir failed: %w", err))
	}
}
