# 快速开始指南

本指南将帮助你快速上手使用 GoAnsible。

## 安装

### 编译命令行工具

```bash
cd cmd/goansible/
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o goansible
```

### 验证安装

```bash
./goansible --help
```

## 初始化项目

执行以下命令在当前目录生成配置文件：

```bash
./goansible init
```

这将生成两个文件：

- `values.yaml` - 定义全局变量，如主机 SSH 信息
- `index.yaml` - 定义 playbook 及其 tasks

## 运行 Playbook

```bash
./goansible
```

## 下一个步骤

- 阅读 [使用教程](./tutorial.md) 了解更多功能
- 查看 [示例](../example) 目录获取实际使用案例
