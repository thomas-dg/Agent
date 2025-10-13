# Agent

AI Agent 资源仓库，包含 System Prompt、评测体系和知识库。
## 核心能力
- **coder Agent**：
  - refactoring：重构Agent，根据重构方案、mermaid流程图等上下文，完成go项目重构
  - flowchart：mermaid流程图 Agent，根据上下文，生成mermaid流程图
- **knowledge（知识库）**：
  - go_best_practices.md：go 最佳实践


## 为什么会有这个项目？

随着LLM能力不断提升，AI编程已经从copilot升级到vibe后，但实际使用发现，AI还是有很多不足，特别对需要长期维护、不断演进的项目（相对weekend project而言）。
秉承着“人类为主，AI辅助”的原则，希望AI在人类指定的规则范围内进行工作，避免“失控”。


## 目录结构

```
coder/
├── refactoring/              # 重构Agent（go项目）
│   ├── system_prompt.md      # 核心提示词
│   ├── requirements.md       # Agent 依赖信息
│   ├── CHANGELOG.md          # 版本变更记录
│   └── evaluating/           # 评测信息
├── flowchart/                # mermaid 流程图Agent
│   ├── system_prompt.md      # 核心提示词
│   └── requirements.md       # Agent 依赖信息
knowledge/
└── go_best_practices.md      # Go 最佳实践知识库
```

## 快速开始
Agent本身包括Prompt、插件、知识库等，本仓库侧重Prompt，其他依赖项放在Agent 同目录的requirements.md文件中。

1. 选择需要使用的Agent。

2. 根据 system_prompt.md 和 requirements.md 搭建Agent

3. 使用新搭建的Agent进行工作。



## 版本历史

### refactoring Agent 版本历史

- **v0.0.3**：强化 Go 最佳实践，增加人类代码优先权
- **v0.0.2**：增加目录规范，解决 Java 风格问题  
- **v0.0.1**：基础重构能力

详见：[CHANGELOG.md](coder/refactoring/CHANGELOG.md)

### mermaid 流程图Agent 版本历史

- **v0.0.1**：基础流程图能力

详见：[CHANGELOG.md](coder/flowchart/CHANGELOG.md)



