package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Tittifer/IEEE/honeypoint_client/client"
)

// 全局变量
var (
	honeypointClient *client.HoneypointClient
	scanner          *bufio.Scanner
)

func main() {
	log.Println("启动后台客户端...")

	// 创建后台客户端
	var err error
	honeypointClient, err = client.NewHoneypointClient()
	if err != nil {
		log.Fatalf("创建后台客户端失败: %v", err)
	}
	defer honeypointClient.Close()

	// 创建输入扫描器
	scanner = bufio.NewScanner(os.Stdin)

	// 如果有命令行参数，按照原来的方式处理
	if len(os.Args) > 1 {
		handleCommandLineArgs()
		return
	}

	// 否则启动交互式界面
	runInteractiveMode()
}

// 处理命令行参数
func handleCommandLineArgs() {
	command := os.Args[1]
	switch command {
	case "listen":
		startEventListener()
	default:
		printUsage()
	}
}

// 运行交互式界面
func runInteractiveMode() {
	for {
		showMainMenu()
		choice := getUserInput("请选择功能")

		switch choice {
		case "1":
			startEventListener()
			return // 监听模式会一直运行，直到用户中断
		case "2":
			fmt.Println("感谢使用，再见！")
			return
		default:
			fmt.Println("无效的选择，请重新输入")
		}

		fmt.Println("\n按回车键继续...")
		scanner.Scan()
	}
}

// 显示主菜单
func showMainMenu() {
	fmt.Println("\n========= 后台客户端 =========")
	fmt.Println("1. 启动事件监听")
	fmt.Println("2. 退出")
	fmt.Println("============================")
}

// 启动事件监听
func startEventListener() {
	fmt.Println("\n===== 启动事件监听 =====")
	fmt.Println("正在监听链码事件，按Ctrl+C退出...")

	// 设置退出信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动事件监听
	eventCh, err := honeypointClient.StartEventListener()
	if err != nil {
		log.Fatalf("启动事件监听失败: %v", err)
	}

	// 处理事件和退出信号
	for {
		select {
		case event := <-eventCh:
			log.Printf("收到事件: %s, DID: %s, 用户名: %s", event.EventType, event.DID, event.Name)
			honeypointClient.ProcessEvent(event)
		case <-quit:
			log.Println("收到退出信号，关闭后台客户端...")
			return
		}
	}
}

// 获取用户输入
func getUserInput(prompt string) string {
	fmt.Printf("%s: ", prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// 打印使用说明
func printUsage() {
	fmt.Println("后台客户端使用说明:")
	fmt.Println("  启动事件监听: go run main.go listen")
	fmt.Println("  或者直接运行 go run main.go 进入交互式界面")
}