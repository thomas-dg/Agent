package sse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// Connection 表示SSE客户端连接
type Connection struct {
	ID      string
	Request *ghttp.Request
	ctx     context.Context
	cancel  context.CancelFunc
}

type Service struct {
}

func NewSSE() *Service {
	return &Service{}
}

func (s *Service) CreateConnection(ctx context.Context, r *ghttp.Request) (*Connection, error) {
	// 设置SSE头
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建新的连接
	connectionID := r.Get("client_id", guid.S()).String()
	connCtx, cancel := context.WithCancel(ctx)
	conn := &Connection{
		ID:      connectionID,
		Request: r,
		ctx:     connCtx,
		cancel:  cancel,
	}

	// 发送连接成功消息
	r.Response.Writefln("id: %s", connectionID)
	r.Response.Writefln("event: connected")
	r.Response.Writefln("data: {\"status\": \"connected\", \"client_id\": \"%s\"}\n", connectionID)
	r.Response.Flush()
	return conn, nil
}

// SendMessage 向客户端发送SSE消息
func (c *Connection) SendMessage(eventType, data string) bool {
	select {
	case <-c.ctx.Done():
		return false // 连接已关闭
	default:
		// 将 data 中的换行符转义，避免破坏 SSE 协议格式（每行 data: 只能有一行）
		escapedData := strings.ReplaceAll(data, "\n", "\\n")
		msg := fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", time.Now().UnixNano(), eventType, escapedData)
		c.Request.Response.Writefln(msg)
		c.Request.Response.Flush()
		return true
	}
}

// Close 关闭连接
func (c *Connection) Close() {
	c.cancel()
}
