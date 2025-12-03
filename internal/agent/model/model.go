package model

import (
	"context"
	"sync"
	"activeOperations/config"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"time"
	"strings"
)

var (
	once      sync.Once
	ChatModel model.ToolCallingChatModel
	initErr   error
)

func LoadModel() error {
	once.Do(func() {
		ChatModel, initErr = qwen.NewChatModel(context.Background(), &qwen.ChatModelConfig{
			BaseURL: strings.TrimSpace("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			APIKey:  config.GetConfig().Model.Apikey,
			Timeout: 15 * time.Second,
			Model:   config.GetConfig().Model.Model,
		})
	})
	return initErr
}
