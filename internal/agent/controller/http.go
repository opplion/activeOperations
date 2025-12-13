package controller

import (
	"activeOperations/internal/agent/graph"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Query string `json:"query"`
	UUID  string `json:"uuid"`
}

type Task struct {
	cancel context.CancelFunc
}

var (
	TaskPool = make(map[string]*Task)
	mu       sync.Mutex
)

func StartChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if req.UUID == "" {
		c.JSON(400, gin.H{"error": "uuid is required"})
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

	mu.Lock()
	TaskPool[req.UUID] = &Task{cancel: cancel}
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(TaskPool, req.UUID)
		mu.Unlock()
	}()

	wf, err := graph.GetWorkflow()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get workflow"})
		return
	}

	msgReader, err := wf.Invoke(ctx, req.Query)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Chat canceled:", req.UUID)
			fmt.Fprintf(c.Writer, "event: end\ndata: {\"reason\": \"canceled\"}\n\n")
			c.Writer.Flush()
			return

		default:
			msg, err := msgReader.Recv()
			if err != nil {
				if err == io.EOF {
					fmt.Println("Stream ended normally")
					return
				}
				fmt.Printf("Recv error: %v\n", err)
				return
			}

			jsonBytes, _ := json.Marshal(msg)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonBytes)
			c.Writer.Flush()
		}
	}
}

func EndChat(c *gin.Context) {
	var req struct {
		UUID string `json:"uuid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.UUID == "" {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	mu.Lock()
	task, ok := TaskPool[req.UUID]
	mu.Unlock()

	if !ok {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	task.cancel()

	c.JSON(200, gin.H{"status": "chat canceled"})
}
