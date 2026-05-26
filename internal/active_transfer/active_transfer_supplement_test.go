package active_transfer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_EmitPersist_NilPersist(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)

	m.emitPersist(1, ActiveTransferSnapshot{})
}

func TestManager_EmitPersist_WithPersistFunc(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)

	persisted := false
	var persistedRunID int64

	m.SetPersistFunc(func(runID int64, snap ActiveTransferSnapshot) {
		persisted = true
		persistedRunID = runID
	})

	snap := ActiveTransferSnapshot{
		RunID:      123,
		TaskID:     456,
		TotalCount: 100,
	}

	m.emitPersist(123, snap)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, persisted)
	assert.Equal(t, int64(123), persistedRunID)
}

func TestManager_FlushPendingPersist_NotExist(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)

	assert.NotPanics(t, func() {
		m.flushPendingPersist(999)
	})
}

func TestManager_FlushPendingPersist_WithPending(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)

	persisted := false
	var persistedRunID int64

	m.SetPersistFunc(func(runID int64, snap ActiveTransferSnapshot) {
		persisted = true
		persistedRunID = runID
	})

	m.mu.Lock()
	m.pendingPersist[123] = ActiveTransferSnapshot{RunID: 123, TotalCount: 50}
	m.mu.Unlock()

	m.flushPendingPersist(123)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, persisted)
	assert.Equal(t, int64(123), persistedRunID)

	m.mu.Lock()
	_, exists := m.pendingPersist[123]
	m.mu.Unlock()
	assert.False(t, exists)
}

func TestManager_SetPersistFunc_Custom(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)

	called := false
	testFunc := func(runID int64, snap ActiveTransferSnapshot) {
		called = true
	}

	m.SetPersistFunc(testFunc)

	m.mu.RLock()
	persist := m.persist
	m.mu.RUnlock()

	assert.NotNil(t, persist)

	persist(1, ActiveTransferSnapshot{})
	assert.True(t, called)
}