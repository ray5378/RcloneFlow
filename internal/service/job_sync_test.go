package service

import (
	"context"
	"testing"
)

func TestJobSyncServiceNoopLifecycle(t *testing.T) {
	svc := NewJobSyncService(nil, nil, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc.Start(ctx)
	svc.Stop()
}
