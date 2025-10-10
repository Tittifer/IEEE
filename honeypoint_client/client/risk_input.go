package client

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// RiskInputManager 风险行为输入管理器
type RiskInputManager struct {
	honeypointClient *HoneypointClient // 后台客户端引用
	scanner          *bufio.Scanner    // 输入扫描器
	inputChan        chan string       // 输入通道
	stopChan         chan struct{}     // 停止通道
	wg               sync.WaitGroup    // 等待组
	running          bool              // 是否正在运行
	mutex            sync.Mutex        // 互斥锁
}

// NewRiskInputManager 创建新的风险行为输入管理器
func NewRiskInputManager(client *HoneypointClient) *RiskInputManager {
	return &RiskInputManager{
		honeypointClient: client,
		scanner:          bufio.NewScanner(os.Stdin),
		inputChan:        make(chan string, 10),
		stopChan:         make(chan struct{}),
	}
}

// Start 启动风险行为输入监听
func (r *RiskInputManager) Start() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.running {
		log.Println("风险行为输入监听已经在运行")
		return
	}
	
	r.running = true
	r.wg.Add(1)
	
	// 启动输入监听协程
	go r.listenForInput()
	
	log.Println("风险行为输入监听已启动，请输入用户DID和风险行为（格式：DID:行为类型）")
	log.Println("例如：did:example:123:A 或 did:example:123:B")
}

// Stop 停止风险行为输入监听
func (r *RiskInputManager) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if !r.running {
		log.Println("风险行为输入监听未在运行")
		return
	}
	
	close(r.stopChan)
	r.wg.Wait()
	r.running = false
	
	log.Println("风险行为输入监听已停止")
}

// listenForInput 监听用户输入的风险行为
func (r *RiskInputManager) listenForInput() {
	defer r.wg.Done()
	
	for {
		select {
		case <-r.stopChan:
			return
		default:
			fmt.Print("请输入用户DID和风险行为（格式：DID:行为类型）: ")
			if r.scanner.Scan() {
				input := strings.TrimSpace(r.scanner.Text())
				if input != "" {
					r.processInput(input)
				}
			}
		}
	}
}

// processInput 处理用户输入的风险行为
func (r *RiskInputManager) processInput(input string) {
	// 查找最后一个冒号的位置，用于分隔DID和行为类型
	lastColonIndex := strings.LastIndex(input, ":")
	if lastColonIndex == -1 || lastColonIndex == len(input)-1 {
		log.Println("输入格式错误，正确格式为：DID:行为类型")
		return
	}
	
	did := strings.TrimSpace(input[:lastColonIndex])
	behaviorStr := strings.TrimSpace(input[lastColonIndex+1:])
	
	// 检查DID是否为空
	if did == "" {
		log.Println("DID不能为空")
		return
	}
	
	// 检查行为类型是否有效
	var behavior RiskBehavior
	switch strings.ToUpper(behaviorStr) {
	case string(RiskBehaviorA):
		behavior = RiskBehaviorA
	case string(RiskBehaviorB):
		behavior = RiskBehaviorB
	default:
		log.Printf("无效的风险行为类型: %s，有效类型为 A 或 B", behaviorStr)
		return
	}
	
	// 处理风险行为，更新风险评分
	newScore, err := r.honeypointClient.riskManager.ProcessRiskBehavior(did, behavior)
	if err != nil {
		log.Printf("处理风险行为失败: %v", err)
		return
	}
	
	log.Printf("用户 %s 的风险行为 %s 已处理，新风险评分: %d", did, behavior, newScore)
	
	// 向链上报告最新风险评分
	if err := r.honeypointClient.riskManager.ReportRiskScore(did); err != nil {
		log.Printf("向链上报告风险评分失败: %v", err)
	}
}