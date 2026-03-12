package aiops

import (
	"context"
	"super-agent/internal/ai/models"
	"super-agent/internal/ai/mytool"

	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/compose"
)

// maxExecutorIterations 限制 Executor 最大迭代次数，防止异常时无限循环消耗资源
const maxExecutorIterations = 30

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	tools := []tool.BaseTool{
		mytool.NewGetCurrentTimeTool(),
		mytool.NewQueryInternalDocsTool(),
		mytool.NewPrometheusAlertsQueryTool(),
	}
	logTools, _ := mytool.GetLogMcpTool(ctx)
	tools = append(tools, logTools...)

	execModel, err := models.GetDeepSeekV3QuickWithOpenAI(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: execModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: maxExecutorIterations,
	})
}
