package main

import (
	"context"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"net/url"
)

func AIEcho(ctx context.Context, q url.Values) interface{} {
	key := q.Get("key")
	arg0, arg1, arg2, arg3, arg4 := q.Get("arg0"), q.Get("arg1"), q.Get("arg2"), q.Get("arg3"), q.Get("arg4")
	ol.Tf(ctx, "AI echo key=%v, args=%v", key, []string{arg0, arg1, arg2, arg3, arg4})

	switch key {
	case "depends_env":
		return DependsEnv(ctx, arg0)
	}
	return "Not sure"
}

// http://one-voice-platform.alibaba-inc.com/robot/entityDetail.htm?appId=21516&id=8759
func DependsEnv(ctx context.Context, env string) string {
	switch env {
	case "CentOS":
		return "很好，官方支持"
	case "Linux":
	}
	return "Not sure"
}
