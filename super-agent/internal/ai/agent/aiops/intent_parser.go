package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"super-agent/internal/ai/models"

	"github.com/cloudwego/eino/schema"
)

// intentParseSystemPrompt 意图解析的 system prompt（few-shot 示例）
const intentParseSystemPrompt = `你是一个运维意图解析助手，将用户的自然语言运维需求解析为 JSON 结构。

规则：
- scene 只能是以下三个值之一：
  * alert_analysis：告警分析（涉及告警、故障、异常等）
  * log_analysis：日志分析（涉及日志、报错、错误信息等）
  * perf_analysis：性能分析（涉及性能、CPU、内存、延迟、吞吐量等）
  * 无法判断时默认使用 alert_analysis
- time_range 格式为数字+单位，单位只能是 m（分钟）、h（小时）、d（天），如 30m、2h、1d；用户未提及时间时默认 1h
- target 为服务名、告警名或实例名等具体分析对象；用户未提及具体目标时为空字符串
- extra 为其他补充信息（如特定错误码、关键词等）；无补充信息时为空字符串
- 只输出合法的 JSON，不要有任何其他内容、解释或 markdown 代码块

示例：
用户：分析 order-service 最近 2 小时的告警
输出：{"scene":"alert_analysis","target":"order-service","time_range":"2h","extra":""}

用户：查看 payment 服务昨天的错误日志
输出：{"scene":"log_analysis","target":"payment","time_range":"1d","extra":"昨天"}

用户：检查 api-gateway 的 CPU 和内存使用情况，最近 30 分钟
输出：{"scene":"perf_analysis","target":"api-gateway","time_range":"30m","extra":""}

用户：帮我看看现在有哪些告警
输出：{"scene":"alert_analysis","target":"","time_range":"1h","extra":""}

用户：user-service 最近频繁出现 500 错误，帮我排查一下日志
输出：{"scene":"log_analysis","target":"user-service","time_range":"1h","extra":"500错误"}`

// parsedIntentJSON LLM 输出的 JSON 结构（仅用于反序列化）
type parsedIntentJSON struct {
	Scene     string `json:"scene"`
	Target    string `json:"target"`
	TimeRange string `json:"time_range"`
	Extra     string `json:"extra"`
}

// ParseIntent 调用 LLM 将自然语言解析为 Options。
// 解析失败时不返回错误，而是 fallback 为默认 Options（全量告警分析），保证主流程不中断。
func ParseIntent(ctx context.Context, query string) (Options, error) {
	llm, err := models.GetDeepSeekV3QuickWithOpenAI(ctx)
	if err != nil {
		return defaultOptions(), fmt.Errorf("ParseIntent: 初始化模型失败: %w", err)
	}

	msgs := []*schema.Message{
		schema.SystemMessage(intentParseSystemPrompt),
		schema.UserMessage(query),
	}

	resp, err := llm.Generate(ctx, msgs)
	if err != nil {
		return defaultOptions(), fmt.Errorf("ParseIntent: LLM 调用失败: %w", err)
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

// defaultOptions 返回默认 Options（全量告警分析，时间范围 1h）
func defaultOptions() Options {
	return Options{
		Scene:     defaultScene,
		TimeRange: "1h",
	}
}
