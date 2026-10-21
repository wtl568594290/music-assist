package tools

import (
	"net/http"
	"time"

	"github.com/mattn/go-ieproxy"
)

var HttpClient *http.Client

func init() {
	HttpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:               ieproxy.GetProxyFunc(),
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 8,
			MaxConnsPerHost:     8,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}
