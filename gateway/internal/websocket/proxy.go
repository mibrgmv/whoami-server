package websocket

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	reverse *httputil.ReverseProxy
	logger  *slog.Logger
}

func NewProxy(targetAddr string, logger *slog.Logger) (*Proxy, error) {
	targetURL, err := url.Parse("http://" + targetAddr)
	if err != nil {
		return nil, err
	}

	reverse := httputil.NewSingleHostReverseProxy(targetURL)
	reverse.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		req.Host = targetURL.Host
	}

	return &Proxy{
		reverse: reverse,
		logger:  logger,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.logger.Info("proxying WebSocket",
		slog.String("path", r.URL.Path),
		slog.String("target", r.URL.Path),
	)
	p.reverse.ServeHTTP(w, r)
}
