package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	fc "github.com/aliyun/fc-go-sdk"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"io"
	"net/http"
	"net/url"
	"os/exec"
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

// Limit the client for N request per second.
var limitEcho chan bool

func init() {
	limitEcho = make(chan bool, 1)
	go func() {
		for {
			limitEcho <- true
			time.Sleep(1100 * time.Millisecond)
		}
	}()
}

func fcExecute(input []byte) ([]byte, error) {
	fcIn := fc.NewInvokeFunctionInput(fcServiceName, fcFuncName)
	fcIn.Payload = &input

	fcOut, err := fcClient.InvokeFunction(fcIn)
	if err != nil {
		return nil, fmt.Errorf("invoke fc %v %v err %v", fcServiceName, fcFuncName, err)
	}

	return fcOut.Payload, nil
}

func directExecute(input []byte) ([]byte, error) {
	cmd := exec.Command("python", "fc.py", string(input))
	return cmd.Output()
}

func AIEcho(ctx context.Context, closeNotify <-chan bool, r *http.Request, q, qFiltered url.Values) (interface{}, error) {
	ts := time.Now()

	select {
	case <-limitEcho:
	case <-closeNotify:
		return nil, context.Canceled
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	limitDuration := time.Now().Sub(ts)

	args := make(map[string]string)
	for k, _ := range q {
		args[k] = q.Get(k)
	}
	bb, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal %v", args)
	}

	//fcResponse, err := fcExecute(bb)
	fcResponse, err := directExecute(bb)
	if err != nil {
		return nil, fmt.Errorf("fc with %v err %v", string(bb), err)
	}
	fcDuration := time.Now().Sub(ts)

	qq := make(map[string]string)
	qq["__tag__:__client_ip__"] = GetOriginalClientIP(r)
	qq["oua"] = r.Header.Get("User-Agent")
	qq["__userAgent__"] = "agent"
	qq["site"] = "dingtalk.com"
	qq["path"] = fmt.Sprintf("/robot/%v", q.Get("key"))
	if query, err := url.QueryUnescape(qFiltered.Encode()); err == nil {
		qq["query"] = query
	}
	qq["cost"] = fmt.Sprint(int64(fcDuration / time.Millisecond))
	referer := r.Header.Get("Referer")
	qq["oreferer"] = referer
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			referer = u.Host
		}
		qq["__referer__"] = referer
	}
	var qqb bytes.Buffer
	if enc := json.NewEncoder(&qqb); true {
		enc.SetEscapeHTML(false)
		if err := enc.Encode(qq); err != nil {
			return nil, err
		}
	}
	if _, err := io.WriteString(gLogfile, qqb.String()); err != nil {
		return nil, err
	}

	var qqbs string
	if qqb.Len() > 1 {
		// Trim the last \n.
		qqbs = string(qqb.Next(qqb.Len() - 1))
	}

	ol.Tf(ctx, "AI echo log %v, args=%v, limit cost=%v, fc cost=%v result is %v",
		qqbs, args, limitDuration, fcDuration, string(fcResponse))
	return string(fcResponse), nil
}
