package main

import (
	"bytes"
	"context"
	"fmt"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"html/template"
	"net/url"
	"strings"
)

const (
	NotSure          = "不确定，请参考下面的说明，或换个提问方式"
	UnknownKnowledge = "不知道，请换个提问方式"
)

func AIEcho(ctx context.Context, q url.Values) (interface{}, error) {
	key := q.Get("key")
	arg0, arg1, arg2, arg3, arg4 := q.Get("arg0"), q.Get("arg1"), q.Get("arg2"), q.Get("arg3"), q.Get("arg4")
	ol.Tf(ctx, "AI echo key=%v, args=%v", key, []string{arg0, arg1, arg2, arg3, arg4})

	var rr []string
	switch key {
	case "depends_env":
		rr = DependsEnv(ctx, arg0)
	case "av_codecs":
		rr = AVCodec(ctx, arg0)
	default:
		return UnknownKnowledge, nil
	}

	const tpl = "" +
		`{{if eq (len .Items) 1}}{{index .Items 0}}` +
		`{{else}}` + "{{range .Items}}\n- {{.}}{{end}}" +
		`{{end}}`

	t, err := template.New("echo").Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("parse %v", tpl)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, &struct {
		Items []string
	}{
		rr,
	}); err != nil {
		return nil, err
	}

	return b.String(), nil
}

// 实体：AVCodec
func AVCodec(ctx context.Context, env string) []string {
	switch env {
	default:
		return nil
	}
}

// 实体：依赖环境
func DependsEnv(ctx context.Context, env string) []string {
	fn := func(env string) string {
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
		default:
			return NotSure
		}
	}

	rr := []string{}
	for _, env := range strings.Split(env, ",") {
		rr = append(rr, fmt.Sprintf("**%v:** %v", env, fn(env)))
	}
	return rr
}
