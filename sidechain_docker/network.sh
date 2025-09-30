#!/bin/bash
#
# 侧链网络启动脚本 - Sidechain
#

# 获取脚本所在目录
ROOTDIR=$(cd "$(dirname "$0")" && pwd)
export PATH=/home/yxt/fabric-samples/bin:$PATH
export FABRIC_CFG_PATH=${ROOTDIR}/configtx

# 定义颜色
C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[0;33m'

# 打印信息函数
function infoln() {
  echo -e "${C_BLUE}${1}${C_RESET}"
}

function errorln() {
  echo -e "${C_RED}${1}${C_RESET}"
}

function successln() {
  echo -e "${C_GREEN}${1}${C_RESET}"
}

function warnln() {
  echo -e "${C_YELLOW}${1}${C_RESET}"
}

# 检查必要的二进制文件
function checkPrereqs() {
  infoln "检查必要的二进制文件..."
  
  ## 检查peer二进制文件
  if ! command -v peer &> /dev/null; then
    errorln "找不到peer二进制文件，请确保fabric-samples/bin目录已添加到PATH中"
    exit 1
  fi
  
  ## 检查configtxgen二进制文件
  if ! command -v configtxgen &> /dev/null; then
    errorln "找不到configtxgen二进制文件，请确保fabric-samples/bin目录已添加到PATH中"
    exit 1
  fi
  
  ## 检查cryptogen二进制文件
  if ! command -v cryptogen &> /dev/null; then
    errorln "找不到cryptogen二进制文件，请确保fabric-samples/bin目录已添加到PATH中"
    exit 1
  fi
  
  successln "所有必要的二进制文件已找到"
}

# 清理容器和生成的文件
function networkDown() {
  infoln "关闭网络..."
  docker-compose -f ${ROOTDIR}/docker-compose.yaml down --volumes --remove-orphans
  
  # 删除通道和创世区块文件
  rm -rf ${ROOTDIR}/channel-artifacts/*
  
  # 删除加密材料
  rm -rf ${ROOTDIR}/crypto-config
  
  successln "网络已关闭，所有文件已清理"
}

# 生成加密材料
function generateCryptoMaterial() {
  infoln "生成加密材料..."
  
  if [ -d "${ROOTDIR}/crypto-config" ]; then
    rm -rf ${ROOTDIR}/crypto-config
  fi
  
  cryptogen generate --config=${ROOTDIR}/crypto-config.yaml --output=${ROOTDIR}/crypto-config
  
  if [ $? -ne 0 ]; then
    errorln "生成加密材料失败"
    exit 1
  fi
  
  successln "加密材料生成成功"
}

# 生成创世区块和通道配置
function generateGenesisBlock() {
  infoln "生成创世区块..."
  
  if [ ! -d "${ROOTDIR}/channel-artifacts" ]; then
    mkdir -p ${ROOTDIR}/channel-artifacts
  fi
  
  # 生成系统通道创世区块
  configtxgen -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock ${ROOTDIR}/channel-artifacts/genesis.block
  
  if [ $? -ne 0 ]; then
    errorln "生成创世区块失败"
    exit 1
  fi
  
  # 生成应用通道交易
  configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ${ROOTDIR}/channel-artifacts/channel.tx -channelID sidechannel
  
  if [ $? -ne 0 ]; then
    errorln "生成通道交易失败"
    exit 1
  fi
  
  # 生成锚节点更新交易
  configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ${ROOTDIR}/channel-artifacts/Org1MSPanchors.tx -channelID sidechannel -asOrg Org1MSP
  
  if [ $? -ne 0 ]; then
    errorln "生成锚节点更新交易失败"
    exit 1
  fi
  
  successln "创世区块和通道配置生成成功"
}

# 启动网络
function networkUp() {
  checkPrereqs
  
  # 如果网络已经在运行，先关闭
  if [ $(docker ps -q --filter name=peer0.org1.sidechain.com | wc -l) -gt 0 ]; then
    warnln "检测到网络已在运行，先关闭..."
    networkDown
  fi
  
  # 生成加密材料
  generateCryptoMaterial
  
  # 生成创世区块和通道配置
  generateGenesisBlock
  
  # 启动网络容器
  infoln "启动网络容器..."
  docker-compose -f ${ROOTDIR}/docker-compose.yaml up -d
  
  if [ $? -ne 0 ]; then
    errorln "启动网络容器失败"
    exit 1
  fi
  
  # 等待容器启动
  sleep 5
  
  successln "网络启动成功！"
  
  # 自动创建通道并加入所有节点
  infoln "开始创建通道并加入所有节点..."
  createChannel
}

# 创建通道
function createChannel() {
  infoln "创建通道..."
  
  # 进入CLI容器创建通道
  docker exec cli_sidechain peer channel create -o orderer.sidechain.com:7050 -c sidechannel -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem
  
  if [ $? -ne 0 ]; then
    errorln "创建通道失败"
    exit 1
  fi
  
  # peer0.org1加入通道
  docker exec cli_sidechain peer channel join -b sidechannel.block
  
  if [ $? -ne 0 ]; then
    errorln "peer0.org1加入通道失败"
    exit 1
  fi
  
  # peer1.org1加入通道
  docker exec -e CORE_PEER_ADDRESS=peer1.org1.sidechain.com:7061 \
              -e CORE_PEER_TLS_CERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/server.crt \
              -e CORE_PEER_TLS_KEY_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/server.key \
              -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/ca.crt \
              cli_sidechain peer channel join -b sidechannel.block
  
  if [ $? -ne 0 ]; then
    errorln "peer1.org1加入通道失败"
    exit 1
  fi
  
  # 更新锚节点
  docker exec cli_sidechain peer channel update -o orderer.sidechain.com:7050 -c sidechannel -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org1MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem
  
  if [ $? -ne 0 ]; then
    errorln "更新锚节点失败"
    exit 1
  fi
  
  successln "通道创建成功，所有节点已加入通道"
}

# 显示帮助信息
function printHelp() {
  echo "用法:"
  echo "  network.sh <命令> [选项]"
  echo "命令:"
  echo "  up - 启动网络并自动创建通道及加入所有节点"
  echo "  down - 关闭网络并清理文件"
  echo "  createChannel - 手动创建通道并加入所有节点"
  echo "  restart - 重启网络并自动创建通道及加入所有节点"
}

# 主函数
if [ "$1" = "up" ]; then
  networkUp
elif [ "$1" = "down" ]; then
  networkDown
elif [ "$1" = "restart" ]; then
  networkDown
  networkUp
  # 不需要显式调用createChannel，因为networkUp已经包含了这个步骤
elif [ "$1" = "createChannel" ]; then
  createChannel
else
  printHelp
  exit 1
fi
