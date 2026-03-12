package chat

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate' in graph 'chat'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{
		FormatType: schema.FString, //  使用{变量名} 的占位符语法来标记需要动态替换的内容
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", false), // 历史对话（动态插入）
			//schema.UserMessage("{content}"),
			schema.UserMessage("参考文档：\n{documents}\n\n问题：{content}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}

var systemPrompt = `
# 角色：对话小助手
## 核心能力
- 上下文理解与对话
- 知识检索
- 搜索网络获得信息
## 互动指南
- 在回复前，请确保：
	- 完全理解用户的需求和问题，如果有不清楚的地方，需要反复跟用户确认，直至完全理解
	- 考虑最合适的解决方案
- 提供帮助时：
	- 回答清晰简洁
	- 适当提供例子
	- 有帮助时参考文档
	- 适用时建议改进或下一步操作
- 如果用户问题超出能力范围：
	- 清晰说明局限性，如果可能，建议其他方法
- 如果问题是复合或复杂的，你需要一步步思考，避免直接给出质量不高的回答。
## 输出要求
- 简洁明了
- 结构良好，必要时换行
- 不包含markdown的语法，直接输出纯文本
## 上下文信息
- 当前日期：{date}
`
