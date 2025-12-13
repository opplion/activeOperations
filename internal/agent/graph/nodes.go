package graph

import (
	"activeOperations/internal/agent/model"
	"activeOperations/internal/agent/rag"
	"context"
	"fmt"
	"strings"
	"github.com/cloudwego/eino/schema"
)


type input struct {
	docs []*schema.Document
	query string
}

func lambdaHandler(_ context.Context, input *input) ([]*schema.Message, error) {
	var contents []string
	for _, doc := range input.docs {
		contents = append(contents, doc.Content)
	}
	knowledge := strings.Join(contents, "\n")
	msgs := fmt.Sprintf("user query: %s\nContext knowledge: %s", input.query, knowledge)
	msgs += "\n请仅根据上述上下文回答问题，不要提及来源或上下文本身。"
	msg:= &schema.Message{
		Role:    schema.User,
		Content: msgs,
	}
	return []*schema.Message{msg}, nil
}

func MilvusRetrieverHandler(ctx context.Context, query string) (*input, error) {
	docs, err := rag.MilvusSDK.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}
	return &input{
		query: query,
		docs:  docs,
	}, nil
}

func ReactAgentHandler(ctx context.Context, query []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	agentModel,err := model.GetAgent()
	if err!=nil {
		return nil, err
	}
	msgReader, err := agentModel.Stream(ctx,query)
	if err != nil {
		return nil, err
	}
	return  msgReader,nil
}