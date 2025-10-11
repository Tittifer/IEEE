package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Tittifer/IEEE/user_client/client"
)

// 全局变量
var (
	userClient *client.UserClient
	scanner    *bufio.Scanner
)

func main() {
	log.Println("启动用户客户端...")

	// 创建用户客户端
	var err error
	userClient, err = client.NewUserClient()
	if err != nil {
		log.Fatalf("创建用户客户端失败: %v", err)
	}
	defer userClient.Close()

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
	case "register":
		if len(os.Args) != 6 {
			fmt.Println("用法: go run main.go register <姓名> <身份证号> <手机号> <车辆ID>")
			return
		}
		name := os.Args[2]
		idNumber := os.Args[3]
		phoneNumber := os.Args[4]
		vehicleID := os.Args[5]

		result, err := userClient.RegisterUser(name, idNumber, phoneNumber, vehicleID)
		if err != nil {
			log.Fatalf("注册用户失败: %v", err)
		}
		fmt.Println(result)

	case "login":
		if len(os.Args) != 4 {
			fmt.Println("用法: go run main.go login <DID> <姓名>")
			return
		}
		did := os.Args[2]
		name := os.Args[3]

		result, err := userClient.UserLogin(did, name)
		if err != nil {
			log.Fatalf("用户登录失败: %v", err)
		}
		fmt.Println(result)

	case "logout":
		if len(os.Args) != 3 {
			fmt.Println("用法: go run main.go logout <DID>")
			return
		}
		did := os.Args[2]

		result, err := userClient.UserLogout(did)
		if err != nil {
			log.Fatalf("用户登出失败: %v", err)
		}
		fmt.Println(result)

	case "get-user":
		if len(os.Args) != 3 {
			fmt.Println("用法: go run main.go get-user <DID>")
			return
		}
		did := os.Args[2]

		result, err := userClient.GetUser(did)
		if err != nil {
			log.Fatalf("获取用户信息失败: %v", err)
		}
		fmt.Println(result)

	case "get-did":
		if len(os.Args) != 6 {
			fmt.Println("用法: go run main.go get-did <姓名> <身份证号> <手机号> <车辆ID>")
			return
		}
		name := os.Args[2]
		idNumber := os.Args[3]
		phoneNumber := os.Args[4]
		vehicleID := os.Args[5]

		result, err := userClient.GetDIDByInfo(name, idNumber, phoneNumber, vehicleID)
		if err != nil {
			log.Fatalf("获取DID失败: %v", err)
		}
		fmt.Println("DID:", result)
		
	case "reset-risk":
		if len(os.Args) != 3 {
			fmt.Println("用法: go run main.go reset-risk <DID>")
			return
		}
		did := os.Args[2]

		result, err := userClient.ResetUserRiskScore(did)
		if err != nil {
			log.Fatalf("重置用户风险评分失败: %v", err)
		}
		fmt.Println(result)

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
			registerUser()
		case "2":
			loginUser()
		case "3":
			logoutUser()
		case "4":
			getUserInfo()
		case "5":
			getDID()
		case "6":
			resetRiskScore()
		case "7":
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
	fmt.Println("\n========= 用户客户端 =========")
	fmt.Println("1. 用户注册")
	fmt.Println("2. 用户登录")
	fmt.Println("3. 用户登出")
	fmt.Println("4. 查询用户信息")
	fmt.Println("5. 获取DID")
	fmt.Println("6. 重置风险评分")
	fmt.Println("7. 退出")
	fmt.Println("============================")
}

// 用户注册
func registerUser() {
	fmt.Println("\n===== 用户注册 =====")
	name := getUserInput("请输入姓名")
	idNumber := getUserInput("请输入身份证号")
	phoneNumber := getUserInput("请输入手机号")
	vehicleID := getUserInput("请输入车辆ID")

	result, err := userClient.RegisterUser(name, idNumber, phoneNumber, vehicleID)
	if err != nil {
		fmt.Printf("注册用户失败: %v\n", err)
		return
	}
	fmt.Println(result)
}

// 用户登录
func loginUser() {
	fmt.Println("\n===== 用户登录 =====")
	did := getUserInput("请输入DID")
	name := getUserInput("请输入姓名")

	result, err := userClient.UserLogin(did, name)
	if err != nil {
		fmt.Printf("用户登录失败: %v\n", err)
		return
	}
	fmt.Println(result)
}

// 用户登出
func logoutUser() {
	fmt.Println("\n===== 用户登出 =====")
	did := getUserInput("请输入DID")

	result, err := userClient.UserLogout(did)
	if err != nil {
		fmt.Printf("用户登出失败: %v\n", err)
		return
	}
	fmt.Println(result)
}

// 获取用户信息
func getUserInfo() {
	fmt.Println("\n===== 查询用户信息 =====")
	did := getUserInput("请输入DID")

	result, err := userClient.GetUser(did)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}
	fmt.Println(result)
}

// 获取DID
func getDID() {
	fmt.Println("\n===== 获取DID =====")
	name := getUserInput("请输入姓名")
	idNumber := getUserInput("请输入身份证号")
	phoneNumber := getUserInput("请输入手机号")
	vehicleID := getUserInput("请输入车辆ID")

	result, err := userClient.GetDIDByInfo(name, idNumber, phoneNumber, vehicleID)
	if err != nil {
		fmt.Printf("获取DID失败: %v\n", err)
		return
	}
	fmt.Println("DID:", result)
}

// 获取用户输入
func getUserInput(prompt string) string {
	fmt.Printf("%s: ", prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// 重置风险评分
func resetRiskScore() {
	fmt.Println("\n===== 重置风险评分 =====")
	did := getUserInput("请输入DID")

	result, err := userClient.ResetUserRiskScore(did)
	if err != nil {
		fmt.Printf("重置风险评分失败: %v\n", err)
		return
	}
	fmt.Println(result)
}

// 打印使用说明
func printUsage() {
	fmt.Println("用户客户端使用说明:")
	fmt.Println("  注册用户: go run main.go register <姓名> <身份证号> <手机号> <车辆ID>")
	fmt.Println("  用户登录: go run main.go login <DID> <姓名>")
	fmt.Println("  用户登出: go run main.go logout <DID>")
	fmt.Println("  获取用户信息: go run main.go get-user <DID>")
	fmt.Println("  获取DID: go run main.go get-did <姓名> <身份证号> <手机号> <车辆ID>")
	fmt.Println("  重置风险评分: go run main.go reset-risk <DID>")
	fmt.Println("  或者直接运行 go run main.go 进入交互式界面")
}