package chat

import (
	"context"
	"super-agent/internal/ai/mytool"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// newReactAgentLambda component initialization function of node 'ReactAgent' in graph 'chat'
func newReactAgentLambda(ctx context.Context) (lba *compose.Lambda, err error) {
	// TODO Modify component configuration here.
	config := &react.AgentConfig{
		MaxStep:            25,
		ToolReturnDirectly: map[string]struct{}{},
	}
	chatModelIns11, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	config.ToolCallingModel = chatModelIns11
	tools := []tool.BaseTool{
		mytool.NewGetCurrentTimeTool(),
		mytool.NewQueryInternalDocsTool(),
		mytool.NewPrometheusAlertsQueryTool(),
	}
	mcpTool, err := mytool.GetLogMcpTool(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, mcpTool...)
	config.ToolsConfig.Tools = tools

	ins, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
