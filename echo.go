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

// 实体：依赖环境
func DependsEnv(ctx context.Context, env string) string {
	switch env {
	case "CentOS", "x86-64":
		return "很好，官方支持，建议直接用docker运行，参考[这里](https://github.com/ossrs/srs-docker/tree/srs3)"
	case "Linux", "Unix":
		return "可以，建议用docker[编译调试](https://github.com/ossrs/srs-docker/tree/dev)和[运行](https://github.com/ossrs/srs-docker/tree/srs3)"
	case "ARM":
		return "可以，需要替换ST(state-threads)，参考[这里](https://github.com/ossrs/state-threads/tree/srs#usage)"
	case "Windows":
		return "不支持，不过可以用docker运行，参考[这里](https://github.com/ossrs/srs/wiki/v1_CN_WindowsSRS)"
	case "thingOS":
		return "不支持"
	}
	return "Not sure"
}
