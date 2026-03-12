package chat

import (
	"context"
	"errors"
	"io"
	"strings"
	v1 "super-agent/api/chat/v1"
	agentchat "super-agent/internal/ai/agent/chat"
	"super-agent/utils/callback"
	"super-agent/utils/mem"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) ChatStream(ctx context.Context, req *v1.ChatStreamReq) (res *v1.ChatStreamRes, err error) {
	id := req.SessionID
	msg := req.Question

	ctx = context.WithValue(ctx, "client_id", req.SessionID)
	client, err := c.service.CreateConnection(ctx, g.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	userMessage := &agentchat.UserMessage{
		ID:      id,
		Query:   msg,
		History: mem.GetSimpleMemory(id).GetMessages(),
	}

	sr, err := c.runner.Stream(ctx, userMessage, compose.WithCallbacks(callback.LogCallback(nil)))
	if err != nil {
		client.SendMessage("error", err.Error())
		return nil, err
	}
	defer sr.Close()

	var fullResponse strings.Builder

	defer func() {
		completeResponse := fullResponse.String()
		if completeResponse != "" {
			mem.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
			mem.GetSimpleMemory(id).SetMessages(schema.SystemMessage(completeResponse))
		}
	}()

	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			client.SendMessage("done", "Stream completed")
			return &v1.ChatStreamRes{}, nil
		}
		if err != nil {
			client.SendMessage("error", err.Error())
			return &v1.ChatStreamRes{}, nil
		}
		fullResponse.WriteString(chunk.Content)
		client.SendMessage("message", chunk.Content)
	}
}
