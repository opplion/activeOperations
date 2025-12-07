# activeOperations 技术文档

## 1. 项目概述

activeOperations 是一个基于 Go 语言开发的智能运营平台，集成了大语言模型（LLM）、检索增强生成（RAG）、工作流管理和 Kubernetes 多集群管理（MCP）等功能，旨在提供智能化的运营解决方案。

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
| 通信协议 | gRPC | 1.73.0 | 服务间通信 |
| 容器化 | Docker | - | 应用容器化 |
| 编排工具 | Kubernetes | - | 容器编排 |
| K8s 管理工具 | MCP | - | Kubernetes 多集群管理 |
| 智能代理 | ReactAgent | - | 基于大模型的智能代理模式 |

### 2.2 关键依赖

- `github.com/milvus-io/milvus/client/v2` - Milvus 客户端 SDK
- `github.com/gin-gonic/gin` - HTTP 框架
- `github.com/spf13/viper` - 配置管理
- `github.com/cloudwego/eino` - LLM 框架
- `github.com/mark3labs/mcp-go` - MCP 客户端
- `github.com/weibaohui/kom` - Kom 工具包
- `google.golang.org/grpc` - gRPC 支持

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
│       ├── proto/        # gRPC 协议定义
│       ├── rag/          # RAG 相关功能
│       ├── router/       # 路由定义
│       └── tools/        # 工具函数
├── k8s/                  # Kubernetes 配置
│   └── k8sUser.yml       # K8s 用户配置
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

### 4.6 MCP (Multi-Cluster Proxy) 服务

#### 4.6.1 功能概述

MCP 服务是一个 Kubernetes 多集群管理代理，允许大语言模型通过 API Server 操作 Kubernetes 集群。它提供了以下核心功能：

- **多集群管理**：支持管理多个 Kubernetes 集群
- **安全访问**：基于配置的 Kubeconfig 进行安全认证
- **SSE 通信模式**：支持 Server-Sent Events 通信模式
- **集成到 LLM 工具链**：通过 MCP 客户端 SDK 与 LLM 框架集成

#### 4.6.2 架构设计

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   大语言模型    │────▶│   MCP 客户端    │────▶│   MCP 服务器    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                      ▲
                                                      │
                                                      ▼
                                              ┌─────────────────┐
                                              │   Kubernetes    │
                                              │   集群1, 集群2   │
                                              └─────────────────┘
```

#### 4.6.3 关键实现

- **MCP 服务器**：
  - 位于 `cmd/mcp/main.go`
  - 使用 `github.com/weibaohui/kom/mcp` 包
  - 配置文件：`k8s/k8sUser.yml`
  - 默认端口：4321
  - 通信模式：SSE (Server-Sent Events)

- **MCP 客户端**：
  - 位于 `internal/agent/tools/tool.go`
  - 使用 `github.com/mark3labs/mcp-go/client` 包
  - 与 Eino 框架集成，提供 K8s 操作工具

#### 4.6.4 工作流程

1. MCP 服务器启动，加载 Kubeconfig 配置
2. LLM 框架初始化时，创建 MCP 客户端连接
3. 大语言模型生成 K8s 操作请求
4. MCP 客户端将请求发送到 MCP 服务器
5. MCP 服务器执行 K8s 操作并返回结果
6. 结果返回给大语言模型进行后续处理

### 4.7 ReactAgent 模式

#### 4.7.1 功能概述

ReactAgent 是一种基于大语言模型的智能代理模式，它将传统的链式调用转换为基于图的工作流，允许更灵活、更强大的交互。ReactAgent 模式的核心特点包括：

- **基于图的工作流**：使用有向图定义工作流节点和依赖关系
- **模块化设计**：每个节点负责特定功能，易于扩展
- **上下文传递**：节点间可以传递上下文信息
- **智能决策**：结合 LLM 的能力进行智能决策

#### 4.7.2 架构设计

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  MilvusRetriever│────▶│     merge       │────▶│  ReactAgent     │
│  (向量检索)      │     │  (上下文合并)    │     │  (智能代理)     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

#### 4.7.3 关键实现

- **工作流定义**：
  - 位于 `internal/agent/graph/graph.go`
  - 使用 Eino 的 `compose` 包定义工作流图
  - 包含三个主要节点：MilvusRetriever、merge 和 ReactAgent

- **节点处理器**：
  - `MilvusRetrieverHandler`：负责从 Milvus 检索相关文档
  - `lambdaHandler`：负责合并检索结果和用户查询
  - `ReactAgentHandler`：负责调用大语言模型生成最终响应

- **工作流编译**：
  - 使用 `wf.Compile(ctx)` 编译工作流图
  - 生成可执行的工作流实例

#### 4.7.4 工作流程

1. 用户发送查询请求
2. 工作流引擎启动，首先执行 MilvusRetriever 节点
3. MilvusRetriever 从向量数据库中检索相关文档
4. merge 节点将检索结果与用户查询合并，生成上下文
5. ReactAgent 节点接收上下文，调用大语言模型生成响应
6. 最终响应返回给用户

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
3. **工作流初始化**：编译 ReactAgent 工作流图
4. **HTTP 服务启动**：启动 HTTP 服务器
5. **优雅关闭**：监听系统信号，支持优雅关闭

## 6. 亮点特性

### 6.1 模块化设计

- 清晰的模块划分，便于扩展和维护
- 支持模块的独立升级和替换
- 松耦合的设计理念

### 6.2 高性能

- 基于 Go 语言的高并发设计
- 模块初始化并发执行
- 高效的向量检索

### 6.3 可扩展性

- 支持多种文档格式扩展
- 可集成不同的大语言模型
- 支持自定义工作流节点
- 可管理多个 Kubernetes 集群

### 6.4 容器化支持

- 提供 Dockerfile 和 docker-compose.yaml
- 支持 Kubernetes 部署
- 便于在不同环境中部署和运行

### 6.5 智能化

- 集成最新的大语言模型技术
- 支持检索增强生成，提高模型输出质量
- 基于图结构的智能工作流
- ReactAgent 模式实现智能代理

### 6.6 Kubernetes 集成

- 通过 MCP 服务实现多集群管理
- 大语言模型可以直接操作 Kubernetes 资源
- 支持安全的 Kubeconfig 认证

## 7. 启动方式

### 7.1 本地启动

#### 7.1.1 前置条件

- Go 1.24.9 或更高版本
- Milvus 向量数据库服务
- Kubernetes 集群（可选，用于 MCP 功能）
- 配置文件 `./config.yaml`

#### 7.1.2 启动步骤

1. **启动 MCP 服务**（可选，用于 K8s 操作）：
   ```bash
   # 进入项目根目录
   cd activeOperations
   
   # 启动 MCP 服务
   go run cmd/mcp/main.go
   ```

2. **启动 Agent 服务**：
   ```bash
   # 进入项目根目录
   cd activeOperations
   
   # 启动 Agent 服务
   go run cmd/agent/main.go
   ```

3. **验证启动**：
   查看日志输出，确认服务启动成功：
   ```
   🚀 应用启动中...
   📥 正在加载配置文件 ./config.yaml
   ✅ 配置加载成功
   # ... 模块初始化日志 ...
   🎉 应用启动成功！
   ```

### 7.2 Docker 启动

#### 7.2.1 前置条件

- Docker 和 Docker Compose 已安装

#### 7.2.2 启动步骤

1. **启动服务**：
   ```bash
   docker-compose up -d
   ```

2. **查看日志**：
   ```bash
   docker-compose logs -f
   ```

### 7.3 Kubernetes 部署

1. **准备 Kubernetes 配置**：
   查看 `k8s/` 目录下的配置文件

2. **应用配置**：
   ```bash
   kubectl apply -f k8s/
   ```

3. **验证部署**：
   ```bash
   kubectl get pods
   ```

## 8. 配置说明

### 8.1 核心配置项

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

### 8.2 配置文件示例

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
```

## 9. API 接口

### 9.1 健康检查

- **URL**: `/health`
- **Method**: GET
- **Response**: 
  ```json
  {
    "status": "ok",
    "timestamp": "2025-01-01T12:00:00Z"
  }
  ```

### 9.2 工作流相关接口

- **获取工作流**: `GET /api/v1/workflow`
- **创建工作流**: `POST /api/v1/workflow`
- **执行工作流**: `POST /api/v1/workflow/{id}/execute`

### 9.3 RAG 相关接口

- **查询文档**: `POST /api/v1/rag/query`
- **加载文档**: `POST /api/v1/rag/load`
- **重载文档**: `POST /api/v1/rag/reload`

### 9.4 MCP 相关接口

- **K8s 资源列表**: `GET /api/v1/k8s/{resource}`
- **创建 K8s 资源**: `POST /api/v1/k8s/{resource}`
- **更新 K8s 资源**: `PUT /api/v1/k8s/{resource}/{name}`
- **删除 K8s 资源**: `DELETE /api/v1/k8s/{resource}/{name}`

## 10. 监控与日志

### 10.1 日志

- 日志输出到控制台
- 支持不同级别日志（INFO, ERROR, FATAL 等）
- 包含模块初始化、服务启动、错误信息等

### 10.2 监控

- 集成 Prometheus 指标（通过 middleware/Metrics.go）
- 支持监控服务请求数、响应时间等
- 可通过 `/metrics` 端点访问指标数据

## 11. 开发指南

### 11.1 代码规范

- 遵循 Go 语言标准规范
- 使用 `go fmt` 格式化代码
- 使用 `golint` 检查代码质量

### 11.2 测试

- 编写单元测试和集成测试
- 运行测试：
  ```bash
  go test ./...
  ```

### 11.3 CI/CD

- 使用 GitHub Actions 进行 CI
- 自动运行测试和代码检查
- 支持自动构建和部署

## 12. 常见问题

### 12.1 Milvus 连接失败

**问题**：启动时出现 Milvus 连接失败

**解决方法**：
- 检查 Milvus 服务是否正常运行
- 验证配置文件中的 Milvus 地址、用户名和密码
- 检查网络连接是否正常

### 12.2 模型加载失败

**问题**：模型加载失败

**解决方法**：
- 检查模型服务端点是否正确
- 验证 API Key 是否有效
- 检查模型服务是否正常运行

### 12.3 MCP 连接失败

**问题**：MCP 客户端连接失败

**解决方法**：
- 检查 MCP 服务器是否正常运行
- 验证 MCP 服务器地址和端口
- 检查 Kubeconfig 配置是否正确

### 12.4 工作流执行失败

**问题**：ReactAgent 工作流执行失败

**解决方法**：
- 检查工作流定义是否正确
- 验证各节点处理器是否正常工作
- 查看日志获取详细错误信息

## 13. 未来规划

- 支持更多类型的大语言模型
- 增强工作流的可视化配置
- 提供更丰富的监控指标
- 支持多租户部署
- 增强系统的安全性
- 提供更完善的 API 文档
- 支持更多的 Kubernetes 资源类型
- 增强 ReactAgent 的智能决策能力
- 支持工作流的动态调整和优化

## 14. 联系方式

- 项目地址：https://github.com/your-org/activeOperations
- 问题反馈：https://github.com/your-org/activeOperations/issues
- 贡献指南：CONTRIBUTING.md

---

**文档版本**: v1.0.0  
**最后更新**: 2025-12-07  
**作者**: activeOperations 开发团队