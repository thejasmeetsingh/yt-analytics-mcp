package mcp

import (
	"context"
	"errors"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"
)

func RateLimiter(limiter *rate.Limiter) goMCP.Middleware {
	return func(next goMCP.MethodHandler) goMCP.MethodHandler {
		return func(ctx context.Context, method string, req goMCP.Request) (goMCP.Result, error) {
			if !limiter.Allow() {
				return nil, errors.New("rate limit exceeded")
			}
			return next(ctx, method, req)
		}
	}
}
