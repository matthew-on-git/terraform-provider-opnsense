// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLockMutex_SerializesConcurrentWrites(t *testing.T) {
	client := newConcurrencyTestClient(t, 10)

	const goroutines = 5
	var order []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a channel to start all goroutines simultaneously.
	start := make(chan struct{})

	for i := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start

			err := client.LockMutex(context.Background())
			if err != nil {
				t.Errorf("goroutine %d: LockMutex failed: %v", id, err)
				return
			}
			// Record that this goroutine is inside the critical section.
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			client.UnlockMutex()
		}(i)
	}

	close(start)
	wg.Wait()

	// All goroutines must have executed (serialized, not skipped).
	if len(order) != goroutines {
		t.Errorf("expected %d goroutines to complete, got %d", goroutines, len(order))
	}
}

func TestLockMutex_RespectsContextCancellation(t *testing.T) {
	client := newConcurrencyTestClient(t, 10)

	// Hold the mutex so the next LockMutex blocks.
	err := client.LockMutex(context.Background())
	if err != nil {
		t.Fatalf("first LockMutex failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = client.LockMutex(ctx)
	if err == nil {
		client.UnlockMutex()
		t.Fatal("expected context cancellation error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}

	// Release the first lock. The cleanup goroutine from the cancelled
	// LockMutex will acquire and immediately release, preventing deadlock.
	client.UnlockMutex()

	// Allow the cleanup goroutine to complete.
	time.Sleep(10 * time.Millisecond)

	// Verify the mutex is NOT deadlocked — a new LockMutex should succeed.
	acquireCtx, acquireCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer acquireCancel()

	err = client.LockMutex(acquireCtx)
	if err != nil {
		t.Fatalf("mutex deadlocked after cancelled LockMutex: %v", err)
	}
	client.UnlockMutex()
}

func TestAcquireRead_RespectsLimit(t *testing.T) {
	const limit = 2
	client := newConcurrencyTestClient(t, limit)

	// Acquire all available slots.
	for range limit {
		if err := client.AcquireRead(context.Background()); err != nil {
			t.Fatalf("AcquireRead failed: %v", err)
		}
	}

	// Next acquire should block — use a short timeout to detect.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := client.AcquireRead(ctx)
	if err == nil {
		client.ReleaseRead()
		t.Fatal("expected AcquireRead to block when semaphore is full")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}

	// Release all slots.
	for range limit {
		client.ReleaseRead()
	}
}

func TestReleaseRead_FreesSlot(t *testing.T) {
	const limit = 1
	client := newConcurrencyTestClient(t, limit)

	// Fill the semaphore.
	if err := client.AcquireRead(context.Background()); err != nil {
		t.Fatalf("AcquireRead failed: %v", err)
	}

	// Release and re-acquire — should succeed.
	client.ReleaseRead()

	if err := client.AcquireRead(context.Background()); err != nil {
		t.Fatalf("AcquireRead after ReleaseRead failed: %v", err)
	}
	client.ReleaseRead()
}

func TestMutexAndSemaphore_AreIndependent(t *testing.T) {
	client := newConcurrencyTestClient(t, 10)

	// Hold the write mutex.
	err := client.LockMutex(context.Background())
	if err != nil {
		t.Fatalf("LockMutex failed: %v", err)
	}

	// Reads should NOT be blocked by the mutex.
	var readsDone atomic.Int32
	var wg sync.WaitGroup

	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := client.AcquireRead(context.Background()); err != nil {
				t.Errorf("AcquireRead blocked by mutex: %v", err)
				return
			}
			readsDone.Add(1)
			client.ReleaseRead()
		}()
	}

	wg.Wait()
	client.UnlockMutex()

	if got := readsDone.Load(); got != 3 {
		t.Errorf("expected 3 reads to complete while mutex held, got %d", got)
	}
}

// newConcurrencyTestClient creates a Client with the given read concurrency for testing.
func newConcurrencyTestClient(t *testing.T, maxReads int64) *Client {
	t.Helper()
	client, err := NewClient(ClientConfig{
		BaseURL:            "http://localhost",
		APIKey:             "testkey",
		APISecret:          "testsecret", //nolint:gosec // Test credentials only
		MaxReadConcurrency: maxReads,
		RetryMax:           1,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}
