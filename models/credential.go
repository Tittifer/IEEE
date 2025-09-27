package models

import (
	"time"
)

// VerifiableCredential 可验证凭证结构体
type VerifiableCredential struct {
	Context           []string          `json:"@context"`           // 凭证上下文
	ID                string            `json:"id"`                 // 凭证唯一标识符
	Type              []string          `json:"type"`               // 凭证类型
	Issuer            Issuer            `json:"issuer"`             // 颁发者信息
	IssuanceDate      time.Time         `json:"issuanceDate"`       // 颁发日期
	ExpirationDate    *time.Time        `json:"expirationDate,omitempty"` // 过期日期（可选）
	CredentialSubject CredentialSubject `json:"credentialSubject"`  // 凭证主体
	Proof             Proof             `json:"proof"`              // 凭证证明
	CredentialStatus  CredentialStatus  `json:"credentialStatus"`   // 凭证状态
}

// Issuer 颁发者信息
type Issuer struct {
	ID   string `json:"id"`   // 颁发者DID
	Name string `json:"name"` // 颁发者名称
}

// CredentialSubject 凭证主体
type CredentialSubject struct {
	ID            string `json:"id"`            // 主体DID
	Name          string `json:"name"`          // 用户姓名
	IDNumber      string `json:"idNumber"`      // 身份证号
	PhoneNumber   string `json:"phoneNumber"`   // 电话号码
	VehiclePlate  string `json:"vehiclePlate"`  // 车牌信息
	VehicleModel  string `json:"vehicleModel"`  // 车辆型号
	VehicleColor  string `json:"vehicleColor"`  // 车辆颜色
	VehicleVIN    string `json:"vehicleVIN"`    // 车辆识别号
}

// Proof 凭证证明
type Proof struct {
	Type               string    `json:"type"`               // 证明类型
	Created            time.Time `json:"created"`            // 创建时间
	VerificationMethod string    `json:"verificationMethod"` // 验证方法
	ProofPurpose       string    `json:"proofPurpose"`       // 证明目的
	ProofValue         string    `json:"proofValue"`         // 证明值（签名）
}

// CredentialStatus 凭证状态
type CredentialStatus struct {
	ID   string `json:"id"`   // 状态标识符
	Type string `json:"type"` // 状态类型
}

// 凭证状态常量
const (
	CredentialStatusActive    = "Active"
	CredentialStatusSuspended = "Suspended"
	CredentialStatusRevoked   = "Revoked"
)
