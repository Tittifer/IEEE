package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Tittifer/IEEE/device_client/client"
)

func main() {
	// 创建设备客户端
	deviceClient, err := client.NewDeviceClient()
	if err != nil {
		log.Fatalf("创建设备客户端失败: %v", err)
	}
	defer deviceClient.Close()

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 命令行交互
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("设备客户端已启动，输入 'help' 查看帮助信息")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := scanner.Text()
		args := strings.Fields(command)

		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "help":
			printHelp()
		case "register":
			if len(args) != 5 {
				fmt.Println("用法: register <设备名称> <设备型号> <设备供应商> <设备ID>")
				continue
			}
			result, err := deviceClient.RegisterDevice(args[1], args[2], args[3], args[4])
			if err != nil {
				fmt.Printf("注册设备失败: %v\n", err)
			} else {
				fmt.Println(result)
			}
		case "info":
			if len(args) != 2 {
				fmt.Println("用法: info <DID>")
				continue
			}
			result, err := deviceClient.GetDevice(args[1])
			if err != nil {
				fmt.Printf("获取设备信息失败: %v\n", err)
			} else {
				fmt.Println(result)
			}
		case "did":
			if len(args) != 5 {
				fmt.Println("用法: did <设备名称> <设备型号> <设备供应商> <设备ID>")
				continue
			}
			result, err := deviceClient.GetDIDByInfo(args[1], args[2], args[3], args[4])
			if err != nil {
				fmt.Printf("获取DID失败: %v\n", err)
			} else {
				fmt.Println("设备DID:", result)
			}
		case "reset":
			if len(args) != 2 {
				fmt.Println("用法: reset <DID>")
				continue
			}
			result, err := deviceClient.ResetDeviceRiskScore(args[1])
			if err != nil {
				fmt.Printf("重置设备风险评分失败: %v\n", err)
			} else {
				fmt.Println(result)
			}
		case "risk":
			if len(args) != 2 {
				fmt.Println("用法: risk <DID>")
				continue
			}
			result, err := deviceClient.GetRiskResponse(args[1])
			if err != nil {
				fmt.Printf("获取风险响应策略失败: %v\n", err)
			} else {
				fmt.Println(result)
			}
		case "exit":
			fmt.Println("退出程序")
			return
		default:
			fmt.Println("未知命令，输入 'help' 查看帮助信息")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "读取输入时出错:", err)
	}
}

// 打印帮助信息
func printHelp() {
	fmt.Println("可用命令:")
	fmt.Println("  help                                       - 显示帮助信息")
	fmt.Println("  register <设备名称> <设备型号> <设备供应商> <设备ID> - 注册新设备")
	fmt.Println("  info <DID>                                 - 获取设备信息")
	fmt.Println("  did <设备名称> <设备型号> <设备供应商> <设备ID>   - 根据设备信息获取DID")
	fmt.Println("  reset <DID>                                - 重置设备风险评分")
	fmt.Println("  risk <DID>                                 - 获取设备风险响应策略")
	fmt.Println("  exit                                       - 退出程序")
}