package main

import (
	"context"
	"super-agent/internal/ai/agent/aiops"
	"super-agent/internal/ai/agent/chat"
	chatcontroller "super-agent/internal/controller/chat"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()

	// 进程启动时构建一次 Chat runner，后续所有请求复用，避免每次请求重建连接
	runner, err := chat.BuildChatAgent(context.Background())
	if err != nil {
		g.Log().Fatalf(ctx, "初始化 Chat Agent 失败: %v", err)
	}

	// 进程启动时构建一次 AIOps runner，后续所有请求复用，避免每次请求重建 Agent 和 LLM 客户端
	aiopsRunner, err := aiops.NewRunner(context.Background())
	if err != nil {
		g.Log().Fatalf(ctx, "初始化 AIOps Runner 失败: %v", err)
	}

	s := g.Server()
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareCORS)
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(chatcontroller.NewV1(runner, aiopsRunner))
	})
	s.SetPort(6872)
	s.Run()
}
