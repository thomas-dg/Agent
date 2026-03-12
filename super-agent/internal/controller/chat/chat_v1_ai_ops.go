package chat

import (
	"context"
	"encoding/json"
	v1 "super-agent/api/chat/v1"
	"super-agent/internal/ai/agent/aiops"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) AIOps(ctx context.Context, req *v1.AIOpsReq) (res *v1.AIOpsRes, err error) {
	// 将自然语言解析为结构化 Options（复用预初始化的 intentLLM）
	opts, err := c.aiopsRunner.ParseIntent(ctx, req.Query)
	if err != nil {
		// ParseIntent 内部已 fallback 为默认 Options，此处记录日志后继续
		g.Log().Warningf(ctx, "AIOps: 意图解析失败，使用默认 Options: %v", err)
	}

	// 建立 SSE 连接，后续通过事件流实时推送每个执行步骤
	client, err := c.service.CreateConnection(ctx, g.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	// 将解析结果以自然语言推送给客户端，让用户确认系统的理解
	client.SendMessage("intent", aiops.FormatIntentDescription(opts))

	// 流式执行：每产生一个 Agent 步骤事件，立即通过 SSE 推送给客户端
	finalResult, err := c.aiopsRunner.RunStream(ctx, opts, func(msg string) {
		client.SendMessage("step", msg)
	})
	if err != nil {
		client.SendMessage("error", err.Error())
		return &v1.AIOpsRes{}, nil
	}
	if finalResult == "" {
		client.SendMessage("error", "内部错误")
		return &v1.AIOpsRes{}, nil
	}

	// 推送最终汇总报告（区别于 step 消息，前端以报告样式展示）
	donePayload, _ := json.Marshal(map[string]any{
		"scene":  string(opts.Scene),
		"result": finalResult,
	})
	client.SendMessage("done", string(donePayload))

	return &v1.AIOpsRes{}, nil
}
