# 🚀 Code Buddy Agent

<p align="center">
  <strong>AI Agent 技术实践合集 —— 从 Prompt Engineering 到多 Agent 协作系统</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/AI_Framework-Eino-FF6B6B?style=flat-square" alt="Eino">
  <img src="https://img.shields.io/badge/LLM-DeepSeek_V3-4A90D9?style=flat-square" alt="DeepSeek">
  <img src="https://img.shields.io/badge/Pattern-Multi_Agent-00B4AB?style=flat-square" alt="Multi-Agent">
  <img src="https://img.shields.io/badge/Protocol-MCP-FF9500?style=flat-square" alt="MCP">
</p>

---

## 🎯 项目概述

本仓库是 **AI Agent** 技术方向的实践项目集，包含两个核心子项目，体现了从 Prompt Engineering（提示词工程）到完整 Agent 系统（多 Agent 协作 + 工具调用 + RAG）的技术演进路径：

```
Code Buddy Agent
├── 🤖 super-agent    →   智能 OnCall 运维助手（完整的多 Agent 系统）
└── ✏️  coder          →   AI 编码 Agent 提示词工程（Prompt Engineering 方法论）
```

---

## 📦 子项目一览

### 🤖 [Super Agent](./super-agent/) — 智能 OnCall 助手

> **🌐 在线体验：[http://49.232.223.185:3001](http://49.232.223.185:3001)**

一个面向运维场景的智能 AI 助手，采用多 Agent 协作架构，集成了当前 AI Agent 领域的核心技术：

|    能力    | 技术方案 | 说明 |
|:--------:|:---:|:---|
| 🧠 智能问答  | RAG + ReAct Agent | 知识检索增强 + 工具增强推理 |
| 🔧 AIOps | Plan-Execute Agent | LLM 驱动的自动化分析流水线 |
| 📚 知识管理  | ETL Pipeline | 文档加载 → 分割 → 向量化 → 存储 |
| 🔌 工具集成  | MCP Protocol | Prometheus / CLS 日志 / 文档检索 |

**技术亮点**：
- 基于 [Eino](https://github.com/cloudwego/eino)（字节跳动开源）的 **Graph 编排** 实现多 Agent 协作
- **Plan-Execute** 模式：Planner(深度推理) → Executor(工具调用) → Replanner(动态调整)
- **ReAct** 模式：RAG 检索 + 多步工具推理，流式输出
- MCP 协议集成腾讯云 CLS 日志服务
- SSE 全链路流式响应
- Docker Compose 一键部署

📖 [查看完整技术文档 →](./super-agent/README.md)

---

### ✏️ [Coder](./coder/) — AI 编码 Agent 提示词工程

一套系统化的 Prompt Engineering 方法论实践，定义了一个高质量的 Go 代码重构 Agent：

| 模块 | 内容 | 说明 |
|:---:|:---:|:---|
| 🔧 重构 Agent | System Prompt | Role + Goals + Rules + Workflow 完整定义 |
| 📚 知识库 | Go Best Practices | Go 1.24+ 最佳实践（8 大章节） |
| 📊 评测体系 | Case + User Prompt | 真实代码案例对比 + 标准化评测 |
| 📋 变更日志 | CHANGELOG | 数据驱动的 Prompt 迭代记录 |

**技术亮点**：
- **评测驱动**的 Prompt 优化循环：设计 → 评测 → 分析 → 改进
- 知识库 (RAG) 解决领域规范类问题，将最佳实践注入 Agent
- Clean Architecture 架构约束，确保生成代码的工程质量
- 版本化管理，清晰记录每次迭代的改进点

📖 [查看完整文档 →](./coder/README.md)



---

## 📁 仓库结构

```
code-buddy-agent/
├── README.md                 # 📌 本文件（仓库概要）
├── super-agent/              # 🤖 智能 OnCall 运维助手
│   ├── main.go               #    主入口
│   ├── internal/ai/agent/    #    核心 Agent 实现
│   ├── frontend/             #    前端页面
│   └── ...                   #    详见 super-agent/README.md
└── coder/                    # ✏️ AI 编码 Agent 提示词工程
    ├── Refactoring/          #    重构 Agent 定义 + 评测
    └── knowledge/            #    Go 最佳实践知识库
```

---

## 📄 License

本项目仅供学习和技术交流使用。
