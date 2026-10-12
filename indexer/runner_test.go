package indexer

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fixtureProcessor struct {
	calls  []string
	cancel context.CancelFunc
	fail   bool
}

func (p *fixtureProcessor) BackfillNames(context.Context) (int, error) {
	p.calls = append(p.calls, "names")
	if p.fail {
		return 0, errors.New("fixture transient database failure")
	}
	return 0, nil
}
func (p *fixtureProcessor) BackfillChunks(context.Context) (int, error) {
	p.calls = append(p.calls, "chunks")
	return 0, nil
}
func (p *fixtureProcessor) PendingFiles(context.Context) ([]string, error) {
	p.calls = append(p.calls, "pending")
	return []string{"a", "b"}, nil
}
func (p *fixtureProcessor) ProcessFile(_ context.Context, id string) error {
	p.calls = append(p.calls, id)
	if p.cancel != nil {
		p.cancel()
	}
	if p.fail {
		return errors.New("fixture unavailable provider")
	}
	return nil
}

func TestBackfillFailureDoesNotStarveIndependentPendingFile(t *testing.T) {
	p := &fixtureProcessor{fail: true}
	if err := sweep(context.Background(), p); err == nil {
		t.Fatal("incomplete sweep hidden")
	}
	if !reflect.DeepEqual(p.calls, []string{"names", "chunks", "pending", "a", "b"}) {
		t.Fatal("pending files starved", p.calls)
	}
	p.calls = nil
	p.fail = false
	if err := sweep(context.Background(), p); err != nil {
		t.Fatal("next sweep did not recover", err)
	}
}
func TestCancellationStopsFurtherFileProcessingAndIdleWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &fixtureProcessor{cancel: cancel}
	if err := sweep(ctx, p); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation not returned", err)
	}
	if !reflect.DeepEqual(p.calls, []string{"names", "chunks", "pending", "a"}) {
		t.Fatal("work continued after cancellation", p.calls)
	}
	done := make(chan error, 1)
	go func() { done <- Run(ctx, p) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel waited for next tick")
	}
}
