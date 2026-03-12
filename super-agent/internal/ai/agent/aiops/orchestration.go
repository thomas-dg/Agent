package aiops

import (
	"context"
	"fmt"
)

// BuildPlanAgent 构建并运行 Plan-Execute Agent。
// 注意：此函数每次调用都会重新初始化所有 Agent，仅供测试/脚本使用。
// 生产环境请使用 NewRunner 预构建 Runner 后调用 Runner.Run。
func BuildPlanAgent(ctx context.Context, opts Options) (string, error) {
	r, err := NewRunner(ctx)
	if err != nil {
		return "", fmt.Errorf("BuildPlanAgent: 初始化 Runner 失败: %w", err)
	}
	return r.Run(ctx, opts)
}
