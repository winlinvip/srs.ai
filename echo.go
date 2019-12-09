package main

import (
	"context"
	"encoding/json"
	"fmt"
	fc "github.com/aliyun/fc-go-sdk"
	ol "github.com/ossrs/go-oryx-lib/logger"
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

func AIEcho(ctx context.Context, q url.Values) (interface{}, error) {
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

	ol.Tf(ctx, "AI echo args=%v, fc %v cost %v", args, fcResponse, fcDuration)
	return fcResponse, nil
}
