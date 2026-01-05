package config

import (
	"fmt"
	"log"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/logging"
)

// InitSentinel 初始化Sentinel配置
func InitSentinel() error {
	// 创建配置
	conf := config.NewDefaultConfig()

	// 配置控制台日志输出
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()

	// 设置应用名称
	conf.Sentinel.App.Name = "gin-ecommerce"

	// 初始化Sentinel
	err := api.InitWithConfig(conf)
	if err != nil {
		return fmt.Errorf("初始化Sentinel失败: %v", err)
	}

	// 加载默认的流控规则
	err = loadDefaultRules()
	if err != nil {
		return fmt.Errorf("加载流控规则失败: %v", err)
	}

	log.Println("Sentinel初始化成功")
	return nil
}

// loadDefaultRules 加载默认的流控规则
func loadDefaultRules() error {
	// 定义需要限流的接口资源
	rules := []*flow.Rule{
		// 订单创建接口：QPS限制为100
		{
			Resource:               "POST:/orders",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              5,
			StatIntervalInMs:       1000,
		},
		//// 商品查询接口：QPS限制为200
		//{
		//	Resource:               "/api/v1/products",
		//	TokenCalculateStrategy: flow.Direct,
		//	ControlBehavior:        flow.Reject,
		//	Threshold:              200,
		//	StatIntervalInMs:       1000,
		//},
		//// 秒杀接口：QPS限制为50（更严格的限制）
		//{
		//	Resource:               "/api/v1/seckill",
		//	TokenCalculateStrategy: flow.Direct,
		//	ControlBehavior:        flow.Reject,
		//	Threshold:              50,
		//	StatIntervalInMs:       1000,
		//},
	}

	// 加载规则
	_, err := flow.LoadRules(rules)
	return err
}

// UpdateRule 动态更新流控规则
func UpdateRule(resource string, threshold float64) error {
	// 首先获取现有规则
	existingRules := flow.GetRulesOfResource(resource)

	// 创建新规则
	newRule := flow.Rule{
		Resource:               resource,
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
		Threshold:              threshold,
		StatIntervalInMs:       1000,
	}

	// 如果已存在规则，则替换；否则添加新规则
	if len(existingRules) > 0 {
		existingRules[0] = newRule
	} else {
		existingRules = append(existingRules, newRule)
	}
	// 将[]flow.Rule转换为[]*flow.Rule
	rulesToLoad := make([]*flow.Rule, len(existingRules))
	for i, rule := range existingRules {
		// 创建临时变量以避免获取循环变量的地址
		r := rule
		rulesToLoad[i] = &r
	}
	// 加载更新后的规则
	_, err := flow.LoadRules(rulesToLoad)
	return err
}
