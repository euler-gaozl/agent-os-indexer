package indexer

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Processor is Knowledge's narrow indexing transaction protocol. The scheduler
// never receives a database connection or reconstructs transaction steps.
type Processor interface {
	BackfillNames(context.Context) (int, error)
	BackfillChunks(context.Context) (int, error)
	PendingFiles(context.Context) ([]string, error)
	ProcessFile(context.Context, string) error
}

func sweep(ctx context.Context, p Processor) error {
	var failures []error
	if _, err := p.BackfillNames(ctx); err != nil {
		failures = append(failures, errors.New("name backfill unavailable"))
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if _, err := p.BackfillChunks(ctx); err != nil {
		failures = append(failures, errors.New("chunk backfill unavailable"))
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	ids, err := p.PendingFiles(ctx)
	if err != nil {
		return errors.New("pending index lookup unavailable")
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = p.ProcessFile(ctx, id); err != nil {
			failures = append(failures, errors.New("file processing unavailable"))
		}
	}
	return errors.Join(failures...)
}

func Run(ctx context.Context, p Processor) error {
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	degraded := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			err := sweep(ctx, p)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil && !degraded {
				slog.Warn("indexing sweep incomplete; retrying on next sweep")
				degraded = true
			}
			if err == nil && degraded {
				slog.Info("indexing sweeps recovered")
				degraded = false
			}
		}
	}
}
