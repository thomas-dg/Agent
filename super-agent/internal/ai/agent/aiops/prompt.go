package aiops

import "fmt"

// Scene 定义 AI 运维分析场景
type Scene string

const (
	// SceneAlertAnalysis 告警分析（默认）
	SceneAlertAnalysis Scene = "alert_analysis"
	// SceneLogAnalysis 日志分析
	SceneLogAnalysis Scene = "log_analysis"
	// ScenePerfAnalysis 性能分析
	ScenePerfAnalysis Scene = "perf_analysis"
)

// defaultScene 未指定时的默认场景
const defaultScene = SceneAlertAnalysis

// systemPrompts 按场景维护 system prompt 模板
var systemPrompts = map[Scene]string{
	SceneAlertAnalysis: `
你是一个智能的服务告警运维分析助手，处理告警时必须遵循以下规则：
1. 首先调用工具 query_prometheus_alerts 获取所有活跃的告警。
2. 分别根据告警名称调用工具 query_internal_docs，获取对应的处理方案。
3. 完全遵循内部文档的内容进行查询和分析，不允许使用文档外的任何信息。
4. 涉及时间的参数，必须先通过工具 get_current_time 获取当前时间，再结合用户要求传参。
5. 涉及日志查询，需先通过日志工具获取相关日志，参数必须携带地域和日志主题。
6. 分别将告警对应查询到的信息进行总结分析，最后汇总所有告警和总结。

输出规则：
- 简洁明了，不重复已知信息，不做无意义的过渡说明
`,
	SceneLogAnalysis: `
你是一个智能的日志分析助手，分析日志时必须遵循以下规则：
1. 必须先通过工具 get_current_time 获取当前时间，再结合用户指定的时间范围确定查询区间。
2. 通过日志工具查询相关日志，参数必须携带地域和日志主题。
3. 根据告警名称或服务名调用工具 query_internal_docs，获取对应的排查方案。
4. 完全遵循内部文档的内容进行分析，不允许使用文档外的任何信息。
5. 对日志中的异常、错误、慢请求进行归类分析，给出根因判断和处理建议。

输出规则：
- 简洁明了，不重复已知信息，不做无意义的过渡说明
`,
	ScenePerfAnalysis: `
你是一个智能的性能分析助手，分析性能问题时必须遵循以下规则：
1. 首先调用工具 query_prometheus_alerts 获取当前活跃的性能相关告警。
2. 必须先通过工具 get_current_time 获取当前时间，再结合用户指定的时间范围确定查询区间。
3. 根据服务名或指标名调用工具 query_internal_docs，获取对应的性能基线和处理方案。
4. 完全遵循内部文档的内容进行分析，不允许使用文档外的任何信息。
5. 对 CPU、内存、延迟、吞吐量等关键指标进行综合分析，给出性能瓶颈判断和优化建议。

输出规则：
- 简洁明了，不重复已知信息，不做无意义的过渡说明
`,
}

// queryTemplates 按场景维护任务描述模板，%s 占位符依次为：target、timeRange
var queryTemplates = map[Scene]string{
	SceneAlertAnalysis: `
请对当前活跃告警进行全面分析。%s%s
分析步骤：
1. 调用工具 query_prometheus_alerts 获取所有活跃告警。
2. 分别根据告警名称调用工具 query_internal_docs 获取处理方案。
3. 涉及时间参数先调用 get_current_time 获取当前时间。
4. 涉及日志查询，通过日志工具获取相关日志，参数携带地域和日志主题。
5. 汇总生成告警运维分析报告，格式如下：
告警分析报告
---
# 告警处理详情
## 活跃告警清单
## 告警根因分析N（第N个告警）
## 处理方案执行N（第N个告警）
## 结论
`,
	SceneLogAnalysis: `
请对相关日志进行深入分析。%s%s
分析步骤：
1. 先调用 get_current_time 获取当前时间，确定查询时间区间。
2. 通过日志工具查询相关日志，参数携带地域和日志主题。
3. 根据服务名或告警名调用 query_internal_docs 获取排查方案。
4. 对异常、错误、慢请求进行归类，给出根因判断和处理建议。
`,
	ScenePerfAnalysis: `
请对服务性能进行全面分析。%s%s
分析步骤：
1. 调用 query_prometheus_alerts 获取当前性能相关告警。
2. 先调用 get_current_time 获取当前时间，确定查询时间区间。
3. 根据服务名调用 query_internal_docs 获取性能基线和处理方案。
4. 对 CPU、内存、延迟、吞吐量等关键指标进行综合分析，给出优化建议。
`,
}

// Options 构建 AIOps Agent 的选项
type Options struct {
	// Scene 分析场景，为空时使用 SceneAlertAnalysis
	Scene Scene
	// Target 分析目标（服务名/告警名/实例名），为空表示全量分析
	Target string
	// TimeRange 时间范围，如 "1h"、"30m"，为空默认 "1h"
	TimeRange string
	// Extra 额外上下文或补充说明
	Extra string
}

// normalize 补全默认值
func (o *Options) normalize() {
	if o.Scene == "" {
		o.Scene = defaultScene
	}
	if o.TimeRange == "" {
		o.TimeRange = "1h"
	}
}

// BuildSystemPrompt 根据场景返回对应的 system prompt
func BuildSystemPrompt(opts Options) string {
	opts.normalize()
	if p, ok := systemPrompts[opts.Scene]; ok {
		return p
	}
	return systemPrompts[defaultScene]
}

// BuildQuery 根据选项构建任务描述
func BuildQuery(opts Options) string {
	opts.normalize()

	tmpl, ok := queryTemplates[opts.Scene]
	if !ok {
		tmpl = queryTemplates[defaultScene]
	}

	// 构造 target 和 timeRange 描述片段
	targetDesc := ""
	if opts.Target != "" {
		targetDesc = fmt.Sprintf("分析目标：%s。", opts.Target)
	}
	timeDesc := fmt.Sprintf("时间范围：最近 %s。", opts.TimeRange)

	query := fmt.Sprintf(tmpl, targetDesc, timeDesc)

	// 追加额外上下文
	if opts.Extra != "" {
		query += fmt.Sprintf("\n补充说明：%s\n", opts.Extra)
	}
	return query
}
