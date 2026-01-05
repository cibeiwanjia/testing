// 原main函数所在包（假设为main包）
package testing

import (
	"fmt"
	"gitee.com/yuanlongxiang123/testing/config"
	"gitee.com/yuanlongxiang123/testing/handler"
	"gitee.com/yuanlongxiang123/testing/middleware"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ====================== 核心封装：服务初始化与启动 ======================
// Server 封装HTTP服务的所有配置和依赖
type Server struct {
	engine *gin.Engine
	port   string
}

// NewServer 创建一个新的HTTP服务实例
func NewServer(port string) *Server {
	// 1. 初始化Gin引擎
	r := gin.Default()

	// 2. 添加全局Sentinel中间件
	r.Use(middleware.SentinelMiddleware())

	// 3. 注册所有路由
	registerRoutes(r)

	return &Server{
		engine: r,
		port:   port,
	}
}

// Start 启动HTTP服务
func (s *Server) Start() error {
	log.Printf("服务器启动在 %s 端口", s.port)
	return s.engine.Run(s.port)
}

// ====================== 封装：路由注册逻辑 ======================
// registerRoutes 注册所有API路由
func registerRoutes(r *gin.Engine) {
	// 1. 创建处理器实例
	orderHandler := handler.NewOrderHandler()

	// 2. 定义API分组
	apiGroup := r.Group("/")
	{
		// 订单接口 - 使用带降级的中间件
		apiGroup.POST("/orders",
			middleware.SentinelMiddlewareWithFallback(orderHandler.FallbackHandler),
			orderHandler.CreateOrder)

		// 获取订单详情 - 使用普通中间件
		apiGroup.GET("/orders/:id", orderHandler.GetOrder)

		// 秒杀接口 - 更严格的限流
		apiGroup.POST("/seckill", orderHandler.Seckill)

		// 商品接口
		apiGroup.GET("/products", productListHandler)

		// 动态更新流控规则的接口（仅用于演示，生产环境需要加权限控制）
		apiGroup.POST("/sentinel/rule", updateFlowRule)

		// 健康检查接口
		apiGroup.GET("/health", healthCheckHandler)
	}
}

// ====================== 封装：各类接口处理器（剥离main函数） ======================
// productListHandler 商品列表查询处理器
func productListHandler(c *gin.Context) {
	// 模拟商品查询
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(20)+5) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    []string{"product1", "product2", "product3"},
	})
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// updateFlowRule 动态更新流控规则的接口
func updateFlowRule(c *gin.Context) {
	var request struct {
		Resource  string  `json:"resource" binding:"required"`
		Threshold float64 `json:"threshold" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 更新规则
	if err := config.UpdateRule(request.Resource, request.Threshold); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("更新规则失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "规则更新成功",
		"resource":  request.Resource,
		"threshold": request.Threshold,
	})
}
