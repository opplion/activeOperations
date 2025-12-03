package controller

import (
	"activeOperations/internal/agent/graph"
	"github.com/gin-gonic/gin"
	"context"
)

type ChatRequest struct {
    Query string `json:"query"`
}

func StartChat(c *gin.Context) {
	ctx:= context.Background()
	wf, _ := graph.GetWorkflow()
	var query ChatRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	msg,err:=wf.Invoke(ctx, query.Query)
	if err!=nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": msg})
}