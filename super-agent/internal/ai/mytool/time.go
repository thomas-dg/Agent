package mytool

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetCurrentTimeInput 获取当前时间的输入参数
type GetCurrentTimeInput struct {
	Timezone string `json:"timezone" jsonschema:"description=时区名称，例如 Asia/Shanghai、UTC，默认为 Asia/Shanghai"`
}

// GetCurrentTimeOutput 获取当前时间的输出结果
type GetCurrentTimeOutput struct {
	CurrentTime string `json:"current_time" jsonschema:"description=RFC3339 格式的当前时间，包含时区偏移信息，例如 2026-03-07T12:00:00+08:00"`
	Timezone    string `json:"timezone"     jsonschema:"description=实际使用的时区名称"`
	Unix        int64  `json:"unix"         jsonschema:"description=Unix 时间戳（秒）"`
}

// NewGetCurrentTimeTool 创建获取当前时间的工具
func NewGetCurrentTimeTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"get_current_time",
		"获取当前系统时间，返回 RFC3339 格式的时间字符串及 Unix 时间戳。可指定时区，默认为 Asia/Shanghai。",
		func(ctx context.Context, input *GetCurrentTimeInput, opts ...tool.Option) (string, error) {
			tz := input.Timezone
			if tz == "" {
				tz = "Asia/Shanghai"
			}

			loc, err := time.LoadLocation(tz)
			if err != nil {
				return "", fmt.Errorf("无效的时区 %q: %w", tz, err)
			}

			now := time.Now().In(loc)
			output := &GetCurrentTimeOutput{
				CurrentTime: now.Format(time.RFC3339),
				Timezone:    tz,
				Unix:        now.Unix(),
			}

			return fmt.Sprintf(`{"current_time":%q,"timezone":%q,"unix":%d}`,
				output.CurrentTime, output.Timezone, output.Unix), nil
		},
	)
	if err != nil {
		log.Fatalf("初始化 get_current_time 工具失败: %v", err)
	}

	return t
}
