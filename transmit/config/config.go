package config

// 配置信息
type Config struct {
	// Mainchain配置
	MainchainConfig NetworkConfig `json:"mainchain_config"`
	// Sidechain配置
	SidechainConfig NetworkConfig `json:"sidechain_config"`
	// 同步间隔（秒）
	SyncInterval int `json:"sync_interval"`
}

// 网络配置
type NetworkConfig struct {
	// 通道名称
	ChannelName string `json:"channel_name"`
	// 链码名称
	ChaincodeName string `json:"chaincode_name"`
	// MSP ID
	MspID string `json:"msp_id"`
	// 证书路径
	CertPath string `json:"cert_path"`
	// 私钥路径
	KeyPath string `json:"key_path"`
	// TLS证书路径
	TlsCertPath string `json:"tls_cert_path"`
	// Peer端点
	PeerEndpoint string `json:"peer_endpoint"`
	// Gateway端点
	GatewayPeer string `json:"gateway_peer"`
}

// 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		MainchainConfig: NetworkConfig{
			ChannelName:   "mainchannel",
			ChaincodeName: "mainchaincc",
			MspID:         "Org1MSP",
			CertPath:      "/home/yxt/hyperledger-fabric/chaincode/IEEE/mainchain_docker/crypto-config/peerOrganizations/org1.mainchain.com/users/User1@org1.mainchain.com/msp/signcerts/User1@org1.mainchain.com-cert.pem",
			KeyPath:       "/home/yxt/hyperledger-fabric/chaincode/IEEE/mainchain_docker/crypto-config/peerOrganizations/org1.mainchain.com/users/User1@org1.mainchain.com/msp/keystore/priv_sk",
			TlsCertPath:   "/home/yxt/hyperledger-fabric/chaincode/IEEE/mainchain_docker/crypto-config/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt",
			PeerEndpoint:  "localhost:8051",
			GatewayPeer:   "peer0.org1.mainchain.com:8051",
		},
		SidechainConfig: NetworkConfig{
			ChannelName:   "sidechannel",
			ChaincodeName: "sidechaincc",
			MspID:         "Org1MSP",
			CertPath:      "/home/yxt/hyperledger-fabric/chaincode/IEEE/sidechain_docker/crypto-config/peerOrganizations/org1.sidechain.com/users/User1@org1.sidechain.com/msp/signcerts/User1@org1.sidechain.com-cert.pem",
			KeyPath:       "/home/yxt/hyperledger-fabric/chaincode/IEEE/sidechain_docker/crypto-config/peerOrganizations/org1.sidechain.com/users/User1@org1.sidechain.com/msp/keystore/priv_sk",
			TlsCertPath:   "/home/yxt/hyperledger-fabric/chaincode/IEEE/sidechain_docker/crypto-config/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt",
			PeerEndpoint:  "localhost:7051",
			GatewayPeer:   "peer0.org1.sidechain.com:7051",
		},
		SyncInterval: 30, // 默认30秒同步一次
	}
}