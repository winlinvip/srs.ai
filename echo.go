package main

import (
	"context"
	"encoding/json"
	"fmt"
	fc "github.com/aliyun/fc-go-sdk"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	fcServiceName = "srs-ai"
	fcFuncName    = "ai_echo"
)

var (
	// By config.
	fcAKID, fcAKSecret, fcEndpoint string
	// Global variable.
	fcClient *fc.Client
)

func AIEcho(ctx context.Context, r *http.Request, q url.Values) (interface{}, error) {
	ts := time.Now()

	args := make(map[string]string)
	for k, _ := range q {
		args[k] = q.Get(k)
	}
	bb, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal %v", args)
	}

	fcIn := fc.NewInvokeFunctionInput(fcServiceName, fcFuncName)
	fcIn.Payload = &bb

	fcOut, err := fcClient.InvokeFunction(fcIn)
	if err != nil {
		return nil, fmt.Errorf("invoke fc %v %v err %v", fcServiceName, fcFuncName, err)
	}
	fcDuration := time.Now().Sub(ts)
	fcResponse := string(fcOut.Payload)

	qq := make(map[string]string)
	qq["__tag__:__client_ip__"] = GetOriginalClientIP(r)
	qq["oua"] = r.Header.Get("User-Agent")
	qq["__userAgent__"] = "agent"
	qq["site"] = "dingtalk.com"
	qq["path"] = "/dingtalk/robot"
	qq["cost"] = fmt.Sprint(int64(fcDuration / time.Millisecond))
	referer := r.Header.Get("Referer")
	qq["oreferer"] = referer
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			referer = u.Host
		}
		qq["__referer__"] = referer
	}
	if bb, err = json.Marshal(qq); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(gLogfile, string(bb)+"\n"); err != nil {
		return nil, err
	}

	ol.Tf(ctx, "AI echo log %v, args=%v, fc cost %v result is %v", string(bb), args, fcDuration, fcResponse)
	return fcResponse, nil
}
