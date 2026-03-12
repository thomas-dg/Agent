package chat

import (
	"context"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino/compose"
)

func BuildChatAgent(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	const (
		InputToChat     = "InputToChat"
		InputToRAG      = "InputToRAG"
		MilvusRetriever = "MilvusRetriever"
		ChatTemplate    = "ChatTemplate"
		ReactAgent      = "ReactAgent"
	)
	g := compose.NewGraph[*UserMessage, *schema.Message]()
	_ = g.AddLambdaNode(InputToChat, compose.InvokableLambdaWithOption(newInputToChatLambda),
		compose.WithNodeName("UserMessageToChat"))
	_ = g.AddLambdaNode(InputToRAG, compose.InvokableLambdaWithOption(newInputToRagLambda),
		compose.WithNodeName("UserMessageToRag"))
	milvusRetrieverKeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(MilvusRetriever, milvusRetrieverKeyOfRetriever, compose.WithOutputKey("documents"))
	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)
	reactAgentKeyOfLambda, err := newReactAgentLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda,
		compose.WithNodeName("ReActAgent"))
	_ = g.AddEdge(compose.START, InputToChat)
	_ = g.AddEdge(compose.START, InputToRAG)
	_ = g.AddEdge(ReactAgent, compose.END)
	_ = g.AddEdge(InputToChat, ChatTemplate)
	_ = g.AddEdge(InputToRAG, MilvusRetriever)
	_ = g.AddEdge(MilvusRetriever, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReactAgent)
	r, err = g.Compile(ctx, compose.WithGraphName("chat"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
