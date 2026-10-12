package main

import (
	"context"
	"github.com/euler-gaozl/agent-os-indexer/indexer"
	"github.com/euler-gaozl/agent-os-knowledge/indexing"
	"github.com/euler-gaozl/agent-os-sdk/servicehost"
	"github.com/euler-gaozl/agent-os-sdk/servicehttp"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func main() {
	servicehost.Main(servicehost.Options{Role: "indexer", Port: "8097", Database: true, Concurrency: 4, Authorize: indexer.Authorize, Readiness: indexing.Readiness}, func(ctx context.Context, p *pgxpool.Pool, c servicehttp.Config) (http.Handler, func(context.Context) error, error) {
		if err := applyLocalDefaults(); err != nil {
			return nil, nil, err
		}
		processor := indexing.New(p, servicehost.Client())
		return servicehttp.New(c, p.Ping, func(*http.ServeMux) {}), func(ctx context.Context) error { return indexer.Run(ctx, processor) }, nil
	})
}
