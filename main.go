package main

import (
	"context"
	"flag"
	"fmt"
	fc "github.com/aliyun/fc-go-sdk"
	oe "github.com/ossrs/go-oryx-lib/errors"
	oh "github.com/ossrs/go-oryx-lib/http"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var gLogfile *os.File

func main() {
	var listen string
	flag.StringVar(&listen, "listen", "", "The listen ip:port.")

	var logfile string
	flag.StringVar(&logfile, "log", "", "Log file path. Default: stdout")

	flag.StringVar(&fcAKID, "akid", "", "The AKID for FC")
	flag.StringVar(&fcAKSecret, "aksecret", "", "The AKSecret for FC")
	flag.StringVar(&fcEndpoint, "endpoint", "", "The endpoint for FC")

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("AI/%v service for SRS", Version()))
		fmt.Println(fmt.Sprintf("Usage: %v [options]", os.Args[0]))
		fmt.Println(fmt.Sprintf("Options:"))
		fmt.Println(fmt.Sprintf("	-listen string"))
		fmt.Println(fmt.Sprintf("		The listen [ip]:port. Empty ip means 0.0.0.0, any interface."))
		fmt.Println(fmt.Sprintf("	-log string"))
		fmt.Println(fmt.Sprintf("		The log file path. Default: stdout."))
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
	ol.Tf(ctx, "SRS AI/%v listen=%v, log=%v, ak=<%v %v %v>", Version(), listen, logfile, fcAKID, fcAKSecret, fcEndpoint)

	if logfile == "" {
		gLogfile = os.Stdout
	} else {
		if lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			ol.Ef(ctx, "Open %v err %v", logfile, err)
			os.Exit(-1)
		} else {
			defer lf.Close()
			gLogfile = lf
		}
	}

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
		q, rr, qFiltered := r.Form, make(map[string]interface{}), url.Values{}
		for k, _ := range q {
			if strings.HasPrefix(k, "sys.ding.") {
				continue
			}

			if v := q.Get(k); v != "" && v != "nil" {
				rr[k] = q.Get(k)
				qFiltered.Set(k, v)
			}
		}

		closeNotify := w.(http.CloseNotifier).CloseNotify()
		if result, err := AIEcho(ctx, closeNotify, r, q, qFiltered); err != nil {
			if oe.Cause(err) == context.Canceled {
				oh.WriteData(ctx, w, r, "Canceled")
			} else {
				oh.WriteError(ctx, w, r, oe.Wrapf(err, "parse %v of %v", r.URL, q))
			}
			return
		} else {
			rr["result"] = result
		}

		ol.Tf(ctx, "Echo %v headers=%v, q=%v with %v", r.URL, r.Header, q, rr)
		oh.WriteData(ctx, w, r, rr)
	})

	pattern = "/ai/v1/stat"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			oh.WriteError(ctx, w, r, oe.Wrapf(err, "parse %v", r.URL))
			return
		}

		// Query string.
		q, rr, qFiltered := r.Form, make(map[string]interface{}), url.Values{}
		for k, _ := range q {
			if strings.HasPrefix(k, "sys.ding.") {
				continue
			}

			if v := q.Get(k); v != "" && v != "nil" {
				rr[k] = q.Get(k)
				qFiltered.Set(k, v)
			}
		}

		if result, err := HTTPStat(ctx, r, q, qFiltered); err != nil {
			oh.WriteError(ctx, w, r, oe.Wrapf(err, "parse %v of %v", r.URL, q))
			return
		} else {
			rr["result"] = result
		}

		ol.Tf(ctx, "Stat %v headers=%v, q=%v with %v", r.URL, r.Header, q, rr)
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

func GetOriginalClientIP(r *http.Request) string {
	// https://gtranslate.io/forum/http-real-http-forwarded-for-t2980.html
	//current the order to get client ip is clientip > X-Forwarded-For > X-Real-IP > remote addr
	var rip string

	q := r.URL.Query()
	if rip = q.Get("clientip"); rip != "" {
		return rip
	}

	if forwordIP := r.Header.Get("X-Forwarded-For"); forwordIP != "" {
		index := strings.Index(forwordIP, ",")
		if index != -1 {
			rip = forwordIP[:index]
		} else {
			rip = forwordIP
		}
		return rip
	}

	if rip = r.Header.Get("X-Real-IP"); rip == "" {
		if nip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			rip = nip
		}
	}

	return rip
}
