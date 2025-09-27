#!/bin/bash
# 部署脚本 - 一键启动Fabric网络并部署链码
# 作者: IEEE链码团队
# 版本: 1.1
# 
# 使用方法:
# ./deploy.sh [fabric-samples路径]
# 
# 示例:
# ./deploy.sh /home/yxt/fabric-samples

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 打印彩色信息
print_info() {
  echo -e "${GREEN}[INFO] $1${NC}"
}

print_warn() {
  echo -e "${YELLOW}[WARN] $1${NC}"
}

print_error() {
  echo -e "${RED}[ERROR] $1${NC}"
}

# 检查必要工具
check_prerequisites() {
  print_info "检查必要组件..."
  
  # 检查Docker
  if ! command -v docker &> /dev/null; then
    print_error "未安装Docker，请先安装Docker"
    exit 1
  fi
  
  # 检查Docker是否正在运行
  if ! docker info &> /dev/null; then
    print_error "Docker未启动，请先启动Docker服务"
    exit 1
  fi
  
  # 检查Docker Compose
  if ! command -v docker compose &> /dev/null; then
    if ! command -v docker-compose &> /dev/null; then
      print_error "未安装Docker Compose，请先安装Docker Compose"
      exit 1
    else
      print_warn "检测到旧版Docker Compose，建议升级到新版"
    fi
  fi
  
  # 检查Go
  if ! command -v go &> /dev/null; then
    print_error "未安装Go，请先安装Go 1.18或更高版本"
    exit 1
  fi
  
  # 检查fabric-samples目录
  if [ -n "$1" ]; then
    FABRIC_SAMPLES_PATH="$1"
  else
    FABRIC_SAMPLES_PATH=${FABRIC_SAMPLES_PATH:-"$HOME/fabric-samples"}
  fi
  
  if [ ! -d "$FABRIC_SAMPLES_PATH" ]; then
    print_error "未找到fabric-samples目录: $FABRIC_SAMPLES_PATH"
    print_info "可以通过以下命令安装fabric-samples:"
    print_info "curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.4.7 1.5.2"
    exit 1
  fi
  
  # 检查test-network目录
  if [ ! -d "$FABRIC_SAMPLES_PATH/test-network" ]; then
    print_error "在 $FABRIC_SAMPLES_PATH 中未找到test-network目录"
    exit 1
  fi
  
  # 检查network.sh脚本
  if [ ! -f "$FABRIC_SAMPLES_PATH/test-network/network.sh" ]; then
    print_error "在 $FABRIC_SAMPLES_PATH/test-network 中未找到network.sh脚本"
    exit 1
  fi
  
  # 确保network.sh有执行权限
  chmod +x "$FABRIC_SAMPLES_PATH/test-network/network.sh"
  
  print_info "所有必要组件已就绪"
  print_info "使用的fabric-samples路径: $FABRIC_SAMPLES_PATH"
}

# 清理Docker环境
clean_docker() {
  print_info "清理Docker环境..."
  
  # 停止所有运行中的容器
  if [ "$(docker ps -q)" ]; then
    docker stop $(docker ps -q)
  fi
  
  # 删除所有容器
  if [ "$(docker ps -a -q)" ]; then
    docker rm -f $(docker ps -a -q)
  fi
  
  # 清理卷和网络
  docker volume prune -f
  docker network prune -f
  
  print_info "Docker环境已清理"
}

# 启动Fabric测试网络
start_network() {
  print_info "启动Fabric测试网络..."
  
  # 保存当前目录
  CURRENT_DIR=$(pwd)
  
  # 进入test-network目录
  cd "$FABRIC_SAMPLES_PATH/test-network" || {
    print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
    exit 1
  }
  
  # 如果网络已经运行，先关闭
  print_info "关闭可能正在运行的网络..."
  ./network.sh down
  
  # 清理Docker环境，确保没有冲突
  clean_docker
  
  # 启动网络
  print_info "启动新网络..."
  ./network.sh up
  
  if [ $? -ne 0 ]; then
    print_error "启动网络失败"
    cd "$CURRENT_DIR"
    exit 1
  fi
  
  # 创建通道
  print_info "创建通道..."
  ./network.sh createChannel
  
  if [ $? -ne 0 ]; then
    print_error "创建通道失败"
    cd "$CURRENT_DIR"
    exit 1
  fi
  
  # 返回原目录
  cd "$CURRENT_DIR"
  
  print_info "Fabric测试网络已成功启动"
}

# 部署链码
deploy_chaincode() {
  print_info "开始部署链码..."
  
  # 保存当前目录
  CURRENT_DIR=$(pwd)
  
  # 链码名称和路径
  CHAINCODE_NAME="powercc"
  CHAINCODE_PATH=$(pwd)
  CHAINCODE_VERSION="1.0"
  
  print_info "链码名称: $CHAINCODE_NAME"
  print_info "链码路径: $CHAINCODE_PATH"
  
  # 进入test-network目录
  cd "$FABRIC_SAMPLES_PATH/test-network" || {
    print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
    exit 1
  }
  
  # 部署链码
  print_info "正在部署链码，这可能需要几分钟时间..."
  ./network.sh deployCC -ccn "$CHAINCODE_NAME" -ccp "$CHAINCODE_PATH" -ccl go -ccv "$CHAINCODE_VERSION" -ccs 1 -verbose
  
  DEPLOY_RESULT=$?
  if [ $DEPLOY_RESULT -ne 0 ]; then
    print_error "部署链码失败，退出码: $DEPLOY_RESULT"
    cd "$CURRENT_DIR"
    exit 1
  fi
  
  # 返回原目录
  cd "$CURRENT_DIR"
  
  print_info "链码已成功部署"
}

# 设置环境变量以便与链码交互
setup_environment() {
  print_info "设置环境变量..."
  
  # 保存当前目录
  CURRENT_DIR=$(pwd)
  
  # 进入test-network目录
  cd "$FABRIC_SAMPLES_PATH/test-network" || {
    print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
    exit 1
  }
  
  # 设置环境变量
  export PATH=${PWD}/../bin:$PATH
  export FABRIC_CFG_PATH=$PWD/../config/
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=localhost:7051
  
  # 保存环境变量到文件，方便后续使用
  cat > env.sh << EOF
#!/bin/bash
# Fabric环境变量设置脚本
# 由deploy.sh自动生成

export PATH=${PWD}/../bin:\$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

echo "Fabric环境变量已设置"
EOF
  
  # 添加执行权限
  chmod +x env.sh
  
  # 创建一个切换到Org2的脚本
  cat > switch_to_org2.sh << EOF
#!/bin/bash
# 切换到Org2的环境变量设置脚本
# 由deploy.sh自动生成

export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

echo "已切换到Org2"
EOF
  
  # 添加执行权限
  chmod +x switch_to_org2.sh
  
  # 返回原目录
  cd "$CURRENT_DIR"
  
  print_info "环境变量已设置，脚本已保存到:"
  print_info "- $FABRIC_SAMPLES_PATH/test-network/env.sh"
  print_info "- $FABRIC_SAMPLES_PATH/test-network/switch_to_org2.sh"
}

# 初始化链码
init_chaincode() {
  print_info "初始化链码..."
  
  # 保存当前目录
  CURRENT_DIR=$(pwd)
  
  # 进入test-network目录
  cd "$FABRIC_SAMPLES_PATH/test-network" || {
    print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
    exit 1
  }
  
  # 获取通道名称和链码名称
  CHANNEL_NAME="mychannel"
  CHAINCODE_NAME="powercc"
  
  # 确保环境变量已设置
  source ./env.sh > /dev/null 2>&1
  
  print_info "调用InitLedger函数..."
  
  # 调用InitLedger函数
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C $CHANNEL_NAME -n $CHAINCODE_NAME --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"InitLedger","Args":[]}' --waitForEvent
  
  INIT_RESULT=$?
  if [ $INIT_RESULT -ne 0 ]; then
    print_warn "初始化链码返回了非零状态码: $INIT_RESULT"
    print_warn "如果链码不需要初始化或InitLedger函数不存在，这可能是正常的"
  else
    print_info "链码已成功初始化"
  fi
  
  # 等待几秒，确保链码初始化完成
  sleep 3
  
  # 返回原目录
  cd "$CURRENT_DIR"
}

# 测试链码
test_chaincode() {
  print_info "测试链码基本功能..."
  
  # 保存当前目录
  CURRENT_DIR=$(pwd)
  
  # 进入test-network目录
  cd "$FABRIC_SAMPLES_PATH/test-network" || {
    print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
    exit 1
  }
  
  # 获取通道名称和链码名称
  CHANNEL_NAME="mychannel"
  CHAINCODE_NAME="powercc"
  
  # 确保环境变量已设置
  source ./env.sh > /dev/null 2>&1
  
  # 注册测试用户
  print_info "注册测试用户..."
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C $CHANNEL_NAME -n $CHAINCODE_NAME --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RegisterUser","Args":["did:example:123", "张三", "110101199001011234", "测试公钥"]}' --waitForEvent
  
  REGISTER_RESULT=$?
  if [ $REGISTER_RESULT -ne 0 ]; then
    print_error "注册用户失败，退出码: $REGISTER_RESULT"
    cd "$CURRENT_DIR"
    return 1
  fi
  
  # 等待交易确认
  sleep 3
  
  # 查询用户信息
  print_info "查询用户信息..."
  USER_INFO=$(peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"function":"GetUser","Args":["did:example:123"]}')
  
  QUERY_RESULT=$?
  if [ $QUERY_RESULT -ne 0 ]; then
    print_error "查询用户信息失败，退出码: $QUERY_RESULT"
    cd "$CURRENT_DIR"
    return 1
  fi
  
  print_info "用户信息查询结果:"
  echo "$USER_INFO" | jq . || echo "$USER_INFO"
  
  # 记录攻击行为
  print_info "记录攻击行为..."
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C $CHANNEL_NAME -n $CHAINCODE_NAME --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RecordAttack","Args":["did:example:123", "honeypot001", "SQL注入", "尝试通过SQL注入获取数据库访问权限", "8"]}' --waitForEvent
  
  # 等待交易确认
  sleep 3
  
  # 再次查询用户信息，验证风险评分是否已更新
  print_info "再次查询用户信息，验证风险评分是否已更新..."
  UPDATED_USER_INFO=$(peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"function":"GetUser","Args":["did:example:123"]}')
  
  print_info "更新后的用户信息:"
  echo "$UPDATED_USER_INFO" | jq . || echo "$UPDATED_USER_INFO"
  
  # 返回原目录
  cd "$CURRENT_DIR"
  
  print_info "链码功能测试完成"
}

# 创建帮助文档
create_help_doc() {
  print_info "创建帮助文档..."
  
  cat > chaincode_commands.md << EOF
# 电网身份认证链码命令指南

本文档提供了与电网身份认证链码交互的常用命令。

## 环境设置

在使用以下命令前，请先设置环境变量：

\`\`\`bash
source env.sh
\`\`\`

## 用户管理命令

### 注册新用户

\`\`\`bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile \${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RegisterUser","Args":["did:example:123", "张三", "110101199001011234", "公钥内容..."]}'
\`\`\`

### 获取用户信息

\`\`\`bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUser","Args":["did:example:123"]}'
\`\`\`

### 验证用户身份

\`\`\`bash
peer chaincode query -C mychannel -n powercc -c '{"function":"VerifyUser","Args":["did:example:123"]}'
\`\`\`

### 更改用户状态

\`\`\`bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile \${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"ChangeUserStatus","Args":["did:example:123", "suspended"]}'
\`\`\`

### 获取所有用户

\`\`\`bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetAllUsers","Args":[]}'
\`\`\`

## 风险管理命令

### 记录攻击行为

\`\`\`bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile \${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RecordAttack","Args":["did:example:123", "honeypot001", "SQL注入", "尝试通过SQL注入获取数据库访问权限", "8"]}'
\`\`\`

### 更新风险评分

\`\`\`bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile \${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles \${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"UpdateRiskScore","Args":["did:example:123", "50"]}'
\`\`\`

### 获取高风险用户

\`\`\`bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetHighRiskUsers","Args":[]}'
\`\`\`

### 获取特定风险评分范围内的用户

\`\`\`bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUsersByRiskScore","Args":["40", "70"]}'
\`\`\`

## 关闭网络

当测试完成后，可以使用以下命令关闭网络：

\`\`\`bash
cd \$FABRIC_SAMPLES_PATH/test-network
./network.sh down
\`\`\`
EOF
  
  print_info "帮助文档已创建: chaincode_commands.md"
}

# 显示使用帮助
show_help() {
  echo "使用方法: $0 [选项] [fabric-samples路径]"
  echo ""
  echo "选项:"
  echo "  -h, --help      显示此帮助信息"
  echo "  -c, --clean     清理Docker环境并停止网络"
  echo "  -n, --network   只启动网络，不部署链码"
  echo "  -d, --deploy    只部署链码，不启动网络"
  echo "  -t, --test      只测试链码，不启动网络和部署"
  echo ""
  echo "示例:"
  echo "  $0                         # 执行完整部署流程"
  echo "  $0 /path/to/fabric-samples  # 使用指定的fabric-samples路径"
  echo "  $0 --clean                 # 清理环境"
  echo "  $0 --network               # 只启动网络"
  echo "  $0 --deploy                # 只部署链码"
}

# 主函数
main() {
  print_info "=== 电网身份认证链码部署脚本 ==="
  
  # 解析命令行参数
  CLEAN_ONLY=false
  NETWORK_ONLY=false
  DEPLOY_ONLY=false
  TEST_ONLY=false
  
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_help
        exit 0
        ;;
      -c|--clean)
        CLEAN_ONLY=true
        shift
        ;;
      -n|--network)
        NETWORK_ONLY=true
        shift
        ;;
      -d|--deploy)
        DEPLOY_ONLY=true
        shift
        ;;
      -t|--test)
        TEST_ONLY=true
        shift
        ;;
      *)
        FABRIC_SAMPLES_PATH="$1"
        shift
        ;;
    esac
  done
  
  # 检查必要组件
  check_prerequisites "$FABRIC_SAMPLES_PATH"
  
  # 如果只是清理环境
  if [ "$CLEAN_ONLY" = true ]; then
    print_info "只执行清理操作..."
    cd "$FABRIC_SAMPLES_PATH/test-network" || {
      print_error "无法进入目录: $FABRIC_SAMPLES_PATH/test-network"
      exit 1
    }
    ./network.sh down
    clean_docker
    print_info "清理完成"
    exit 0
  fi
  
  # 如果只是测试链码
  if [ "$TEST_ONLY" = true ]; then
    print_info "只执行测试操作..."
    setup_environment
    test_chaincode
    exit 0
  fi
  
  # 如果只启动网络
  if [ "$NETWORK_ONLY" = true ]; then
    print_info "只启动网络..."
    start_network
    setup_environment
    print_info "网络已启动，环境变量已设置"
    exit 0
  fi
  
  # 如果只部署链码
  if [ "$DEPLOY_ONLY" = true ]; then
    print_info "只部署链码..."
    deploy_chaincode
    setup_environment
    init_chaincode
    print_info "链码已部署并初始化"
    exit 0
  fi
  
  # 完整流程
  print_info "执行完整部署流程..."
  
  # 启动网络
  start_network
  
  # 部署链码
  deploy_chaincode
  
  # 设置环境变量
  setup_environment
  
  # 初始化链码
  init_chaincode
  
  # 测试链码
  test_chaincode
  
  # 创建帮助文档
  create_help_doc
  
  print_info "=== 部署完成! ==="
  print_info "您可以使用 chaincode_commands.md 中的命令与链码交互"
  print_info "使用前请先运行: source $FABRIC_SAMPLES_PATH/test-network/env.sh"
}

# 执行主函数
main "$@"
