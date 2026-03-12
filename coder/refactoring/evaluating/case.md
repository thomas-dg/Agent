

实现偏复杂
```go
package pipeline

import (
	"context"
	"fmt"

	"git.code.oa.com/qqcms/bypass_content/src/repo/pipelines/plugin"
	"git.code.oa.com/qqcms/bypass_content/src/repo/pipelines/stage"
)

// BasePipeline 通用Pipeline基础结构
type BasePipeline struct {
	name   string
	stages []stage.Stage
}

// Name 返回流程名称
func (p *BasePipeline) Name() string {
	return p.name
}

// Execute 执行流程中的所有步骤
func (p *BasePipeline) Execute(ctx context.Context, msg plugin.Message) error {
	for _, s := range p.stages {
		if err := s.Execute(ctx, msg); err != nil {
			return fmt.Errorf("pipeline %s stage %s execute failed: %w", p.name, s.Name(), err)
		}
	}
	return nil
}

// StageFactory Stage工厂函数类型
type StageFactory func(*plugin.PluginRegistry) stage.Stage

// PipelineConfig Pipeline配置结构
type PipelineConfig struct {
	Name   string
	Stages []StageFactory
}

// NewPipelineFromConfig 根据配置创建Pipeline
func NewPipelineFromConfig(config PipelineConfig, registry *plugin.PluginRegistry) Pipeline {
	stages := make([]stage.Stage, 0, len(config.Stages))
	for _, stageFactory := range config.Stages {
		stages = append(stages, stageFactory(registry))
	}

	return &BasePipeline{
		name:   config.Name,
		stages: stages,
	}
}

// NewStaticPipeline 创建静态增量流分发处理流程
func NewStaticPipeline(registry *plugin.PluginRegistry) Pipeline {
	config := PipelineConfig{
		Name: "static_increment_pipeline",
		Stages: []StageFactory{
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewGetStaticJoinDataStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataFilterStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataProcessStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataDeliveryStage(r) },
		},
	}
	return NewPipelineFromConfig(config, registry)
}

// NewRelationPipeline 创建关系流分发处理流程
func NewRelationPipeline(registry *plugin.PluginRegistry) Pipeline {
	config := PipelineConfig{
		Name: "relation_pipeline",
		Stages: []StageFactory{
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataFilterStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataProcessStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataDeliveryStage(r) },
		},
	}
	return NewPipelineFromConfig(config, registry)
}

// NewDispatchPipeline 创建通用分发处理流程
func NewDispatchPipeline(registry *plugin.PluginRegistry) Pipeline {
	config := PipelineConfig{
		Name: "dispatch_pipeline",
		Stages: []StageFactory{
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewGetStaticJoinDataStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataFilterStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataProcessStage(r) },
			func(r *plugin.PluginRegistry) stage.Stage { return stage.NewDataDeliveryStage(r) },
		},
	}
	return NewPipelineFromConfig(config, registry)
}

```

```go
package pipeline

import (
	"context"
	"fmt"

	"git.code.oa.com/qqcms/bypass_content/src/repo/pipelines/plugin"
	"git.code.oa.com/qqcms/bypass_content/src/repo/pipelines/stage"
)

// BasePipeline 通用Pipeline基础结构
type BasePipeline struct {
	name   string
	stages []stage.Stage
}

// Name 返回流程名称
func (p *BasePipeline) Name() string {
	return p.name
}

// Execute 执行流程中的所有步骤
func (p *BasePipeline) Execute(ctx context.Context, msg plugin.Message) error {
	for _, s := range p.stages {
		if err := s.Execute(ctx, msg); err != nil {
			return fmt.Errorf("pipeline %s stage %s execute failed: %w", p.Name, s.Name(), err)
		}
	}
	return nil
}

type static struct {
	*BasePipeline
}

// NewStaticPipeline 创建静态增量流分发处理流程
func NewStaticPipeline(registry *plugin.PluginRegistry) Pipeline {
	stages := []stage.Stage{
		stage.NewGetStaticJoinDataStage(registry),
		stage.NewDataFilterStage(registry),
		stage.NewDataProcessStage(registry),
		stage.NewDataDeliveryStage(registry),
	}

	return &static{
		BasePipeline: &BasePipeline{
			name:   "static_pipeline",
			stages: stages,
		},
	}
}

type relation struct {
	*BasePipeline
}

// NewRelationPipeline 创建关系流分发处理流程
var NewRelationPipeline = func(registry *plugin.PluginRegistry) Pipeline {
	stages := []stage.Stage{
		stage.NewDataFilterStage(registry),
		stage.NewDataProcessStage(registry),
		stage.NewDataDeliveryStage(registry),
	}

	return &relation{
		BasePipeline: &BasePipeline{
			name:   "relation_pipeline",
			stages: stages,
		},
	}
}

```
