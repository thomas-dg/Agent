package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

//type ChatReq struct {
//	g.Meta   `path:"/chat" method:"post" summary:"对话"`
//	ID       string `v:"required#会话ID不能为空"`
//	Question string `v:"required#问题内容不能为空"`
//}
//
//type ChatRes struct {
//	Answer string `json:"answer"`
//}

type ChatStreamReq struct {
	g.Meta    `path:"/chat_stream" method:"post" summary:"流式对话"`
	UserID    string `json:"userId"    dc:"用户ID，由前端生成并持久化到本地存储"`
	SessionID string `json:"sessionId" dc:"会话ID，由前端生成" v:"required#会话ID不能为空"`
	Question  string `json:"question"  dc:"问题内容" v:"required#问题内容不能为空"`
}

type ChatStreamRes struct {
}

type FileUploadReq struct {
	g.Meta `path:"/upload" method:"post" mime:"multipart/form-data" summary:"文件上传"`
}

type FileUploadRes struct {
	FileName string `json:"fileName" dc:"保存的文件名"`
	FilePath string `json:"filePath" dc:"文件保存路径"`
	FileSize int64  `json:"fileSize" dc:"文件大小(字节)"`
	Indexing bool   `json:"indexing" dc:"知识库构建是否已在后台异步进行"`
}

type AIOpsReq struct {
	g.Meta    `path:"/ai_ops" method:"post" summary:"AI运维"`
	UserID    string `json:"userId" dc:"用户ID，由前端生成并持久化到本地存储"`
	SessionID string `json:"sessionId" dc:"会话ID，由前端生成"`
	Query     string `v:"required#运维分析需求不能为空" json:"query" dc:"自然语言描述，如：分析 order-service 最近 2 小时的告警"`
}

type AIOpsRes struct {
}

//type PingReq struct {
//	g.Meta `path:"/ping" method:"get" summary:"健康检查"`
//}
//
//type PingRes struct {
//	Message string `json:"message" dc:"服务状态信息"`
//}
