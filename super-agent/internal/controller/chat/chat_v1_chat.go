package chat

//
//import (
//	"context"
//	v1 "super-agent/api/chat/v1"
//	"super-agent/utils/callback"
//	"super-agent/utils/mem"
//
//	agentchat "super-agent/internal/ai/agent/chat"
//
//	"github.com/cloudwego/eino/compose"
//	"github.com/cloudwego/eino/schema"
//)
//
//func (c *ControllerV1) Chat(ctx context.Context, req *v1.ChatReq) (res *v1.ChatRes, err error) {
//	id := req.ID
//	msg := req.Question
//
//	userMessage := &agentchat.UserMessage{
//		ID:      id,
//		Query:   msg,
//		History: mem.GetSimpleMemory(id).GetMessages(),
//	}
//
//	out, err := c.runner.Invoke(ctx, userMessage, compose.WithCallbacks(callback.LogCallback(nil)))
//	if err != nil {
//		return nil, err
//	}
//	res = &v1.ChatRes{
//		Answer: out.Content,
//	}
//	mem.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
//	mem.GetSimpleMemory(id).SetMessages(schema.SystemMessage(out.Content))
//
//	return res, nil
//}
