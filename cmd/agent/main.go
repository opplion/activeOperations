package main

import (
	"activeOperations/config"
	"activeOperations/internal/agent/graph"
	"activeOperations/internal/agent/model"
	"activeOperations/internal/agent/rag"
	"activeOperations/internal/agent/router"
	"net/http"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.Println("🚀 应用启动中...")

	// 同步加载配置（必须先完成！）
	log.Println("📥 正在加载配置文件 ./config.yaml")
	config.LoadConfig("./config.yaml") 
	log.Println("✅ 配置加载成功")

	var g errgroup.Group

	// 并发初始化各模块
	g.Go(func() error {
		log.Println("【LoaderInit】开始初始化文档加载器...")
		err := rag.LoaderInit()
		if err != nil {
			log.Printf("【LoaderInit】失败: %v", err)
			return err
		}
		log.Println("【LoaderInit】完成")
		return nil
	})

	g.Go(func() error {
		log.Println("【LoadModel】开始加载大语言模型（可能耗时较长）...")
		err := model.LoadModel()
		if err != nil {
			log.Printf("【LoadModel】失败: %v", err)
			return err
		}
		log.Println("【LoadModel】模型加载完成")
		return nil
	})

	g.Go(func() error {
		log.Println("【MilvusInit】开始初始化 RAG 相关组件...")
		rag.NewSchema()
		if err := rag.LoadEmbedder(); err != nil {
			log.Printf("【MilvusInit】Embedder 加载失败: %v", err)
			return err
		}
		if err := rag.MilvusInit(); err != nil {
			log.Printf("【MilvusInit】Milvus 初始化失败: %v", err)
			return err
		}
		log.Println("【MilvusInit】完成")
		return nil
	})

	// 等待所有初始化完成
	log.Println("⏳ 等待所有模块初始化完成...")
	if err := g.Wait(); err != nil {
		log.Fatal("❌ 初始化失败:", err)
	}

	log.Println("✅ 所有初始化完成，开始加载 RAG 数据...")
	if err := graph.ReloadRAG(); err != nil {
		log.Fatalf("💥 ReloadRAG 失败: %v", err)
	}
	_,err := graph.GetWorkflow()
	if err!=nil {
		log.Fatalf("💥 获取工作流失败: %v", err)
	}

	app:= router.StartServer()

	server := &http.Server{
        Addr:    fmt.Sprintf(":%s", config.GetConfig().HTTPPort),
        Handler: app,
    }
	go func(){
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("服务异常退出: %v", err)
        }
	}()

	log.Println("🎉 应用启动成功！")
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("🛑 收到关闭信号，正在优雅关闭...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("优雅关闭失败: %v", err)
    }

    log.Println("✅ 服务已安全关闭")
}