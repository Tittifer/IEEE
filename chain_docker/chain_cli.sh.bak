#!/bin/bash
# 由deploy.sh自动生成 - Fri Oct  3 14:17:59 CST 2025 - Fri Oct  3 13:29:52 CST 2025

# 定义颜色
C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[0;33m'

# 定义链码相关变量
CHANNEL_NAME="mainchannel"
CHAINCODE_NAME="chaincc"

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

# 显示帮助信息
function printHelp() {
  echo "用法:"
  echo "  chain_cli.sh <命令> <函数名> [参数...]"
  echo "命令:"
  echo "  query    - 查询链码"
  echo "  invoke   - 调用链码"
  echo "函数名:"
  echo "  链码中定义的函数名，例如 GetUser, RegisterUser 等"
  echo "参数:"
  echo "  函数所需的参数，多个参数用空格分隔"
  echo ""
  echo "示例:"
  echo "  ./chain_cli.sh query GetDIDByInfo 张三 110101199001011234 13800138000 京A12345"
  echo "  ./chain_cli.sh invoke RegisterUser 李四 110101199001012345 13900139000 京B12345"
}

# 查询链码
function queryChaincode() {
  if [ $# -lt 1 ]; then
    errorln "需要至少提供函数名参数"
    printHelp
    exit 1
  fi
  
  FUNC_NAME=$1
  shift
  ARGS=$(echo $@ | sed 's/ /", "/g')
  if [ -z "$ARGS" ]; then
    ARGS="[]"
  else
    ARGS="[\"$ARGS\"]"
  fi
  
  infoln "查询链码: ${FUNC_NAME} 参数: ${ARGS}"
  
  docker exec cli_chain peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c "{\"function\":\"$FUNC_NAME\",\"Args\":$ARGS}"
}

# 调用链码
function invokeChaincode() {
  if [ $# -lt 1 ]; then
    errorln "需要至少提供函数名参数"
    printHelp
    exit 1
  fi
  
  FUNC_NAME=$1
  shift
  ARGS=$(echo $@ | sed 's/ /", "/g')
  if [ -z "$ARGS" ]; then
    ARGS="[]"
  else
    ARGS="[\"$ARGS\"]"
  fi
  
  infoln "调用链码: ${FUNC_NAME} 参数: ${ARGS}"
  
  docker exec cli_chain peer chaincode invoke \
    -o orderer.chain.com:8050 \
    --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/chain.com/orderers/orderer.chain.com/msp/tlscacerts/tlsca.chain.com-cert.pem \
    -C $CHANNEL_NAME \
    -n $CHAINCODE_NAME \
    -c "{\"function\":\"$FUNC_NAME\",\"Args\":$ARGS}" \
    --peerAddresses peer0.org1.chain.com:8051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt \
    --waitForEvent
}

# 主函数
if [ $# -lt 2 ]; then
  printHelp
  exit 0
fi

COMMAND=$1
shift

case $COMMAND in
  "query")
    queryChaincode $@
    ;;
  "invoke")
    invokeChaincode $@
    ;;
  *)
    errorln "未知命令: $COMMAND"
    printHelp
    exit 1
    ;;
esac

