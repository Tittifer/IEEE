
export PATH=/home/yxt/fabric-samples/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/configtx
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/home/yxt/hyperledger-fabric/chaincode/IEEE/chain_docker/crypto-config/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/home/yxt/hyperledger-fabric/chaincode/IEEE/chain_docker/crypto-config/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/msp
export CORE_PEER_ADDRESS=localhost:8050

echo "Fabric环境变量已设置"