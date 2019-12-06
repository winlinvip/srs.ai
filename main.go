package main

import (
	"context"
	"flag"
	"fmt"
	oe "github.com/ossrs/go-oryx-lib/errors"
	oh "github.com/ossrs/go-oryx-lib/http"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	var listen string
	flag.StringVar(&listen, "listen", "", "The listen ip:port.")

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("AI/%v service for SRS", Version()))
		fmt.Println(fmt.Sprintf("Usage: %v [options]", os.Args[0]))
		fmt.Println(fmt.Sprintf("Options:"))
		fmt.Println(fmt.Sprintf("	-listen string"))
		fmt.Println(fmt.Sprintf("		The listen [ip]:port. Empty ip means 0.0.0.0, any interface."))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v -listen=:1988", os.Args[0]))
	}

	flag.Parse()

	if listen == "" {
		flag.Usage()
		os.Exit(-1)
	}

	if !strings.Contains(listen, ":") {
		listen = "0.0.0.0:" + listen
	}

	ctx := context.Background()
	oh.Server = fmt.Sprintf("AI/%v", Version())
	ol.Tf(ctx, "SRS AI/%v listen=%v", Version(), listen)

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
			if v := q.Get(k); v != "" && v != "nil" {
				rr[k] = q.Get(k)
			}
		}

		rr["result"] = AIEcho(ctx, q)

		ol.Tf(ctx, "Echo %v with %v", r.URL, rr)
		oh.WriteData(ctx, w, r, rr)
	})

	// https://yuque.antfin-inc.com/bot_factory/botfactory/service_use
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

func AIEcho(ctx context.Context, q url.Values) interface{} {
	key := q.Get("key")
	arg0, arg1, arg2, arg3, arg4 := q.Get("arg0"), q.Get("arg1"), q.Get("arg2"), q.Get("arg3"), q.Get("arg4")
	ol.Tf(ctx, "AI echo key=%v, args=%v", key, []string{arg0, arg1, arg2, arg3, arg4})

	return key
}
