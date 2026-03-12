package callback

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cloudwego/eino/callbacks"
)

type LogCallbackConfig struct {
	Detail bool
	Debug  bool
}

func LogCallback(conf *LogCallbackConfig) callbacks.Handler {
	if conf == nil {
		conf = &LogCallbackConfig{
			Detail: true,
		}
	}

	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		log.Printf("[view start]:[%s:%s:%s]\n", info.Component, info.Type, info.Name)
		if conf.Detail {
			var b []byte
			if conf.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			log.Printf("%s\n", string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		log.Printf("[view end]:[%s:%s:%s]\n", info.Component, info.Type, info.Name)
		return ctx
	})

	return builder.Build()
}
