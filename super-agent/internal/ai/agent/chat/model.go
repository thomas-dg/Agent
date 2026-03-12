package chat

import (
	"context"
	"super-agent/internal/ai/models"

	"github.com/cloudwego/eino/components/model"
)

func newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	// TODO Modify component configuration here.
	cm, err := models.GetDeepSeekV3QuickWithOpenAI(ctx)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
