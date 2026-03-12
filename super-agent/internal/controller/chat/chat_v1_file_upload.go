package chat

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	v1 "super-agent/api/chat/v1"
	"super-agent/internal/ai/agent/knowledge"
	myloader "super-agent/internal/ai/loader"
	"super-agent/repo"
	"super-agent/utils/callback"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

var fileDir = "./docs/"

const (
	// maxUploadSize 单文件最大上传大小：1MB
	maxUploadSize = 1 * 1024 * 1024
)

// allowedExtensions 允许上传的文件扩展名白名单
var allowedExtensions = map[string]bool{
	//".txt":  true,
	".md": true,
	//".pdf":  true,
	//".docx": true,
	//".doc":  true,
	//".csv":  true,
	//".json": true,
}

func (c *ControllerV1) FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error) {
	// 从请求中获取上传的文件
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	if uploadFile == nil {
		return nil, gerror.New("请上传文件")
	}

	// 校验文件大小
	if uploadFile.Size > maxUploadSize {
		return nil, gerror.Newf("文件大小超出限制，最大允许 %dMB，当前文件 %.2fMB",
			maxUploadSize/1024/1024, float64(uploadFile.Size)/1024/1024)
	}

	// 校验文件类型（扩展名白名单）
	ext := strings.ToLower(filepath.Ext(uploadFile.Filename))
	if !allowedExtensions[ext] {
		allowed := make([]string, 0, len(allowedExtensions))
		for k := range allowedExtensions {
			allowed = append(allowed, k)
		}
		return nil, gerror.Newf("不支持的文件类型 %q，允许的类型：%s", ext, strings.Join(allowed, ", "))
	}

	// 确保保存目录存在
	if !gfile.Exists(fileDir) {
		if err := gfile.Mkdir(fileDir); err != nil {
			return nil, gerror.Wrapf(err, "创建目录失败: %s", fileDir)
		}
	}

	// 获取原始文件名
	newFileName := uploadFile.Filename
	// 完整的保存路径
	savePath := filepath.Join(fileDir)

	// 保存文件
	_, err = uploadFile.Save(savePath, false)
	if err != nil {
		return nil, gerror.Wrapf(err, "保存文件失败")
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filepath.Join(savePath, newFileName))
	if err != nil {
		return nil, gerror.Wrapf(err, "获取文件信息失败")
	}

	res = &v1.FileUploadRes{
		FileName: newFileName,
		FilePath: savePath,
		FileSize: fileInfo.Size(),
		Indexing: true,
	}

	// 知识库构建异步执行，避免阻塞 HTTP 响应
	// 构建失败通过日志记录，不影响文件上传结果
	fullPath := fileDir + "/" + newFileName
	go func() {
		bgCtx := context.Background()
		if err = buildIntoIndex(bgCtx, fullPath); err != nil {
			log.Printf("[error] async buildIntoIndex failed, file=%s: %v\n", fullPath, err)
		}
	}()

	return res, nil
}

func buildIntoIndex(ctx context.Context, path string) error {
	r, err := knowledge.BuildKnowledge(ctx)
	if err != nil {
		return err
	}

	// 删除biz数据metadata中_source一样的数据
	loader, err := myloader.NewFileLoader(ctx)
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
	}
	if len(queryResult) > 0 {
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
				log.Printf("[warn] delete existing data failed: %v\n", err)
			} else {
				log.Printf("[info] deleted %d existing records with _source: %s\n", len(idsToDelete), docs[0].MetaData["_source"])
			}
		}
	}

	// 重新构建
	ids, err := r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(callback.LogCallback(nil)))
	if err != nil {
		return fmt.Errorf("invoke index graph failed: %w", err)
	}
	log.Printf("[done] indexing file: %s, len of parts: %d\n", path, len(ids))
	return nil
}
