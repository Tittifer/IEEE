package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Tittifer/IEEE/honeypoint_client/client"
)

func main() {
	// 创建蜜点客户端
	honeypointClient, err := client.NewHoneypointClient()
	if err != nil {
		log.Fatalf("创建蜜点客户端失败: %v", err)
	}
	defer honeypointClient.Close()

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 启动事件监听
	if err := honeypointClient.StartEventListener(); err != nil {
		log.Fatalf("启动事件监听失败: %v", err)
	}

	// 命令行交互
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("蜜点客户端已启动，输入 'help' 查看帮助信息")

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
		case "risk":
			if len(args) != 3 {
				fmt.Println("用法: risk <设备DID> <风险行为类型>")
				continue
			}
			did := args[1]
			behaviorType := args[2]
			
			err := honeypointClient.ProcessRiskBehavior(did, behaviorType)
			if err != nil {
				fmt.Printf("处理风险行为失败: %v\n", err)
			} else {
				fmt.Printf("已成功处理设备 %s 的风险行为 %s\n", did, behaviorType)
			}
		case "list":
			fmt.Println("可用的风险行为类型:")
			fmt.Println("侦察阶段:")
			fmt.Println("  visit_trap_ip - 访问陷阱IP")
			fmt.Println("  connect_bait_wifi - 连接诱饵WiFi")
			fmt.Println("  port_scan_honeypot - 对蜜点进行端口扫描")
			fmt.Println("初始接入阶段:")
			fmt.Println("  weak_password_login - 尝试弱口令登录")
			fmt.Println("  exploit_known_vulnerability - 利用已知漏洞攻击")
			fmt.Println("执行阶段:")
			fmt.Println("  execute_info_gathering - 执行信息收集命令")
			fmt.Println("  upload_script - 上传脚本文件")
			fmt.Println("  upload_known_backdoor - 上传已知后门程序")
			fmt.Println("  modify_config_file - 修改系统配置文件")
			fmt.Println("持久化阶段:")
			fmt.Println("  create_scheduled_task - 创建定时任务")
			fmt.Println("  modify_system_service - 修改系统服务")
			fmt.Println("防御规避阶段:")
			fmt.Println("  clear_stop_log_service - 清空或停止日志服务")
			fmt.Println("  use_rootkit - 使用Rootkit技术")
			fmt.Println("凭证访问阶段:")
			fmt.Println("  read_fake_credential - 读取伪造的凭证文件")
			fmt.Println("  attempt_memory_credential - 尝试内存抓取凭证")
			fmt.Println("横向移动阶段:")
			fmt.Println("  login_with_stolen_credential - 使用窃取的凭证登录")
			fmt.Println("数据收集阶段:")
			fmt.Println("  compress_sensitive_files - 打包压缩敏感文件")
			fmt.Println("渗出阶段:")
			fmt.Println("  transfer_data_outside - 向外网传输数据")
			fmt.Println("  trigger_bait_file_callback - 触发诱饵文件回调")
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
	fmt.Println("  help                       - 显示帮助信息")
	fmt.Println("  risk <设备DID> <风险行为类型>  - 模拟设备风险行为")
	fmt.Println("  list                       - 列出可用的风险行为类型")
	fmt.Println("  exit                       - 退出程序")
}