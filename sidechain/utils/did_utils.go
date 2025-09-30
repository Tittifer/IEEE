package utils

import (
	"strings"
)

// ValidateDID 验证DID格式是否正确
func ValidateDID(did string) bool {
	// 检查DID前缀
	if !strings.HasPrefix(did, "did:") {
		return false
	}
	
	// 检查DID长度
	parts := strings.Split(did, ":")
	if len(parts) < 2 {
		return false
	}
	
	return true
}
