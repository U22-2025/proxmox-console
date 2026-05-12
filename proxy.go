package main
import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func kratosReverseProxy() http.Handler {
	target, _ := url.Parse("http://127.0.0.1:3000")

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		// /kratos を取り除いて Kratos に渡す
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/kratos")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		r.Host = target.Host
		r.Header.Set("X-Forwarded-Host", r.Host)
		r.Header.Set("X-Forwarded-Proto", "http")
	}

	return proxy
}