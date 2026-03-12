package aiops

import (
	"context"
	"super-agent/internal/ai/models"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/schema"
)

// sysPromptKey 是存储在 context 中的 system prompt 的 key 类型，避免 key 冲突
type sysPromptKey struct{}

// withSysPrompt 将 sysPrompt 注入 context，供 GenInputFn 读取
func withSysPrompt(ctx context.Context, sysPrompt string) context.Context {
	return context.WithValue(ctx, sysPromptKey{}, sysPrompt)
}

// getSysPromptFromCtx 从 context 中读取 sysPrompt，不存在时返回默认值
func getSysPromptFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(sysPromptKey{}).(string); ok && v != "" {
		return v
	}
	return BuildSystemPrompt(defaultOptions())
}

// NewPlanner 创建 Planner Agent。
// sysPrompt 通过 context 传递（使用 withSysPrompt 注入），使同一个 Planner 实例
// 可以服务不同 opts 的并发请求，完全无状态。
func NewPlanner(ctx context.Context) (adk.Agent, error) {
	planModel, err := models.GetDeepSeekV3ThinkWithOpenAI(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: planModel,
		GenInputFn: func(ctx context.Context, userInput []adk.Message) ([]adk.Message, error) {
			msgs := []adk.Message{
				schema.SystemMessage(getSysPromptFromCtx(ctx)),
			}
			msgs = append(msgs, userInput...)
			return msgs, nil
		},
	})
}
