package tools

import (
	"github.com/mark3labs/mcp-go/client"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	 mcpCli "github.com/mark3labs/mcp-go/mcp"
	"github.com/cloudwego/eino/components/tool"
	"context"
)

func GetK8sTools() ([]tool.BaseTool, error) {
	ctx := context.Background()
	cli, err := client.NewSSEMCPClient("http://mcp:4321/sse")
	if err != nil {
		return nil, err
	}
	err = cli.Start(ctx)
	if err != nil {
		return nil, err
	}
	initRequest := mcpCli.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpCli.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpCli.Implementation{
		Name:    "k8s-client",
		Version: "1.0.0",
	}
	_, err = cli.Initialize(ctx, initRequest)
	mcpTools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
	if  err != nil {
		return nil, err
	}
	return mcpTools, nil
}
