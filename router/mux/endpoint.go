package mux

import (
	"net/http"

	opencensus "github.com/philippheuer/krakend-opencensus"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router/mux"
	"go.opencensus.io/plugin/ochttp"
)

func New(hf mux.HandlerFactory) mux.HandlerFactory {
	if !opencensus.IsRouterEnabled() {
		return hf
	}
	return func(cfg *config.EndpointConfig, p proxy.Proxy) http.HandlerFunc {
		h := ochttp.Handler{Handler: metricsReportingMiddleware(hf(cfg, p), cfg)}
		return h.ServeHTTP
	}
}

func metricsReportingMiddleware(next http.Handler, cfg *config.EndpointConfig) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ochttp.SetRoute(ctx, opencensus.GetStatisticsPathForEndpoint(cfg, r))
		next.ServeHTTP(w, r)
    })
}
