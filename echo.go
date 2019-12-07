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

	fnArg0 := func(pfn func(context.Context, string) string, ctx context.Context, arg0 string) []string {
		rr := []string{}

		keys := make(map[string]bool)
		for _, arg0 := range strings.Split(arg0, ",") {
			// Filter the duplicated key.
			if _, ok := keys[arg0]; ok {
				continue
			}
			keys[arg0] = true

			rr = append(rr, fmt.Sprintf("**%v:** %v", arg0, pfn(ctx, arg0)))
		}
		return rr
	}

	var rr []string
	switch key {
	case "depends_env":
		rr = fnArg0(DependsEnv, ctx, arg0)
	case "av_codecs":
		rr = fnArg0(AVCodec, ctx, arg0)
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
func AVCodec(ctx context.Context, env string) string {
	switch env {
	case "H.264", "AAC":
		return "很好，各种浏览器和平台都支持"
	case "H.263", "SPEEX", "PCM":
		return "不支持，[太老的编码格式](https://github.com/ossrs/srs/blob/b4870a6d6f94ad26c7cc9c7fb39a4246180b5864/trunk/src/kernel/srs_kernel_codec.hpp#L35)，建议用[FFMPEG](https://github.com/ossrs/srs/wiki/v1_CN_SampleFFMPEG)转码为h.264"
	case "H.265":
		return "不支持，浏览器支持得不好，参考[#465](https://github.com/ossrs/srs/issues/465#issuecomment-562794207)"
	case "AV1":
		return "不支持，是未来的趋势，参考[#1070](https://github.com/ossrs/srs/issues/1070#issuecomment-562794926)"
	case "MP3":
		return "部分支持，不推荐，参考[#301](https://github.com/ossrs/srs/issues/301)和[#296](https://github.com/ossrs/srs/issues/296)"
	case "Opus":
		return "不支持，是WebRTC的音频编码，参考[#307](https://github.com/ossrs/srs/issues/307)"
	case "SRT":
		return "不支持，是广电常用的协议，参考[#1147](https://github.com/ossrs/srs/issues/1147)"
	default:
		return NotSure
	}
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
	default:
		return NotSure
	}
}
