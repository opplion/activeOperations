package controller

import (
	"activeOperations/internal/agent/graph"
	"github.com/gin-gonic/gin"
	"context"
	"fmt"
	"encoding/json"
	"io"
)

type ChatRequest struct {
    Query string `json:"query"`
}

func StartChat(c *gin.Context) {
	ctx := context.Background()
	wf, err := graph.GetWorkflow()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get workflow"})
		return
	}

	var query ChatRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	msgReader, err := wf.Invoke(ctx, query.Query)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Stream ended normally (EOF)")
				break
			}
			fmt.Printf("Error receiving message: %v", err)
			fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())
			break
		}
		jsonBytes, _ := json.Marshal(msg)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonBytes))
		c.Writer.Flush()
	}
}