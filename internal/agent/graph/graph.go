package graph

import (
	"activeOperations/internal/agent/model"
	"activeOperations/internal/agent/rag"
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

var (
	globalGraph compose.Runnable[string, *schema.Message]
	globalGraphErr error
	onceGraph   sync.Once
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
	log.Printf("Merged message content: %s", msg.Content)
	return []*schema.Message{msg}, nil
}

func GetWorkflow() (compose.Runnable[string, *schema.Message], error) {
	onceGraph.Do(func() {
		ctx := context.Background()
		var err error
		wf := compose.NewWorkflow[string, *schema.Message]()
		// _ = g.AddChatModelNode("QwenModel",model.ChatModel)
		// _ = g.AddChatTemplateNode("Template",model.NewRAGTemplate())
		wf.AddLambdaNode("MilvusRetriever", compose.InvokableLambda(func(ctx context.Context, query string) (*input, error) {
			docs, err := rag.MilvusSDK.Retrieve(ctx, query)
			if err != nil {
				return nil, err
			}
			return &input{
				query: query,
				docs:  docs,
			}, nil
		})).AddInput(compose.START)
		wf.AddLambdaNode("merge", compose.InvokableLambda(lambdaHandler)).AddInput("MilvusRetriever")
		wf.AddChatModelNode("QwenModel", model.ChatModel).AddInput("merge")
		wf.End().AddInput("QwenModel")

		globalGraph,err = wf.Compile(ctx)
		if err!=nil {
			globalGraphErr = fmt.Errorf("failed to compile graph: %v", err)
			return
		}
	})
	return globalGraph,globalGraphErr
}

func ReloadRAG() error {
	ctx:= context.Background()
	chain:= compose.NewChain[document.Source, []string]()
	_ = chain.AppendLoader(rag.Loader)
	_ = chain.AppendDocumentTransformer(rag.Splitter)
	_ = chain.AppendIndexer(rag.MilvusSDK)
	run,err := chain.Compile(ctx)
	if err!=nil {
		return err
	}
	err = rag.ReloadRAG()
	if err!=nil {
		return err
	}
	var files []string
	root := "./website/content/zh-cn/docs/concepts"
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	for i, path := range files {
		_, err := run.Invoke(ctx, document.Source{URI: path})
		fmt.Printf("Processing file %d/%d: %s\n", i+1, len(files), path)
		if err!=nil {
			return err
		}
	}
	return nil
}

// ProcessMessage 处理用户消息并生成回复
func ProcessMessage(ctx context.Context, message, sessionID string) (string, error) {
	// 获取工作流
	workflow, err := GetWorkflow()
	if err != nil {
		return "", fmt.Errorf("获取工作流失败: %v", err)
	}

	// 执行工作流处理消息
	response, err := workflow.Invoke(ctx, message)
	if err != nil {
		return "", fmt.Errorf("执行工作流失败: %v", err)
	}

	// 检查响应是否为空
	if response == nil || response.Content == "" {
		return "抱歉，我无法生成回复", nil
	}

	return response.Content, nil
}