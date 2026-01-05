package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// SentinelMiddleware Sentinel限流中间件
func SentinelMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用请求路径作为资源名称
		resource := c.FullPath()
		if resource == "" {
			resource = c.Request.URL.Path
		}

		// 添加请求方法作为前缀，区分不同HTTP方法的同一路径
		resource = fmt.Sprintf("%s:%s", c.Request.Method, resource)

		// 尝试通过Sentinel Entry
		entry, blockErr := api.Entry(
			resource,
			api.WithTrafficType(base.Inbound),
			api.WithResourceType(base.ResTypeWeb),
		)

		if blockErr != nil {
			// 请求被限流，返回429状态码
			fmt.Println("请求被限流：" + c.Request.URL.Path + "请求过于频繁，请稍后再试")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":      429,
				"message":   "请求过于频繁，请稍后再试",
				"data":      nil,
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}

		// 请求通过，记录entry到上下文，在请求结束后退出
		c.Set("sentinelEntry", entry)

		// 继续处理请求
		c.Next()

		// 获取entry并退出
		if entryVal, exists := c.Get("sentinelEntry"); exists {
			if entry, ok := entryVal.(*base.SentinelEntry); ok {
				entry.Exit()
			}
		}
	}
}

// SentinelMiddlewareWithFallback 带降级处理的Sentinel中间件
func SentinelMiddlewareWithFallback(fallback gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		resource := c.FullPath()
		if resource == "" {
			resource = c.Request.URL.Path
		}
		resource = fmt.Sprintf("%s:%s", c.Request.Method, resource)

		entry, blockErr := api.Entry(resource)

		if blockErr != nil {
			// 如果提供了降级处理函数，则执行降级逻辑
			if fallback != nil {
				fallback(c)
			} else {
				// 默认降级响应
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":      429,
					"message":   "系统繁忙，请稍后再试",
					"data":      nil,
					"timestamp": time.Now().Unix(),
				})
			}
			c.Abort()
			return
		}

		c.Set("sentinelEntry", entry)
		defer func() {
			if entryVal, exists := c.Get("sentinelEntry"); exists {
				if entry, ok := entryVal.(*base.SentinelEntry); ok {
					entry.Exit()
				}
			}
		}()

		c.Next()
	}
}

// ParamBasedSentinelMiddleware 基于参数的Sentinel中间件
func ParamBasedSentinelMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource string

		// 如果是POST/PUT请求，尝试从body获取参数
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var body map[string]interface{}
			if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil {
				if val, ok := body[paramName].(string); ok && val != "" {
					resource = fmt.Sprintf("%s:%s:%s", c.Request.Method, c.FullPath(), val)
				}
			}
		}

		// 如果从body没获取到，尝试从query获取
		if resource == "" {
			val := c.Query(paramName)
			if val != "" {
				resource = fmt.Sprintf("%s:%s:%s", c.Request.Method, c.FullPath(), val)
			}
		}

		// 如果还是没有获取到参数，使用默认资源名
		if resource == "" {
			resource = fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())
		}

		entry, blockErr := api.Entry(resource)

		if blockErr != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":      429,
				"message":   "请求过于频繁，请稍后再试",
				"data":      nil,
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Set("sentinelEntry", entry)
		c.Next()

		if entryVal, exists := c.Get("sentinelEntry"); exists {
			if entry, ok := entryVal.(*base.SentinelEntry); ok {
				entry.Exit()
			}
		}
	}
}
