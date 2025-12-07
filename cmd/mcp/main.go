package main

import (
    "github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/mcp"
	"activeOperations/config"
	"log"
)

func main() {
    // 注册回调，务必先注册
	config.LoadConfig("./config.yaml")
	log.Println("✅ 配置加载成功")

    callbacks.RegisterInit()
		cfg := &mcp.ServerConfig{
		Name:  "mcp-server",
		Port:  4321,
		Mode:  mcp.ServerModeSSE,
		Kubeconfigs: []mcp.KubeconfigConfig{
				{ID: "k8s", Path: "./k8s/k8sUser.yml", IsDefault: true},
			},
		}
	mcp.RunMCPServerWithOption(cfg)
}