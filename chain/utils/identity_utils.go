package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// GenerateDID 根据设备信息生成DID
func GenerateDID(name, model, vendor, deviceID string) string {
	// 合并设备信息
	info := fmt.Sprintf("%s:%s:%s:%s", name, model, vendor, deviceID)
	
	// 计算SHA256哈希
	hash := sha256.Sum256([]byte(info))
	hashStr := hex.EncodeToString(hash[:])
	
	// 生成DID
	did := fmt.Sprintf("did:ieee:device:%s", hashStr[:16])
	
	return did
}

// ValidateDID 验证DID格式
func ValidateDID(did string) bool {
	// DID格式: did:ieee:device:<16位十六进制字符>
	pattern := `^did:ieee:device:[0-9a-f]{16}$`
	match, _ := regexp.MatchString(pattern, did)
	return match
}

// CalculateInitialRiskScore 计算初始风险评分
func CalculateInitialRiskScore() float64 {
	return 0.0 // 设备初始风险评分为0
}

// ExtractDIDFromSubject 从证书主题中提取DID
func ExtractDIDFromSubject(subject string) (string, error) {
	// 假设证书主题格式为 "CN=<did>,OU=..."
	if strings.HasPrefix(subject, "CN=") {
		parts := strings.Split(subject, ",")
		if len(parts) > 0 {
			cn := parts[0]
			did := strings.TrimPrefix(cn, "CN=")
			if ValidateDID(did) {
				return did, nil
			}
		}
	}
	
	return "", fmt.Errorf("无法从证书主题中提取有效的DID: %s", subject)
}