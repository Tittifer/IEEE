# Sidechain

这个目录用于存放侧链相关的代码和配置。

## 目录结构

- `docker/`: 包含Docker配置文件和网络启动脚本
- 链码文件应直接放在这个目录下，而不是放在单独的chaincode目录中

## 使用方法

1. 将您的链码文件放在此目录下
2. 在CLI容器中，您可以通过以下路径访问链码文件：`/opt/gopath/src/github.com/sidechain/`
3. 安装链码时，使用以下命令：

```bash
peer chaincode install -n yourcc -v 1.0 -p github.com/sidechain/your_chaincode_dir
```

## 启动网络

```bash
cd docker
./network.sh up
```

## 关闭网络

```bash
cd docker
./network.sh down
```
