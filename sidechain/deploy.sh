#!/bin/bash
#
# 侧链部署脚本 - 用于部署和测试IEEE侧链链码
#

# 获取脚本所在目录
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
NETWORK_DIR=${SCRIPT_DIR}/../sidechain_docker
CHAINCODE_DIR=${SCRIPT_DIR}

# 设置环境变量
export PATH=/home/yxt/fabric-samples/bin:$PATH
# 初始设置FABRIC_CFG_PATH，不同操作会临时修改此变量
export FABRIC_CFG_PATH=/home/yxt/fabric-samples/config

# 定义颜色
C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[0;33m'

# 定义链码相关变量
CHANNEL_NAME="sidechannel"
CHAINCODE_NAME="sidechaincc"
CHAINCODE_VERSION="1.0"
CHAINCODE_SEQUENCE="1"
CHAINCODE_INIT_REQUIRED="false"
CHAINCODE_LANGUAGE="golang"
CHAINCODE_LABEL="${CHAINCODE_NAME}_${CHAINCODE_VERSION}"

# 命令行参数标志
START_NETWORK=false
DEPLOY_CHAINCODE=false
CLEAN_NETWORK=false
TEST_CHAINCODE=false

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
    warnln "找不到peer二进制文件，尝试使用绝对路径"
    if [ -f "/home/yxt/fabric-samples/bin/peer" ]; then
      export PATH=/home/yxt/fabric-samples/bin:$PATH
      infoln "已添加fabric-samples/bin到PATH"
    else
      errorln "在/home/yxt/fabric-samples/bin中找不到peer二进制文件"
      errorln "请确保Hyperledger Fabric已正确安装"
      exit 1
    fi
  fi
  
  ## 再次检查peer是否可用
  if ! command -v peer &> /dev/null; then
    errorln "无法找到peer二进制文件，请检查安装路径"
    exit 1
  fi
  
  ## 检查docker
  if ! command -v docker &> /dev/null; then
    errorln "找不到docker命令，请安装docker"
    exit 1
  fi
  
  ## 检查docker-compose
  if ! command -v docker-compose &> /dev/null; then
    errorln "找不到docker-compose命令，请安装docker-compose"
    exit 1
  fi
  
  successln "所有必要的二进制文件已找到"
}

# 启动网络
function networkUp() {
  infoln "启动Hyperledger Fabric网络..."
  
  # 临时设置FABRIC_CFG_PATH为网络配置目录
  local PREV_CFG_PATH=$FABRIC_CFG_PATH
  export FABRIC_CFG_PATH=${NETWORK_DIR}/configtx
  
  cd ${NETWORK_DIR}
  ./network.sh up
  
  if [ $? -ne 0 ]; then
    errorln "启动网络失败"
    # 恢复原来的FABRIC_CFG_PATH
    export FABRIC_CFG_PATH=$PREV_CFG_PATH
    exit 1
  fi
  
  # 恢复原来的FABRIC_CFG_PATH
  export FABRIC_CFG_PATH=$PREV_CFG_PATH
  
  successln "网络启动成功"
}

# 关闭网络
function networkDown() {
  infoln "关闭Hyperledger Fabric网络..."
  
  # 临时设置FABRIC_CFG_PATH为网络配置目录
  local PREV_CFG_PATH=$FABRIC_CFG_PATH
  export FABRIC_CFG_PATH=${NETWORK_DIR}/configtx
  
  cd ${NETWORK_DIR}
  ./network.sh down
  
  if [ $? -ne 0 ]; then
    errorln "关闭网络失败"
    # 恢复原来的FABRIC_CFG_PATH
    export FABRIC_CFG_PATH=$PREV_CFG_PATH
    exit 1
  fi
  
  # 恢复原来的FABRIC_CFG_PATH
  export FABRIC_CFG_PATH=$PREV_CFG_PATH
  
  successln "网络已关闭"
}

# 打包链码
function packageChaincode() {
  infoln "打包链码..."
  
  # 确保当前目录是链码目录
  cd ${CHAINCODE_DIR}
  
  # 检查core.yaml文件是否存在
  if [ ! -f "$FABRIC_CFG_PATH/core.yaml" ]; then
    warnln "在 $FABRIC_CFG_PATH 中找不到core.yaml文件"
    
    # 尝试在fabric-samples目录中找到core.yaml
    if [ -f "/home/yxt/fabric-samples/config/core.yaml" ]; then
      export FABRIC_CFG_PATH=/home/yxt/fabric-samples/config
      infoln "已设置FABRIC_CFG_PATH为: $FABRIC_CFG_PATH"
    else
      errorln "找不到core.yaml配置文件，无法打包链码"
      exit 1
    fi
  fi
  
  # 创建chaincode.tar.gz包
  peer lifecycle chaincode package ${CHAINCODE_NAME}.tar.gz \
    --path ${CHAINCODE_DIR} \
    --lang ${CHAINCODE_LANGUAGE} \
    --label ${CHAINCODE_LABEL}
  
  if [ $? -ne 0 ]; then
    errorln "打包链码失败"
    exit 1
  fi
  
  successln "链码打包成功: ${CHAINCODE_NAME}.tar.gz"
}

# 安装链码
function installChaincode() {
  infoln "安装链码到节点..."
  
  # 复制链码包到CLI容器
  docker cp ${CHAINCODE_DIR}/${CHAINCODE_NAME}.tar.gz cli_sidechain:/opt/gopath/src/github.com/hyperledger/fabric/peer/${CHAINCODE_NAME}.tar.gz
  
  if [ $? -ne 0 ]; then
    errorln "复制链码包到CLI容器失败"
    exit 1
  fi
  
  # 安装链码到peer0.org1
  docker exec cli_sidechain peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz
  
  if [ $? -ne 0 ]; then
    errorln "在peer0.org1上安装链码失败"
    exit 1
  fi
  
  # 安装链码到peer1.org1
  docker exec -e CORE_PEER_ADDRESS=peer1.org1.sidechain.com:7061 \
          -e CORE_PEER_TLS_CERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/server.crt \
          -e CORE_PEER_TLS_KEY_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/server.key \
          -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer1.org1.sidechain.com/tls/ca.crt \
          cli_sidechain peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz
  
  if [ $? -ne 0 ]; then
    errorln "在peer1.org1上安装链码失败"
    exit 1
  fi
  
  successln "链码安装成功"
}

# 获取已安装的链码包ID
function getInstalledPackageID() {
  infoln "获取已安装的链码包ID..."
  
  PACKAGE_ID=$(docker exec cli_sidechain peer lifecycle chaincode queryinstalled | grep ${CHAINCODE_LABEL} | awk '{print $3}' | sed 's/,$//')
  
  if [ -z "$PACKAGE_ID" ]; then
    errorln "无法获取链码包ID"
    exit 1
  fi
  
  successln "链码包ID: ${PACKAGE_ID}"
}

# 批准链码定义
function approveChaincode() {
  infoln "批准链码定义..."
  
  # 获取链码包ID
  getInstalledPackageID
  
  # Org1批准链码
  docker exec cli_sidechain peer lifecycle chaincode approveformyorg \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --package-id ${PACKAGE_ID} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --init-required=${CHAINCODE_INIT_REQUIRED}
  
  if [ $? -ne 0 ]; then
    errorln "批准链码定义失败"
    exit 1
  fi
  
  successln "链码定义已批准"
}

# 提交链码定义
function commitChaincode() {
  infoln "提交链码定义到通道..."
  
  # 检查批准状态
  docker exec cli_sidechain peer lifecycle chaincode checkcommitreadiness \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --init-required=${CHAINCODE_INIT_REQUIRED} \
    --output json
  
  # 检查是否有序列号错误
  if docker exec cli_sidechain peer lifecycle chaincode checkcommitreadiness \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --init-required=${CHAINCODE_INIT_REQUIRED} \
    --output json 2>&1 | grep -q "must be sequence"; then
    
    warnln "检测到序列号问题"
    return 1
  fi
  
  # 提交链码定义
  docker exec cli_sidechain peer lifecycle chaincode commit \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    --channelID ${CHANNEL_NAME} \
    --name ${CHAINCODE_NAME} \
    --version ${CHAINCODE_VERSION} \
    --sequence ${CHAINCODE_SEQUENCE} \
    --init-required=${CHAINCODE_INIT_REQUIRED} \
    --peerAddresses peer0.org1.sidechain.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt
  
  if [ $? -ne 0 ]; then
    errorln "提交链码定义失败"
    return 1
  fi
  
  successln "链码定义已提交到通道"
  return 0
}

# 初始化链码
function initChaincode() {
  infoln "初始化链码..."
  
  if [ "$CHAINCODE_INIT_REQUIRED" = "true" ]; then
    docker exec cli_sidechain peer chaincode invoke \
      -o orderer.sidechain.com:7050 \
      --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
      -C ${CHANNEL_NAME} \
      -n ${CHAINCODE_NAME} \
      --isInit \
      -c '{"function":"InitLedger","Args":[]}' \
      --peerAddresses peer0.org1.sidechain.com:7051 \
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
      --waitForEvent
    
    if [ $? -ne 0 ]; then
      errorln "初始化链码失败"
      exit 1
    fi
  else
    infoln "链码不需要初始化"
  fi
  
  successln "链码已成功部署"
}

# 部署链码（包括打包、安装、批准和提交）
function deployChaincode() {
  infoln "开始部署链码..."
  
  packageChaincode
  installChaincode
  approveChaincode
  commitChaincode
  initChaincode
  
  # 生成CLI交互脚本
  generateCLIScript
  
  successln "链码部署完成！"
}

# 生成CLI交互脚本
function generateCLIScript() {
  infoln "生成CLI交互脚本..."
  
  # 复制模板脚本到sidechain_docker目录
  cp ${NETWORK_DIR}/sidechain_cli.sh ${NETWORK_DIR}/sidechain_cli.sh.bak
  
  # 更新脚本中的注释
  sed -i "s/# 由deploy.sh自动生成/# 由deploy.sh自动生成 - $(date)/" ${NETWORK_DIR}/sidechain_cli.sh
  
  # 确保脚本有执行权限
  chmod +x ${NETWORK_DIR}/sidechain_cli.sh
  
  successln "CLI交互脚本已更新"
}

# 测试链码
function testChaincode() {
  infoln "测试链码功能..."
  
  # 检查链码容器状态
  infoln "检查链码容器状态..."
  docker ps -a | grep ${CHAINCODE_NAME}
  
  # 查看链码日志
  CHAINCODE_CONTAINER=$(docker ps -a | grep ${CHAINCODE_NAME} | head -n 1 | awk '{print $1}')
  if [ -n "$CHAINCODE_CONTAINER" ]; then
    infoln "链码容器ID: $CHAINCODE_CONTAINER"
    infoln "链码容器日志:"
    docker logs $CHAINCODE_CONTAINER
  else
    warnln "找不到链码容器"
  fi
  
  # 尝试初始化链码
  infoln "尝试初始化链码..."
  docker exec cli_sidechain peer chaincode invoke \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    -c '{"function":"InitLedger","Args":[]}' \
    --peerAddresses peer0.org1.sidechain.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
    --waitForEvent
  
  sleep 3
  
  # 测试1：创建DID记录
  infoln "测试1：创建DID记录"
  docker exec cli_sidechain peer chaincode invoke \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    -c '{"function":"CreateDIDRecord","Args":["did:example:1234567890abcdef"]}' \
    --peerAddresses peer0.org1.sidechain.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
    --waitForEvent
  
  sleep 3
  
  # 测试2：更新用户状态（登录）
  infoln "测试2：更新用户状态（登录）"
  docker exec cli_sidechain peer chaincode invoke \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    -c '{"function":"UpdateUserStatus","Args":["did:example:1234567890abcdef", "online"]}' \
    --peerAddresses peer0.org1.sidechain.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
    --waitForEvent
  
  sleep 3
  
  # 测试3：报告风险行为
  infoln "测试3：报告风险行为"
  docker exec cli_sidechain peer chaincode invoke \
    -o orderer.sidechain.com:7050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    -c '{"function":"ReportRiskBehavior","Args":["did:example:1234567890abcdef", "A"]}' \
    --peerAddresses peer0.org1.sidechain.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
    --waitForEvent
  
  sleep 3
  
  # 测试4：评估用户风险评分
  infoln "测试4：评估用户风险评分"
  docker exec cli_sidechain peer chaincode query \
    -C ${CHANNEL_NAME} \
    -n ${CHAINCODE_NAME} \
    -c '{"function":"EvaluateRiskScore","Args":["did:example:1234567890abcdef"]}'
  
  successln "链码测试完成！"
}

# 显示帮助信息
function printHelp() {
  echo "用法:"
  echo "  deploy.sh [选项]"
  echo ""
  echo "选项:"
  echo "  -h    显示帮助信息"
  echo "  -n    启动网络"
  echo "  -d    部署链码"
  echo "  -c    清理网络"
  echo "  -t    测试链码"
  echo ""
  echo "示例:"
  echo "  ./deploy.sh -n -d -t    启动网络、部署链码并测试"
  echo "  ./deploy.sh -c          清理网络"
  echo "  ./deploy.sh -n -d       启动网络并部署链码"
}

# 解析命令行参数
if [ $# -eq 0 ]; then
  printHelp
  exit 0
fi

while getopts ":hncdt" opt; do
  case ${opt} in
    h )
      printHelp
      exit 0
      ;;
    n )
      START_NETWORK=true
      ;;
    c )
      CLEAN_NETWORK=true
      ;;
    d )
      DEPLOY_CHAINCODE=true
      ;;
    t )
      TEST_CHAINCODE=true
      ;;
    \? )
      echo "无效选项: -$OPTARG" 1>&2
      printHelp
      exit 1
      ;;
  esac
done

# 主函数
checkPrereqs

# 根据命令行参数执行操作
if [ "$CLEAN_NETWORK" = true ]; then
  networkDown
fi

if [ "$START_NETWORK" = true ]; then
  networkUp
fi

if [ "$DEPLOY_CHAINCODE" = true ]; then
  deployChaincode
fi

if [ "$TEST_CHAINCODE" = true ]; then
  testChaincode
fi

# 如果没有指定任何操作参数，显示帮助
if [ "$START_NETWORK" = false ] && [ "$DEPLOY_CHAINCODE" = false ] && [ "$CLEAN_NETWORK" = false ] && [ "$TEST_CHAINCODE" = false ]; then
  printHelp
  exit 0
fi
