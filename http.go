package opencensus

import (
	"context"
	"net/http"
	"fmt"

	"github.com/devopsfaith/krakend/config"
	transport "github.com/devopsfaith/krakend/transport/http/client"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"go.opencensus.io/tag"
)

var defaultClient = &http.Client{Transport: &ochttp.Transport{}}

type krakendStatsTransport struct {
	Base http.RoundTripper
	Cfg *config.Backend
}

func NewHTTPClient(ctx context.Context) *http.Client {
	if !IsBackendEnabled() {
		return transport.NewHTTPClient(ctx)
	}
	return defaultClient
}

func HTTPRequestExecutor(clientFactory transport.HTTPClientFactory, cfg *config.Backend) transport.HTTPRequestExecutor {
	if !IsBackendEnabled() {
		return transport.DefaultHTTPRequestExecutor(clientFactory)
	}
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		// ctx, _ = tag.New(ctx, tag.Upsert(ochttp.KeyClientPath, GetAggregatedPathForBackendMetrics(cfg, req)))
		fmt.Printf("%v\n", ctx)
		client := clientFactory(ctx)
		client.Transport = &ochttp.Transport{ Base: krakendStatsTransport{ Base: client.Transport, Cfg: cfg } }

		fmt.Printf("%v\n", ctx)
		ctx = trace.NewContext(ctx, fromContext(ctx))
		fmt.Printf("%v\n", ctx)
		return client.Do(req.WithContext(ctx))
	}
}

// RoundTrip implements http.RoundTripper, delegating to Base and recording stats for the request.
func (t krakendStatsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, _ := tag.New(req.Context(),
		tag.Upsert(ochttp.KeyClientPath, "OVERWRITE TEST"))
	req = req.WithContext(ctx)
	
	// Perform request.
	resp, err := t.base().RoundTrip(req)

	return resp, err
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t krakendStatsTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (t krakendStatsTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}