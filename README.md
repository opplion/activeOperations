# activeOperations 技术文档

## 1. 项目概述

activeOperations 是一个基于 Go 语言开发的智能运营平台，集成了大语言模型（LLM）、检索增强生成（RAG）和工作流管理等功能，旨在提供智能化的k8s运维解决方案。

## 2. 技术栈

### 2.1 核心技术

| 类别 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 编程语言 | Go | 1.24.9 | 后端开发 |
| Web 框架 | Gin | 1.11.0 | HTTP 服务 |
| 配置管理 | Viper | 1.21.0 | 配置加载与管理 |
| 向量数据库 | Milvus | 2.6.1 | 向量存储与检索 |
| LLM 框架 | Eino | v0.7.0 | 大语言模型集成 |
| 模型支持 | Qwen | v0.1.2 | 通义千问模型 |
| 容器化 | Docker | - | 应用容器化 |
| K8s 管理工具 | MCP | - | Kubernetes 集群管理 |

### 2.2 关键依赖

- `github.com/milvus-io/milvus/client/v2` - Milvus 客户端 SDK
- `github.com/gin-gonic/gin` - HTTP 框架
- `github.com/spf13/viper` - 配置管理
- `github.com/cloudwego/eino` - LLM 框架
- `github.com/weibaohui/kom` - Kom 工具包
- `github.com/mark3labs/mcp-go` - MCP 客户端

## 3. 项目结构

```
activeOperations/
├── .github/              # GitHub 配置
│   └── workflow/
│       └── ci.yml        # CI 工作流
├── bash/                 # Bash 脚本
│   └── reload.sh         # 重载脚本
├── cmd/                  # 可执行程序入口
│   ├── agent/            # Agent 服务
│   │   ├── dockerfile    # Docker 配置
│   │   └── main.go       # 主入口
│   └── mcp/              # MCP 服务
│       ├── dockerfile    # Docker 配置
│       └── main.go       # 主入口
├── config/               # 配置管理
│   ├── config.go         # 配置结构
│   ├── fs.go             # 文件系统配置
│   └── vars.go           # 配置变量
├── internal/             # 内部包
│   └── agent/            # Agent 核心逻辑
│       ├── controller/   # HTTP 控制器
│       ├── graph/        # 工作流图处理
│       ├── middleware/   # 中间件
│       ├── model/        # LLM 模型集成
│       ├── rag/          # RAG 相关功能
│       ├── router/       # 路由定义
│       └── tools/        # 工具函数
├── k8s/                  # Kubernetes 配置
│   └── XXX.yml       # K8s 用户配置
├── .gitignore            # Git 忽略文件
├── README.md             # 项目说明
├── docker-compose.yaml   # Docker Compose 配置
├── go.mod                # Go 模块定义
└── go.sum                # Go 依赖校验
```

## 4. 核心功能模块

### 4.1 配置管理

- 基于 Viper 实现的配置加载
- 支持 YAML 格式配置文件
- 提供配置变量的访问接口

### 4.2 RAG 系统

- **文档加载器**：支持多种格式文档加载
- **文本嵌入**：将文本转换为向量表示
- **向量存储**：使用 Milvus 存储向量数据
- **检索功能**：基于向量相似度的高效检索
- **动态重载**：支持文档动态更新

### 4.3 大语言模型集成

- 通过 Eino 框架集成 Qwen 模型
- 支持模型动态加载
- 提供统一的模型调用接口

### 4.4 工作流管理

- 基于图结构的工作流定义
- 支持工作流节点的动态创建和执行
- 提供工作流状态管理

### 4.5 HTTP 服务

- 基于 Gin 框架的 RESTful API
- 支持中间件扩展（如 Metrics）
- 提供优雅的服务启动和关闭机制

### 4.6 MCP

提供 Kubernetes 多集群管理代理，允许大语言模型通过 API Server 操作 Kubernetes 集群。它提供了以下核心功能：

- **多集群管理**：支持管理多个 Kubernetes 集群
- **安全访问**：基于配置的 Kubeconfig 进行安全认证
- **SSE 通信模式**：支持 Server-Sent Events 通信模式
- **集成到 LLM 工具链**：通过 MCP 客户端 SDK 与 LLM 框架集成

## 5. 系统架构

### 5.1 模块间依赖关系

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   配置管理      │────▶│   RAG 系统      │────▶│   向量数据库    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        ▲                        ▲                        ▲
        │                        │                        │
        ▼                        ▼                        ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   LLM 集成      │────▶│   ReactAgent    │────▶│   HTTP 服务     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        ▲                        ▲
        │                        │
        │                        │
        └────────────┬───────────┘
                     ▼
             ┌─────────────────┐
             │    MCP 服务      │────▶┌─────────────────┐
             └─────────────────┘     │   Kubernetes    │
                                     └─────────────────┘
```


### 5.2 启动流程

1. **配置加载**：加载 `./config.yaml` 配置文件
2. **模块初始化**（并发）：
   - 文档加载器初始化
   - 大语言模型加载
   - Milvus 向量数据库初始化
3. **工作流初始化**：获取并初始化工作流
4. **HTTP 服务启动**：启动 HTTP 服务器
5. **优雅关闭**：监听系统信号，支持优雅关闭

## 6. 启动方式

### Docker 启动

#### 前置条件

- Docker 和 Docker Compose 已安装

#### 启动步骤

1. **配置文件准备**：
   创建 `./config.yaml` 文件，配置必要参数：
   ```yaml
   HTTPPort: "8080"
   Milvus:
     Address: "localhost:19530"
     Username: "root"
     Password: "password"
     Database: "default"
   K8s:
     Path: "./k8s/k8sUser.yml"
   # 其他配置...
   ```


2. **启动服务**：
   ```bash
   docker-compose up -d
   ```

## 7. 配置说明

### 7.1 核心配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `AppName` | string | - | 应用名称 |
| `Mode` | string | "dev" | 运行模式（dev 或 prod） |
| `Version` | string | - | 应用版本 |
| `Host` | string | - | 主机地址 |
| `HttpPort` | string | "8080" | HTTP 服务端口 |
| `Model.Model` | string | - | 模型名称 |
| `Model.Apikey` | string | - | 模型 API 密钥 |
| `Milvus.Host` | string | - | Milvus 主机地址 |
| `Milvus.Port` | string | - | Milvus 服务端口 |
| `Milvus.CollectionName` | string | - | Milvus 集合名称 |
| `Embedding.Model` | string | - | 嵌入模型名称 |
| `Embedding.Apikey` | string | - | 嵌入模型 API 密钥 |
| `Embedding.Dimensions` | int | - | 嵌入向量维度 |

### 7.2 配置文件示例

```yaml
# 应用基本配置
AppName: "activeOperations"
Mode: "dev"
Version: "v1.0.0"
Host: "0.0.0.0"
HttpPort: "8080"

# 模型配置
Model:
  Model: "qwen"
  Apikey: "your-model-api-key"

# Milvus 配置
Milvus:
  Host: "localhost"
  Port: "19530"
  CollectionName: "active_operations_collection"

# 嵌入模型配置
Embedding:
  Model: "text-embedding-v1"
  Apikey: "your-embedding-api-key"
  Dimensions: 1536

# K8s 角色配置
K8s:
  Path: "./k8s/k8sUser.yml"
```
---

**文档版本**: v1.0.0  
**最后更新**: 2025-12-07  