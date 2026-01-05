package test

import (
	"bytes"
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cibeiwanjia/testing/config"
	"github.com/cibeiwanjia/testing/handler"
	"github.com/cibeiwanjia/testing/middleware"
	"github.com/gin-gonic/gin"
)

// 设置测试模式
func init() {
	gin.SetMode(gin.TestMode)
}

// 创建测试服务器
func createTestServer() *gin.Engine {
	// 初始化Sentinel
	config.InitSentinel()

	// 创建Gin引擎
	r := gin.Default()

	// 添加全局Sentinel中间件
	r.Use(middleware.SentinelMiddleware())

	// 创建处理器
	orderHandler := handler.NewOrderHandler()

	// 定义路由
	apiGroup := r.Group("/")
	{
		// 订单接口 - 使用带降级的中间件
		apiGroup.POST("/orders",
			middleware.SentinelMiddlewareWithFallback(orderHandler.FallbackHandler),
			orderHandler.CreateOrder)
	}

	return r
}

// 基准测试：测试单个请求
func BenchmarkCreateOrderSingle(b *testing.B) {
	r := createTestServer()

	// 测试数据
	testData := map[string]interface{}{
		"product_id": "test_product_123",
		"quantity":   1,
		"price":      99.99,
	}

	jsonData, _ := json.Marshal(testData)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建响应记录器
		w := httptest.NewRecorder()
		// 执行请求
		r.ServeHTTP(w, req)
	}
}

// 基准测试：测试并行请求
func BenchmarkCreateOrderParallel(b *testing.B) {
	r := createTestServer()

	// 测试数据
	testData := map[string]interface{}{
		"product_id": "test_product_123",
		"quantity":   1,
		"price":      99.99,
	}

	jsonData, _ := json.Marshal(testData)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()
			// 执行请求
			r.ServeHTTP(w, req)
		}
	})
}

// loadDefaultRules 加载默认的流控规则
func loadDefaultRules() error {
	// 定义需要限流的接口资源
	rules := []*flow.Rule{
		// 订单创建接口：QPS限制为5（降低阈值，方便测试触发限流）
		{
			Resource:               "POST:/api/v1/orders",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              5,
			StatIntervalInMs:       1000,
		},
		// 订单创建接口（测试用，无API前缀）：QPS限制为5
		{
			Resource:               "POST:/orders",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              5,
			StatIntervalInMs:       1000,
		},
		// 商品查询接口：QPS限制为200
		{
			Resource:               "GET:/api/v1/products",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              200,
			StatIntervalInMs:       1000,
		},
		// 秒杀接口：QPS限制为50（更严格的限制）
		{
			Resource:               "POST:/api/v1/seckill",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              50,
			StatIntervalInMs:       1000,
		},
	}

	// 加载规则
	_, err := flow.LoadRules(rules)
	return err
}
