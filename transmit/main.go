package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tittifer/IEEE/transmit/config"
	"github.com/Tittifer/IEEE/transmit/connection"
	"github.com/Tittifer/IEEE/transmit/handlers"
	"github.com/Tittifer/IEEE/transmit/models"
)

func main() {
	log.Println("启动链间数据传输服务...")

	// 加载配置
	cfg := config.GetDefaultConfig()

	// 连接主链
	mainchainConn, err := connection.NewMainchainConnection(&cfg.MainchainConfig)
	if err != nil {
		log.Fatalf("连接主链失败: %v", err)
	}
	defer mainchainConn.Close()

	// 连接侧链
	sidechainConn, err := connection.NewSidechainConnection(&cfg.SidechainConfig)
	if err != nil {
		log.Fatalf("连接侧链失败: %v", err)
	}
	defer sidechainConn.Close()

	// 创建处理器
	mainchainHandler := handlers.NewMainchainHandler(mainchainConn.GetContract())
	sidechainHandler := handlers.NewSidechainHandler(sidechainConn.GetContract())
	
	// 创建信号通道，用于优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 创建同步通道，用于同步主链和侧链数据
	syncTicker := time.NewTicker(time.Duration(cfg.SyncInterval) * time.Second)
	defer syncTicker.Stop()
	
	log.Printf("开始监听链上事件并同步数据，同步间隔: %d秒", cfg.SyncInterval)
	
	// 主循环
	for {
		select {
		case <-syncTicker.C:
			// 定期同步数据
			syncData(mainchainHandler, sidechainHandler)
		case <-quit:
			log.Println("正在关闭链间数据传输服务...")
			return
		}
	}
}

// 同步数据
func syncData(mainchainHandler *handlers.MainchainHandler, sidechainHandler *handlers.SidechainHandler) {
	log.Println("开始同步数据...")
	
	// 1. 从主链获取所有用户
	users, err := mainchainHandler.GetAllUsers()
	if err != nil {
		log.Printf("获取主链用户失败: %v", err)
		return
	}
	
	log.Printf("从主链获取到 %d 个用户", len(users))
	
	// 2. 检查每个用户在侧链是否有对应的DID记录
	for _, user := range users {
		// 检查用户DID在侧链是否存在
		didRecord, err := sidechainHandler.GetDIDRecord(user.DID)
		if err != nil {
			// 如果DID记录不存在，则在侧链创建
			log.Printf("用户 %s 在侧链没有DID记录，正在创建...", user.DID)
			err = sidechainHandler.CreateDIDRecord(user.DID)
			if err != nil {
				log.Printf("在侧链创建用户 %s 的DID记录失败: %v", user.DID, err)
				continue
			}
			log.Printf("在侧链成功创建用户 %s 的DID记录", user.DID)
		} else {
			// 同步风险评分到主链
			if user.RiskScore != didRecord.RiskScore {
				log.Printf("用户 %s 的风险评分需要同步: 主链 %d -> 侧链 %d", user.DID, user.RiskScore, didRecord.RiskScore)
				err = mainchainHandler.UpdateRiskScore(user.DID, didRecord.RiskScore)
				if err != nil {
					log.Printf("更新用户 %s 风险评分失败: %v", user.DID, err)
				} else {
					log.Printf("已将用户 %s 的风险评分从 %d 更新为 %d", user.DID, user.RiskScore, didRecord.RiskScore)
				}
			}
		}
		
		// 3. 同步用户登录/登出状态
		// 根据主链用户状态同步侧链会话状态
		if user.Status == "online" {
			// 如果用户在主链是在线状态，检查侧链是否也是在线状态
			session, err := sidechainHandler.GetUserSession(user.DID)
			if err != nil || session.Status != "online" {
				// 如果侧链没有会话记录或不是在线状态，则更新为在线状态
				log.Printf("用户 %s 在主链是在线状态，但在侧链不是，正在同步...", user.DID)
				err = sidechainHandler.UpdateUserStatus(user.DID, "online")
				if err != nil {
					log.Printf("更新用户 %s 在侧链的状态失败: %v", user.DID, err)
				} else {
					log.Printf("已将用户 %s 在侧链的状态更新为在线状态", user.DID)
				}
			}
		} else if user.Status == "offline" {
			// 如果用户在主链是离线状态，检查侧链是否也是离线状态
			session, err := sidechainHandler.GetUserSession(user.DID)
			if err == nil && session.Status != "offline" {
				// 如果侧链有会话记录且不是离线状态，则更新为离线状态
				log.Printf("用户 %s 在主链是离线状态，但在侧链不是，正在同步...", user.DID)
				err = sidechainHandler.UpdateUserStatus(user.DID, "offline")
				if err != nil {
					log.Printf("更新用户 %s 在侧链的状态失败: %v", user.DID, err)
				} else {
					log.Printf("已将用户 %s 在侧链的状态更新为离线状态", user.DID)
				}
			}
		}
		
		// 4. 同步侧链会话状态到主链
		session, err := sidechainHandler.GetUserSession(user.DID)
		if err == nil {
			// 如果用户在侧链处于在线状态，但在主链状态不是在线，则更新主链状态
			if session.Status == "online" && user.Status != "online" {
				log.Printf("用户 %s 在侧链是在线状态，但在主链是 %s 状态，正在同步...", user.DID, user.Status)
				err = mainchainHandler.UpdateUserStatus(user.DID, "online")
				if err != nil {
					log.Printf("更新用户 %s 在主链的状态失败: %v", user.DID, err)
				} else {
					log.Printf("已将用户 %s 在主链的状态更新为在线状态", user.DID)
				}
			}
			
			// 如果用户在侧链处于离线状态，但在主链状态是在线，则更新主链状态
			if session.Status == "offline" && user.Status == "online" {
				log.Printf("用户 %s 在侧链是离线状态，但在主链是在线状态，正在同步...", user.DID)
				err = mainchainHandler.UpdateUserStatus(user.DID, "offline")
				if err != nil {
					log.Printf("更新用户 %s 在主链的状态失败: %v", user.DID, err)
				} else {
					log.Printf("已将用户 %s 在主链的状态更新为离线状态", user.DID)
				}
			}
		}
	}
	
	// 5. 从侧链获取所有DID记录
	didRecords, err := sidechainHandler.GetAllDIDRecords()
	if err != nil {
		log.Printf("获取侧链DID记录失败: %v", err)
		return
	}
	
	log.Printf("从侧链获取到 %d 个DID记录", len(didRecords))
	
	// 6. 检查每个DID记录在主链是否有对应的用户
	for _, record := range didRecords {
		// 尝试从主链获取用户信息
		_, err := mainchainHandler.GetUser(record.DID)
		if err != nil {
			log.Printf("DID %s 在主链没有对应的用户: %v", record.DID, err)
			// 这里可以根据需要决定是否删除侧链上的孤立DID记录
		}
	}
	
	log.Println("数据同步完成")
}