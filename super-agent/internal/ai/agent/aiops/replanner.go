package aiops

import (
	"context"
	"super-agent/internal/ai/models"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func NewRePlanAgent(ctx context.Context) (adk.Agent, error) {
	model, err := models.GetDeepSeekV3ThinkWithOpenAI(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model,
	})
}
