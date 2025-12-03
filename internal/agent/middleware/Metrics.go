package middleware

import (
	"context"
	"strings"
	"google.golang.org/grpc"
	"github.com/gin-gonic/gin"
	"log"
)

func HttpMetrics(app *gin.Engine) *gin.Engine {
	app.Use(func(c *gin.Context) {
		api := extractAPIName(c.Request.RequestURI)

		c.Next()

		status := c.Writer.Status()
		result := "ok"
		if status >= 400 {
			result = "fail"
		}
		log.Printf("[HTTP] api=%s status=%d result=%s", api, status, result)
	})

	return app
}

func GrpcMetrics(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {

	api := extractAPIName(info.FullMethod)
	resp, err = handler(ctx, req)

	result := "ok"
	if err != nil {
		result = "fail"
	}
	log.Printf("[grpc] api=%s result=%s", api, result)

	return resp, err
}

func extractAPIName(path string) string {
	if path == "" {
		return "unknown"
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.ToLower(parts[len(parts)-1])
}
