# 🤖 Super Agent — 智能 OnCall 运维助手

<p align="center">
  <strong>基于多 Agent 协作架构的智能运维平台，集成 RAG 知识检索、Plan-Execute 自动化分析与 ReAct 推理于一体</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Framework-Eino-FF6B6B?style=flat-square" alt="Eino">
  <img src="https://img.shields.io/badge/LLM-DeepSeek_V3-4A90D9?style=flat-square" alt="DeepSeek">
  <img src="https://img.shields.io/badge/VectorDB-Milvus-00B4AB?style=flat-square&logo=milvus" alt="Milvus">
  <img src="https://img.shields.io/badge/Deploy-Docker_Compose-2496ED?style=flat-square&logo=docker" alt="Docker">
</p>

---

## 📖 目录

- [项目简介](#-项目简介)
- [在线体验](#-在线体验)
- [核心功能](#-核心功能)
- [技术架构](#-技术架构)
- [Agent 实现方案](#-agent-实现方案)
- [工具链集成](#-工具链集成)
- [目录结构](#-目录结构)
- [快速开始](#-快速开始)
- [配置说明](#-配置说明)
- [后续优化](#-后续优化)

---

## 🎯 项目简介

**Super Agent** 是一个面向运维场景的智能 AI 助手，旨在解决传统 OnCall 值班中的痛点：信息分散、排查效率低、知识传承困难。

系统采用 **多 Agent 协作** 的架构设计，融合了当前 AI Agent 领域的主流技术方案：

| 能力 | 方案 | 说明 |
|:---:|:---:|:---|
| 🧠 智能问答 | **RAG + ReAct Agent** | 检索增强生成 + 工具增强推理 |
| 🔧 运维分析 | **Plan-Execute Agent** | LLM 驱动的自动化分析流水线 |
| 📚 知识管理 | **ETL Pipeline** | 文档 → 分割 → 向量化 → 存储 |
| 💬 会话管理 | **Summary Buffer Memory** | 滑动窗口 + LLM 摘要压缩 |

---

## 🌐 在线体验

> **体验地址：[http://49.232.223.185:3001](http://49.232.223.185:3001)**

### 业务问答模式
输入业务相关问题，系统会自动检索知识库并结合 LLM 推理给出专业回答：

```
📝 示例：「分发表达式是什么？它有哪些类型？」
```

### AI Ops 运维分析模式
描述运维场景需求，系统自动解析意图、制定分析计划并执行：

```
📝 示例：「帮我分析一下最近 1 小时的告警情况」
📝 示例：「查看 bypass 服务的 CLS 日志，排查请求超时问题」
```

---

## 🔥 核心功能

### 1. 💬 智能业务问答（Chat Agent）

- **RAG 检索增强**：基于 Milvus 向量数据库，对用户问题进行语义检索，召回最相关的知识文档作为上下文
- **ReAct 推理模式**：Agent 能够在回答过程中自主调用工具（搜索、文档检索、时间查询等），实现多步推理
- **SSE 流式输出**：全链路流式响应，输出即时呈现，用户体验流畅
- **多轮对话记忆**：基于 bigcache 的会话管理，支持滑动窗口 + LLM 摘要压缩

### 2. 🔧 AI Ops 智能运维（AIOps Agent）

- **自然语言意图解析**：用户用自然语言描述运维需求，LLM 自动提取场景类型、分析目标、时间范围等结构化参数
- **Plan-Execute 自动化**：Planner 制定分析计划 → Executor 逐步执行 → Replanner 动态调整
- **三大分析场景**：
  - 🚨 **告警分析**：查询 Prometheus 活跃告警 → 检索内部文档 → 分析日志 → 输出诊断报告
  - 📋 **日志分析**：通过 MCP 协议调用腾讯云 CLS → 模式识别 → 根因定位
  - 📊 **性能分析**：Prometheus 指标查询 → 趋势分析 → 优化建议
- **全程流式反馈**：意图确认 → 步骤进度卡片 → 最终分析报告，用户全程可观测

### 3. 📚 知识库管理

- **文件上传入库**：支持通过 API 上传 Markdown 文档，自动触发异步索引
- **ETL Pipeline**：`文件加载 → Markdown 按标题分割 → 千问 Embedding → Milvus 存储`
- **增量更新**：上传同名文档自动清除旧数据，实现增量更新

---

## 🏗 技术架构

### 整体架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                        Frontend (Nginx :3001)                    │
│                  原生 HTML/CSS/JS · 暗色主题 · SSE 流式          │
├──────────────────────────────────────────────────────────────────┤
│                              │ /api/*                            │
│                     Nginx Reverse Proxy                          │
│                              ▼                                   │
├──────────────────────────────────────────────────────────────────┤
│                     Backend (GoFrame :6872)                       │
│  ┌──────────────────┬───────────────────┬──────────────────┐     │
│  │ POST /chat_stream │ POST /ai_ops      │ POST /upload     │     │
│  │ (SSE 流式问答)    │ (SSE 运维分析)     │ (文件上传+索引)   │     │
│  └────────┬─────────┴─────────┬─────────┴─────────┬────────┘     │
│           ▼                   ▼                   ▼              │
│  ┌────────────────┐ ┌──────────────────┐ ┌────────────────────┐  │
│  │  Chat Agent    │ │  AIOps Runner    │ │ Knowledge Pipeline │  │
│  │  (ReAct Mode)  │ │ (Plan-Execute)   │ │   (ETL Graph)      │  │
│  └────────┬───────┘ └──────┬───────────┘ └──────┬─────────────┘  │
│           ▼                ▼                     ▼               │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Eino AI Agent Framework (字节跳动开源)           │ │
│  │   Graph 编排 · ReAct Agent · ChatTemplate · Tool Calling    │ │
│  └───────────────────┬─────────────────────────────────────────┘ │
├──────────────────────┼──────────────────────────────────────────┤
│              External Services                                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ ┌────────────────┐  │
│  │DeepSeek  │ │ Milvus   │ │ 千问 Embedding│ │  腾讯云 CLS    │  │
│  │V3 (LLM)  │ │(VectorDB)│ │  (DashScope)  │ │  (MCP 协议)    │  │
│  └──────────┘ └──────────┘ └──────────────┘ └────────────────┘  │
│  ┌──────────────────┐                                            │
│  │ Prometheus       │                                            │
│  │ (告警/指标查询)    │                                            │
│  └──────────────────┘                                            │
└──────────────────────────────────────────────────────────────────┘
```

### 技术栈

| 层次 | 技术 | 说明 |
|:---|:---|:---|
| **Web 框架** | GoFrame v2 | 高性能 Go Web 框架，内置路由、配置管理 |
| **AI 框架** | [Cloudwego Eino](https://github.com/cloudwego/eino) | 字节跳动开源的 Go 语言 AI Agent 框架，支持图编排 |
| **LLM** | DeepSeek V3 | 通过火山引擎 API 接入（兼容 OpenAI 协议） |
| **向量数据库** | Milvus | 开源分布式向量数据库，用于知识库语义检索 |
| **Embedding** | 千问 text-embedding-v4 | 阿里云 DashScope，用于文档和查询的向量化 |
| **前端** | 原生 HTML/CSS/JS | 暗色主题，ChatGPT 风格交互，SSE 流式 |
| **部署** | Docker Compose | 前端 Nginx + 后端 Go 二进制，一键部署 |
| **会话存储** | bigcache | 进程内 LRU 缓存，带 TTL 淘汰 |

---

## 🧠 Agent 实现方案

### Chat Agent —— RAG + ReAct 模式

#### 图编排流程

基于 Eino 框架的 Graph 编排能力，Chat Agent 的数据流如下：

```
                    ┌───────────────┐
                    │    START      │
                    │ (UserMessage) │
                    └───────┬───────┘
                            │
                  ┌─────────┴──────────┐
                  ▼                    ▼
        ┌─────────────────┐  ┌─────────────────┐
        │  InputToRAG     │  │  InputToChat    │
        │ (提取检索 Query) │  │ (提取对话输入)  │
        └────────┬────────┘  └────────┬────────┘
                 ▼                    │
        ┌─────────────────┐           │
        │ MilvusRetriever │           │
        │ (向量检索 Top-K) │           │
        └────────┬────────┘           │
                 │                    │
                 └──────────┬─────────┘
                            ▼
                  ┌─────────────────┐
                  │  ChatTemplate   │
                  │ (系统 Prompt +   │
                  │  检索文档 + 历史  │
                  │  对话 + 用户问题) │
                  └────────┬────────┘
                           ▼
                  ┌─────────────────┐
                  │   ReAct Agent   │
                  │ (工具增强推理)    │
                  │ Tools: 搜索/文档  │
                  │ /时间/Prometheus │
                  └────────┬────────┘
                           ▼
                  ┌─────────────────┐
                  │      END        │
                  │  (流式输出)      │
                  └─────────────────┘
```

**关键设计**：
- 并行分支：用户输入同时分发到 RAG 检索和对话输入两条路径，减少延迟
- ChatTemplate 融合多源信息：系统提示 + 历史对话 + 检索文档 + 用户问题
- ReAct Agent 保留工具调用能力，在 RAG 结果不足时可主动调用工具补充信息

### AIOps Agent —— Plan-Execute 模式

#### 整体流程

```
  用户自然语言输入
        │
        ▼
┌─────────────────┐
│   ParseIntent   │   ◀── DeepSeek V3 意图解析
│ (LLM 结构化提取) │       提取: Scene / Target / TimeRange / Extra
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Planner      │   ◀── DeepSeek V3 Think 模型 (深度推理)
│  (制定分析计划)   │       输出: steps[] 有序步骤列表
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│              Execute Loop                    │
│  ┌────────────┐                              │
│  │  Executor  │ ◀── DeepSeek V3 Quick 模型   │
│  │ (执行步骤)  │     + 工具集 (Prometheus /    │
│  │            │       CLS日志 / 文档检索)     │
│  └─────┬──────┘                              │
│        │ 步骤结果                              │
│        ▼                                     │
│  ┌─────────────┐                             │
│  │  Replanner  │ ◀── DeepSeek V3 Think 模型  │
│  │ (评估&调整)  │     判断: 继续 / 调整 / 完成  │
│  └─────┬───────┘                             │
│        │                                     │
│        ▼                                     │
│   [完成?] ──No──▶ 回到 Executor              │
│      │                                       │
│     Yes                                      │
└──────┼───────────────────────────────────────┘
       ▼
┌─────────────────┐
│  Response 汇总   │   ◀── 流式输出最终分析报告
│  (Markdown 报告) │
└─────────────────┘
```

#### 三种分析场景

| 场景 | 关键词示例 | 可用工具 | 输出 |
|:---|:---|:---|:---|
| 🚨 告警分析 | 「告警」「异常」「报警」 | Prometheus + 内部文档 + CLS 日志 | 告警诊断报告 |
| 📋 日志分析 | 「日志」「log」「错误日志」 | CLS 日志 (MCP) + 内部文档 | 日志分析报告 |
| 📊 性能分析 | 「性能」「延迟」「CPU」 | Prometheus + 内部文档 | 性能优化建议 |

#### Planner vs Executor 模型选择

| Agent | 模型 | 原因 |
|:---|:---|:---|
| **Planner** | DeepSeek V3 Think | 需要深度推理能力来制定分析计划 |
| **Executor** | DeepSeek V3 Quick | 快速执行工具调用，降低延迟 |
| **Replanner** | DeepSeek V3 Think | 需要评估执行结果并决定下一步 |

### 会话记忆管理

采用 **ConversationSummaryBufferMemory** 策略：

```
┌────────────────────────────────────────────────────────┐
│                    会话记忆                              │
│                                                        │
│   [摘要区]  ← LLM 压缩超出窗口的旧消息                   │
│   「用户先前询问了分发表达式的类型...」                     │
│                                                        │
│   [缓冲区]  ← 最近 6 轮原始消息完整保留                   │
│   User: 那实体表达式有哪些字段？                          │
│   AI: 实体表达式包含以下字段...                           │
│   User: 如何配置优先级？                                 │
│   AI: 优先级配置方式如下...                               │
│                                                        │
│   底层: bigcache (LRU, TTL=30min, Max=256MB)           │
└────────────────────────────────────────────────────────┘
```

---

## 🔧 工具链集成

| 工具 | 协议 | 数据源 | 用途 |
|:---|:---|:---|:---|
| `get_current_time` | 原生 | 系统时钟 | 获取当前时间（支持时区） |
| `query_internal_docs` | 原生 | Milvus | RAG 检索内部知识库文档 |
| `query_prometheus_alerts` | HTTP API | Prometheus | 查询活跃告警和指标数据 |
| CLS 日志查询 | **MCP** | 腾讯云 CLS | 通过 MCP 协议调用腾讯云日志服务 |
| `duckduckgo_search` | HTTP API | DuckDuckGo | 网络搜索（预留扩展） |

> **亮点**：日志工具通过 MCP (Model Context Protocol) 协议集成，展示了对前沿 AI 工具协议的实践运用。

---

## 📁 目录结构

```
super-agent/
├── main.go                          # 🚀 主入口（初始化 Agent + 启动 HTTP 服务）
├── go.mod / go.sum                  # 依赖管理
├── Makefile                         # 构建脚本
├── Dockerfile.backend               # 后端 Docker 镜像
├── Dockerfile.frontend              # 前端 Docker 镜像 (Nginx)
├── docker-compose.yml               # 容器编排
│
├── api/                             # 📐 API 接口定义
│   └── chat/
│       ├── chat.go                  # 路由接口定义 IChatV1
│       └── v1/chat.go              # 请求/响应结构体
│
├── cmd/                             # 🔧 独立命令入口
│   ├── aiops/                       # AIOps Agent 独立调试
│   ├── chat/                        # Chat Agent 独立调试
│   └── knowledge/                   # 知识库批量索引工具
│       └── docs/                    # 知识库源文档
│
├── entity/                          # 📦 全局常量定义
├── manifest/config/                 # ⚙️ 主配置文件 (YAML)
│
├── frontend/                        # 🎨 前端页面
│   ├── index.html                   # 页面结构
│   ├── app.js                       # 交互逻辑（SSE、会话管理、Markdown 渲染）
│   ├── style.css                    # 暗色主题样式
│   └── nginx.conf                   # Nginx 反向代理 + SSE 配置
│
├── internal/                        # 🏗 核心业务逻辑
│   ├── ai/
│   │   ├── agent/
│   │   │   ├── aiops/               # ★ AIOps Agent (Plan-Execute)
│   │   │   │   ├── runner.go        #   运行器（核心调度循环）
│   │   │   │   ├── planner.go       #   计划器 Agent
│   │   │   │   ├── executor.go      #   执行器 Agent
│   │   │   │   ├── replanner.go     #   重规划 Agent
│   │   │   │   ├── intent_parser.go #   自然语言意图解析
│   │   │   │   ├── prompt.go        #   场景 Prompt + 配置
│   │   │   │   └── orchestration.go #   图编排入口
│   │   │   ├── chat/                # ★ Chat Agent (ReAct)
│   │   │   │   ├── orchestration.go #   图编排（RAG + ReAct）
│   │   │   │   ├── flow.go          #   ReAct Agent 构建
│   │   │   │   ├── retriever.go     #   Milvus 检索器
│   │   │   │   └── prompt.go        #   系统 Prompt 模板
│   │   │   └── knowledge/           # ★ 知识库 ETL Pipeline
│   │   │       ├── orchestration.go #   ETL 图编排
│   │   │       ├── loader.go        #   文件加载器
│   │   │       ├── transformer.go   #   Markdown 分割器
│   │   │       ├── embedding.go     #   向量嵌入（千问）
│   │   │       └── indexer.go       #   Milvus 索引器
│   │   ├── models/deepseek.go       # LLM 模型管理（单例模式）
│   │   ├── mytool/                  # 自定义工具集
│   │   ├── embedder/                # Embedding 封装
│   │   ├── indexer/                 # 索引器封装
│   │   ├── loader/                  # 加载器封装
│   │   └── retriever/               # 检索器封装
│   ├── controller/chat/             # HTTP Controller 层
│   └── logic/sse/                   # SSE 连接管理
│
├── repo/milvus.go                   # Milvus 客户端连接 + Collection 管理
│
└── utils/
    ├── callback/                    # Eino 框架回调日志
    └── mem/                         # 会话记忆管理（bigcache + 摘要压缩）
```

---

## 🚀 快速开始

### 前置要求

- Go 1.23+
- Docker & Docker Compose
- Milvus 实例（可使用 Docker 部署）
- DeepSeek API Key（火山引擎）
- 阿里云 DashScope API Key（千问 Embedding）

### 本地开发

```bash
# 1. 克隆仓库
git clone <repo-url> && cd super-agent

# 2. 配置参数（编辑配置文件）
vim manifest/config/config.yaml

# 3. 运行服务
go run main.go
```

### Docker 部署

```bash
# 一键构建并启动
docker-compose up -d --build

# 服务访问
# 前端：http://localhost:3001
# 后端：http://localhost:6872
```

---

## ⚙️ 配置说明

配置文件位于 `manifest/config/config.yaml`：

```yaml
server:
  address: ":6872"                    # 后端服务端口

ai:
  api_key: "your-deepseek-api-key"    # DeepSeek API Key
  base_url: "https://ark.cn-beijing.volces.com/api/v3"
  model_quick: "ep-xxx-quick"         # 快速模型 (Executor 用)
  model_think: "ep-xxx-think"         # 思考模型 (Planner 用)

embedding:
  api_key: "your-dashscope-api-key"   # 千问 Embedding Key

milvus:
  address: "172.17.0.1:19530"         # Milvus 连接地址
  collection: "super_agent"           # Collection 名称
```

---

## 🔮 后续优化

### 架构层面
- [ ] **Agent 记忆持久化**：当前基于 bigcache 的会话记忆在重启后丢失，计划接入 Redis 或数据库持久化
- [ ] **多模型适配**：抽象 LLM 接口，支持 GPT-4o、Claude、千问等更多模型的快速切换
- [ ] **分布式部署**：当前为单机部署，计划支持多实例水平扩展 + 负载均衡

### 功能层面
- [ ] **更多运维工具**：集成 K8s API、Grafana 看板截图、告警自动恢复等能力
- [ ] **知识库增强**：支持更多文档格式（PDF、Confluence），增加自动知识抽取
- [ ] **用户认证**：增加用户登录体系，支持多租户隔离
- [ ] **Agent 可观测性**：增加完整的 Agent 执行链路追踪和耗时分析面板

### 体验层面
- [ ] **前端重构**：使用 React/Vue 重构前端，增加更丰富的交互组件
- [ ] **对话评分**：用户可对 AI 回答进行评分，形成反馈闭环
- [ ] **提示词优化**：基于用户反馈，持续优化各 Agent 的 System Prompt

---

## 📄 License

本项目仅供学习和技术交流使用。
