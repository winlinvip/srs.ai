package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"io"
	"net/http"
	"net/url"
	"time"
)

func HTTPStat(ctx context.Context, r *http.Request, q, qFiltered url.Values) (interface{}, error) {
	ts := time.Now()

	args := make(map[string]string)
	for k, _ := range q {
		args[k] = q.Get(k)
	}

	qq := make(map[string]string)
	qq["__tag__:__client_ip__"] = GetOriginalClientIP(r)
	qq["oua"] = r.Header.Get("User-Agent")
	qq["__userAgent__"] = "agent"
	qq["site"] = "dingtalk.com"
	qq["path"] = fmt.Sprintf("/robot/%v", q.Get("key"))
	if query, err := url.QueryUnescape(qFiltered.Encode()); err == nil {
		qq["query"] = query
	}
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

	statDuration := time.Now().Sub(ts)
	ol.Tf(ctx, "AI echo log %v, args=%v, stat cost %v", qqbs, args, statDuration)
	return "Success", nil
}
