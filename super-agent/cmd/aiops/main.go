package main

import (
	"context"
	"fmt"
	"super-agent/internal/ai/agent/aiops"
)

func main() {
	ctx := context.Background()
	opts := aiops.Options{
		Scene: aiops.SceneAlertAnalysis,
		Extra: "帮我分析当前所有活跃告警，给出处理建议",
	}
	resp, err := aiops.BuildPlanAgent(ctx, opts)
	if err != nil {
		panic(err)
	}
	fmt.Println("----- Final Response -----")
	fmt.Println(resp)
}
