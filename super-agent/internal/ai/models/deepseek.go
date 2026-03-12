package models

import (
	"context"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/gogf/gf/v2/frame/g"
)

// quickModelOnce 保证 DeepSeekV3Quick 客户端全局只初始化一次
var (
	quickModelOnce     sync.Once
	quickModelInstance model.ToolCallingChatModel
	quickModelErr      error
)

// thinkModelOnce 保证 DeepSeekV3Think 客户端全局只初始化一次
var (
	thinkModelOnce     sync.Once
	thinkModelInstance model.ToolCallingChatModel
	thinkModelErr      error
)

// GetDeepSeekV3QuickWithOpenAI 返回 DeepSeekV3Quick 的单例 LLM 客户端。
// 首次调用时读取配置并初始化，后续调用直接返回缓存实例。
func GetDeepSeekV3QuickWithOpenAI(ctx context.Context) (model.ToolCallingChatModel, error) {
	quickModelOnce.Do(func() {
		modelName, err := g.Cfg().Get(ctx, "ds_quick_chat_model.model")
		if err != nil {
			quickModelErr = err
			return
		}
		apiKey, err := g.Cfg().Get(ctx, "ds_quick_chat_model.api_key")
		if err != nil {
			quickModelErr = err
			return
		}
		baseURL, err := g.Cfg().Get(ctx, "ds_quick_chat_model.base_url")
		if err != nil {
			quickModelErr = err
			return
		}
		cfg := &openai.ChatModelConfig{
			Model:   modelName.String(),
			APIKey:  apiKey.String(),
			BaseURL: baseURL.String(),
		}
		quickModelInstance, quickModelErr = openai.NewChatModel(ctx, cfg)
	})
	return quickModelInstance, quickModelErr
}

// GetDeepSeekV3ThinkWithOpenAI 返回 DeepSeekV3Think 的单例 LLM 客户端。
// 首次调用时读取配置并初始化，后续调用直接返回缓存实例。
func GetDeepSeekV3ThinkWithOpenAI(ctx context.Context) (model.ToolCallingChatModel, error) {
	thinkModelOnce.Do(func() {
		modelName, err := g.Cfg().Get(ctx, "ds_think_chat_model.model")
		if err != nil {
			thinkModelErr = err
			return
		}
		apiKey, err := g.Cfg().Get(ctx, "ds_think_chat_model.api_key")
		if err != nil {
			thinkModelErr = err
			return
		}
		baseURL, err := g.Cfg().Get(ctx, "ds_think_chat_model.base_url")
		if err != nil {
			thinkModelErr = err
			return
		}
		cfg := &openai.ChatModelConfig{
			Model:   modelName.String(),
			APIKey:  apiKey.String(),
			BaseURL: baseURL.String(),
		}
		thinkModelInstance, thinkModelErr = openai.NewChatModel(ctx, cfg)
	})
	return thinkModelInstance, thinkModelErr
}