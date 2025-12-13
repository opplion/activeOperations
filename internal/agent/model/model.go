package model

import (
	"activeOperations/config"
	"activeOperations/internal/agent/tools"
	"context"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

var (
	once      sync.Once
	agent     *react.Agent
	initErr   error
)

func LoadModel() error {
	once.Do(func() {
		ctx := context.Background()
		ChatModel, Err := qwen.NewChatModel(context.Background(), &qwen.ChatModelConfig{
			BaseURL: strings.TrimSpace("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			APIKey:  config.GetConfig().Model.Apikey,
			Timeout: 0,
			Model:   config.GetConfig().Model.Model,
		})
		if Err != nil {
			initErr = Err
			return
		}
		toolsList, Err := tools.GetK8sTools()
		if Err != nil {
			initErr = Err
			return
		}
		agent, Err = react.NewAgent(ctx, &react.AgentConfig{
			Model:       ChatModel,
			ToolsConfig: compose.ToolsNodeConfig{Tools: toolsList},
		})
		if Err != nil {
			initErr = Err
			return 
		}
	})
	return initErr
}

func GetAgent() (*react.Agent, error) {
	return agent, nil
}
