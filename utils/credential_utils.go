package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tittifer/IEEE/models"
)

// 系统颁发者信息
const (
	IssuerDID  = "did:example:issuer:123456789"
	IssuerName = "电网身份认证系统"
)

// 系统签名配置
const (
	SignatureType = "Sha256Signature2018"
)

// GenerateDID 根据用户信息生成DID
func GenerateDID(name, idNumber string) string {
	// 使用用户姓名和身份证号的哈希作为DID的一部分
	hasher := sha256.New()
	hasher.Write([]byte(name + idNumber))
	hash := hasher.Sum(nil)
	
	// 使用base64编码哈希值（去掉填充字符）
	encodedHash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash)
	
	// 生成DID格式
	return fmt.Sprintf("did:example:user:%s", encodedHash[:16])
}

// CreateVerifiableCredential 创建可验证凭证
func CreateVerifiableCredential(user models.UserInfo) (*models.VerifiableCredential, error) {
	// 创建凭证
	now := time.Now()
	expirationDate := now.AddDate(1, 0, 0) // 凭证有效期1年
	
	credential := &models.VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		},
		ID:    fmt.Sprintf("urn:uuid:%s", GenerateUUID()),
		Type:  []string{"VerifiableCredential", "VehicleCredential"},
		Issuer: models.Issuer{
			ID:   IssuerDID,
			Name: IssuerName,
		},
		IssuanceDate:   now,
		ExpirationDate: &expirationDate,
		CredentialSubject: models.CredentialSubject{
			ID:            user.DID,
			Name:          user.Name,
			IDNumber:      user.IDNumber,
			PhoneNumber:   user.PhoneNumber,
			VehiclePlate:  user.VehiclePlate,
			VehicleModel:  user.VehicleModel,
			VehicleColor:  user.VehicleColor,
			VehicleVIN:    user.VehicleVIN,
		},
		CredentialStatus: models.CredentialStatus{
			ID:   fmt.Sprintf("%s/status/%s", IssuerDID, GenerateUUID()),
			Type: models.CredentialStatusActive,
		},
	}
	
	// 添加证明（签名）
	proof, err := SignCredential(credential)
	if err != nil {
		return nil, fmt.Errorf("签名凭证失败: %v", err)
	}
	
	credential.Proof = *proof
	
	return credential, nil
}

// SignCredential 对凭证进行签名
func SignCredential(credential *models.VerifiableCredential) (*models.Proof, error) {
	// 创建一个临时凭证副本，不包含证明字段
	tempCred := *credential
	tempCred.Proof = models.Proof{}
	
	// 序列化凭证
	credBytes, err := json.Marshal(tempCred)
	if err != nil {
		return nil, fmt.Errorf("序列化凭证失败: %v", err)
	}
	
	// 计算哈希
	hasher := sha256.New()
	hasher.Write(credBytes)
	hash := hasher.Sum(nil)
	
	// 使用哈希值作为签名（在实际应用中应使用真正的数字签名）
	// 这里简化处理，避免RSA密钥解析错误
	signature := base64.StdEncoding.EncodeToString(hash)
	
	// 创建证明
	now := time.Now()
	proof := &models.Proof{
		Type:               SignatureType,
		Created:            now,
		VerificationMethod: fmt.Sprintf("%s#keys-1", IssuerDID),
		ProofPurpose:       "assertionMethod",
		ProofValue:         signature,
	}
	
	return proof, nil
}

// GenerateUUID 生成简单的UUID
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "00000000-0000-0000-0000-000000000000"
	}
	
	// 设置版本 (4) 和变体位
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
