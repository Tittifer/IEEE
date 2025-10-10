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
	log.Println("启动蜜点后台客户端...")

	// 创建蜜点后台客户端
	var err error
	honeypointClient, err = client.NewHoneypointClient()
	if err != nil {
		log.Fatalf("创建蜜点后台客户端失败: %v", err)
	}
	defer honeypointClient.Close()

	// 创建输入扫描器
	scanner = bufio.NewScanner(os.Stdin)

	// 启动事件监听
	if err := honeypointClient.StartEventListener(); err != nil {
		log.Fatalf("启动事件监听失败: %v", err)
	}

	// 处理命令行参数
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
	case "risk":
		if len(os.Args) != 4 {
			fmt.Println("用法: go run main.go risk <DID> <风险行为类型>")
			return
		}
		did := os.Args[2]
		behaviorType := os.Args[3]

		err := honeypointClient.ProcessRiskBehavior(did, behaviorType)
		if err != nil {
			log.Fatalf("处理风险行为失败: %v", err)
		}
		fmt.Println("风险行为处理成功")

	default:
		printUsage()
	}
}

// 运行交互式界面
func runInteractiveMode() {
	// 设置信号处理，优雅退出
	setupSignalHandler()

	fmt.Println("\n========= 蜜点后台客户端 =========")
	fmt.Println("已启动事件监听，等待用户注册和登录事件...")
	fmt.Println("您可以输入命令来模拟用户风险行为")
	fmt.Println("格式: risk <DID> <风险行为类型>")
	fmt.Println("输入 'exit' 或 'quit' 退出程序")
	fmt.Println("===================================")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "" {
			continue
		}

		// 检查退出命令
		if input == "exit" || input == "quit" {
			fmt.Println("正在退出程序...")
			break
		}

		// 解析命令
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "risk":
			if len(parts) != 3 {
				fmt.Println("用法: risk <DID> <风险行为类型>")
				continue
			}
			did := parts[1]
			behaviorType := parts[2]

			err := honeypointClient.ProcessRiskBehavior(did, behaviorType)
			if err != nil {
				fmt.Printf("处理风险行为失败: %v\n", err)
				continue
			}
			fmt.Println("风险行为处理成功")

		case "help":
			printHelp()

		default:
			fmt.Println("未知命令，输入 'help' 查看帮助")
		}
	}
}

// 设置信号处理器
func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n接收到退出信号，正在关闭...")
		honeypointClient.Close()
		os.Exit(0)
	}()
}

// 打印帮助信息
func printHelp() {
	fmt.Println("\n可用命令:")
	fmt.Println("  risk <DID> <风险行为类型> - 模拟用户风险行为")
	fmt.Println("  help - 显示此帮助信息")
	fmt.Println("  exit/quit - 退出程序")
}

// 打印使用说明
func printUsage() {
	fmt.Println("蜜点后台客户端使用说明:")
	fmt.Println("  模拟风险行为: go run main.go risk <DID> <风险行为类型>")
	fmt.Println("  或者直接运行 go run main.go 进入交互式界面")
}
