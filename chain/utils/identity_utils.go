package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// GenerateDID 根据用户信息生成分布式身份标识符
func GenerateDID(name, idNumber, phoneNumber, vehicleID string) string {
	// 将用户信息组合起来
	combined := name + idNumber + phoneNumber + vehicleID
	
	// 计算SHA-256哈希
	hash := sha256.Sum256([]byte(combined))
	hashStr := hex.EncodeToString(hash[:])
	
	// 生成DID
	did := "did:example:" + hashStr[:16]
	
	return did
}

// GenerateInfoHash 生成用户信息哈希值
func GenerateInfoHash(name, idNumber, phoneNumber, vehicleID string) string {
	// 将用户信息组合起来
	combined := name + idNumber + phoneNumber + vehicleID
	
	// 计算SHA-256哈希
	hash := sha256.Sum256([]byte(combined))
	hashStr := hex.EncodeToString(hash[:])
	
	return hashStr
}

// ValidateDID 验证DID格式是否正确
func ValidateDID(did string) bool {
	// 检查DID前缀
	if !strings.HasPrefix(did, "did:example:") {
		return false
	}
	
	// 检查DID长度
	parts := strings.Split(did, ":")
	if len(parts) != 3 {
		return false
	}
	
	// 检查标识符部分
	identifier := parts[2]
	if len(identifier) < 16 {
		return false
	}
	
	return true
}

// CalculateInitialRiskScore 计算用户初始风险值
// 目前简单返回固定值，未来可以基于用户信息进行计算
func CalculateInitialRiskScore() float64 {
	return 0.00 // 初始风险值为0，修改为浮点数
}