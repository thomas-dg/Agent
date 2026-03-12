package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"super-agent/internal/ai/models"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// 确保 planexecute 包被引用
var _ = planexecute.Config{}

// Runner 封装 AIOps 所需的所有重量级资源，进程启动时初始化一次，后续所有请求复用。
// planExecuteAgent 和 intentLLM 均为无状态对象，并发安全。
type Runner struct {
	planExecuteAgent adk.ResumableAgent
	intentLLM        model.ToolCallingChatModel
}

// NewRunner 初始化 AIOpsRunner，构建所有重量级资源。
// 应在进程启动时调用一次，返回的 Runner 供所有请求复用。
func NewRunner(ctx context.Context) (*Runner, error) {
	// 初始化意图解析 LLM（models 包内已保证单例）
	intentLLM, err := models.GetDeepSeekV3QuickWithOpenAI(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiops.NewRunner: 初始化意图解析模型失败: %w", err)
	}

	// 构建 Planner：sysPrompt 通过 context 在每次 Run 时动态注入，完全无状态
	planAgent, err := NewPlanner(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiops.NewRunner: 构建 Planner 失败: %w", err)
	}

	executeAgent, err := NewExecutor(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiops.NewRunner: 构建 Executor 失败: %w", err)
	}

	replanAgent, err := NewRePlanAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiops.NewRunner: 构建 Replanner 失败: %w", err)
	}

	planExecuteAgent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("aiops.NewRunner: 编译 PlanExecuteAgent 失败: %w", err)
	}

	return &Runner{
		planExecuteAgent: planExecuteAgent,
		intentLLM:        intentLLM,
	}, nil
}

// Run 使用预构建的 planExecuteAgent 执行 AIOps 分析。
// 通过 context 注入 per-request sysPrompt，每次调用都是独立的执行上下文，并发安全。
func (r *Runner) Run(ctx context.Context, opts Options) (string, error) {
	query := BuildQuery(opts)

	// 将当次请求的 sysPrompt 注入 context，供 Planner 的 GenInputFn 读取
	ctx = withSysPrompt(ctx, BuildSystemPrompt(opts))

	adkRunner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: r.planExecuteAgent,
	})
	iter := adkRunner.Query(ctx, query)

	var lastMessage adk.Message
	var iterErr error

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Output != nil {
			msg, _, err := adk.GetMessage(event)
			if err != nil {
				iterErr = fmt.Errorf("aiops.Run: 解析事件消息失败: %w", err)
				continue
			}
			lastMessage = msg
		}
	}

	if iterErr != nil && lastMessage == nil {
		return "", iterErr
	}
	if lastMessage == nil {
		return "", fmt.Errorf("aiops.Run: Agent 未返回任何消息")
	}
	return lastMessage.Content, nil
}

// RunStream 流式执行 AIOps 分析，每产生一个 Agent 步骤事件时调用 onStep 回调。
// onStep 接收当前步骤的消息内容，调用方可通过 SSE 等机制实时推送给客户端。
// 返回最终汇总结果和错误信息；各步骤详情已通过 onStep 实时推送，无需再次返回。
func (r *Runner) RunStream(ctx context.Context, opts Options, onStep func(msg string)) (string, error) {
	query := BuildQuery(opts)

	// 将当次请求的 sysPrompt 注入 context，供 Planner 的 GenInputFn 读取
	ctx = withSysPrompt(ctx, BuildSystemPrompt(opts))

	adkRunner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: r.planExecuteAgent,
	})
	iter := adkRunner.Query(ctx, query)

	var lastMessage adk.Message
	var iterErr error

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Output != nil {
			msg, _, err := adk.GetMessage(event)
			if err != nil {
				iterErr = fmt.Errorf("aiops.RunStream: 解析事件消息失败: %w", err)
				continue
			}
			lastMessage = msg
			// 每步结果实时回调，供调用方流式推送
			if onStep != nil {
				onStep(msg.String())
			}
		}
	}

	if iterErr != nil && lastMessage == nil {
		return "", iterErr
	}
	if lastMessage == nil {
		return "", fmt.Errorf("aiops.RunStream: Agent 未返回任何消息")
	}
	return lastMessage.Content, nil
}

// FormatIntentDescription 将解析后的 Options 转为自然语言描述，用于向用户确认系统的理解。
func FormatIntentDescription(opts Options) string {
	opts.normalize()

	sceneDesc := map[Scene]string{
		SceneAlertAnalysis: "告警分析",
		SceneLogAnalysis:   "日志分析",
		ScenePerfAnalysis:  "性能分析",
	}
	scene := sceneDesc[opts.Scene]
	if scene == "" {
		scene = string(opts.Scene)
	}

	desc := fmt.Sprintf("已理解您的需求：执行【%s】", scene)
	if opts.Target != "" {
		desc += fmt.Sprintf("，分析目标：%s", opts.Target)
	}
	desc += fmt.Sprintf("，时间范围：最近 %s", opts.TimeRange)
	if opts.Extra != "" {
		desc += fmt.Sprintf("，补充说明：%s", opts.Extra)
	}
	desc += "。正在启动分析，请稍候……"
	return desc
}

// ParseIntent 使用复用的 intentLLM 将自然语言解析为 Options。
// 解析失败时 fallback 为默认 Options，保证主流程不中断。
func (r *Runner) ParseIntent(ctx context.Context, query string) (Options, error) {
	msgs := []*schema.Message{
		schema.SystemMessage(intentParseSystemPrompt),
		schema.UserMessage(query),
	}

	resp, err := r.intentLLM.Generate(ctx, msgs)
	if err != nil {
		return defaultOptions(), fmt.Errorf("aiops.ParseIntent: LLM 调用失败: %w", err)
	}

	raw := strings.TrimSpace(resp.Content)
	// 兼容 LLM 偶尔输出 markdown 代码块的情况
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed parsedIntentJSON
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		// JSON 解析失败，fallback 为默认值，将原始 query 作为 Extra 保留语义
		opts := defaultOptions()
		opts.Extra = query
		return opts, nil
	}

	opts := Options{
		Scene:     Scene(parsed.Scene),
		Target:    parsed.Target,
		TimeRange: parsed.TimeRange,
		Extra:     parsed.Extra,
	}
	opts.normalize()
	return opts, nil
}
