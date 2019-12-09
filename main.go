package main

import (
	"context"
	"flag"
	"fmt"
	fc "github.com/aliyun/fc-go-sdk"
	oe "github.com/ossrs/go-oryx-lib/errors"
	oh "github.com/ossrs/go-oryx-lib/http"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"net/http"
	"os"
	"strings"
)

func main() {
	var listen string
	flag.StringVar(&listen, "listen", "", "The listen ip:port.")

	flag.StringVar(&fcAKID, "akid", "", "The AKID for FC")
	flag.StringVar(&fcAKSecret, "aksecret", "", "The AKSecret for FC")
	flag.StringVar(&fcEndpoint, "endpoint", "", "The endpoint for FC")

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("AI/%v service for SRS", Version()))
		fmt.Println(fmt.Sprintf("Usage: %v [options]", os.Args[0]))
		fmt.Println(fmt.Sprintf("Options:"))
		fmt.Println(fmt.Sprintf("	-listen string"))
		fmt.Println(fmt.Sprintf("		The listen [ip]:port. Empty ip means 0.0.0.0, any interface."))
		fmt.Println(fmt.Sprintf("	-akid string"))
		fmt.Println(fmt.Sprintf("	-aksecret string"))
		fmt.Println(fmt.Sprintf("		The AK(AccessKey) ID and Secret for FC(Function Compute)."))
		fmt.Println(fmt.Sprintf("	-endpoint string"))
		fmt.Println(fmt.Sprintf("		The endpoint for FC."))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v -listen=:1988 -akid=xxx -aksecret=xxx -endpoint=xxx", os.Args[0]))
	}

	flag.Parse()

	if listen == "" || fcAKID == "" || fcAKSecret == "" {
		flag.Usage()
		os.Exit(-1)
	}

	if !strings.Contains(listen, ":") {
		listen = "0.0.0.0:" + listen
	}

	ctx := context.Background()
	oh.Server = fmt.Sprintf("AI/%v", Version())
	ol.Tf(ctx, "SRS AI/%v listen=%v, ak=<%v %v %v>", Version(), listen, fcAKID, fcAKSecret, fcEndpoint)

	if client, err := fc.NewClient(fcEndpoint, "2016-08-15", fcAKID, fcAKSecret); err != nil {
		ol.Ef(ctx, "fc err %+v", err)
		os.Exit(-1)
	} else {
		fcClient = client
	}

	pattern := "/ai/v1/versions"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		oh.WriteVersion(w, r, Version())
	})

	pattern = "/ai/v1/echo"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			oh.WriteError(ctx, w, r, oe.Wrapf(err, "parse %v", r.URL))
			return
		}

		// Query string.
		q := r.Form
		rr := make(map[string]interface{})
		for k, _ := range q {
			if strings.HasPrefix(k, "sys.ding.") {
				continue
			}

			if v := q.Get(k); v != "" && v != "nil" {
				rr[k] = q.Get(k)
			}
		}

		if result, err := AIEcho(ctx, q); err != nil {
			oh.WriteError(ctx, w, r, oe.Wrapf(err, "parse %v of %v", r.URL, q))
			return
		} else {
			rr["result"] = result
		}

		ol.Tf(ctx, "Echo %v with %v", r.URL, rr)
		oh.WriteData(ctx, w, r, rr)
	})

	oh.FilterData = func(ctx ol.Context, w http.ResponseWriter, r *http.Request, o interface{}) interface{} {
		return &struct {
			Success      bool        `json:"success"`
			ErrorCode    int         `json:"errorCode"`
			ErrorMessage string      `json:"errorMsg"`
			Fields       interface{} `json:"fields"`
		}{
			true, 200, "Success", o,
		}
	}

	http.ListenAndServe(listen, nil)
}
