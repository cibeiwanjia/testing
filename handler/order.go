package handler

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct{}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// 模拟业务处理时间
	time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)

	// 这里应该是实际的订单创建逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "订单创建成功",
		"data": gin.H{
			"order_id":   generateOrderID(),
			"order_time": time.Now().Format("2006-01-02 15:04:05"),
			"amount":     rand.Float64()*1000 + 1,
		},
		"timestamp": time.Now().Unix(),
	})
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	// 模拟业务处理
	time.Sleep(time.Duration(rand.Intn(30)+5) * time.Millisecond)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取订单成功",
		"data": gin.H{
			"order_id":   orderID,
			"status":     "已支付",
			"amount":     299.99,
			"created_at": time.Now().Add(-time.Hour * 2).Format("2006-01-02 15:04:05"),
		},
		"timestamp": time.Now().Unix(),
	})
}

// Seckill 秒杀接口
func (h *OrderHandler) Seckill(c *gin.Context) {
	productID := c.Query("product_id")

	// 模拟秒杀处理时间
	time.Sleep(time.Duration(rand.Intn(100)+20) * time.Millisecond)

	// 模拟秒杀结果
	success := rand.Intn(100) > 30 // 70%成功率

	if success {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "秒杀成功",
			"data": gin.H{
				"product_id": productID,
				"order_id":   generateOrderID(),
			},
			"timestamp": time.Now().Unix(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":      1001,
			"message":   "秒杀失败，商品已售罄",
			"data":      nil,
			"timestamp": time.Now().Unix(),
		})
	}
}

// FallbackHandler 降级处理函数
func (h *OrderHandler) FallbackHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "降级处理：服务繁忙，已为您排队，请稍后查看订单",
		"data":      nil,
		"timestamp": time.Now().Unix(),
	})
}

// generateOrderID 生成订单ID
func generateOrderID() string {
	return time.Now().Format("20060102150405") + string([]byte{
		byte(rand.Intn(10) + '0'),
		byte(rand.Intn(10) + '0'),
		byte(rand.Intn(10) + '0'),
		byte(rand.Intn(10) + '0'),
	})
}
