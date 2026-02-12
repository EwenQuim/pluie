package main

import (
	"sync"
	"testing"
	"time"
)

func TestEmbeddingProgressGetStatus(t *testing.T) {
	ep := NewEmbeddingProgress()
	ep.UpdateProgress(5, 10, "test_note", true)

	status := ep.GetStatus()
	if status.EmbeddedNotes != 5 {
		t.Errorf("EmbeddedNotes = %d, want 5", status.EmbeddedNotes)
	}
	if status.TotalNotes != 10 {
		t.Errorf("TotalNotes = %d, want 10", status.TotalNotes)
	}
	if status.CurrentNote != "test_note" {
		t.Errorf("CurrentNote = %q, want %q", status.CurrentNote, "test_note")
	}
	if !status.IsEmbedding {
		t.Error("IsEmbedding = false, want true")
	}
}

func TestEmbeddingProgressSubscribeReceivesUpdates(t *testing.T) {
	ep := NewEmbeddingProgress()
	ch := ep.Subscribe()

	ep.UpdateProgress(1, 10, "note1", true)

	select {
	case status := <-ch:
		if status.EmbeddedNotes != 1 {
			t.Errorf("EmbeddedNotes = %d, want 1", status.EmbeddedNotes)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for update")
	}
}

func TestEmbeddingProgressUnsubscribeStopsUpdates(t *testing.T) {
	ep := NewEmbeddingProgress()
	ch := ep.Subscribe()
	ep.Unsubscribe(ch)

	ep.UpdateProgress(1, 10, "note1", true)

	// Channel should not receive updates after unsubscribe
	select {
	case <-ch:
		t.Fatal("Received update after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// Expected - no update received
	}
}

func TestEmbeddingProgressConcurrentAccess(t *testing.T) {
	ep := NewEmbeddingProgress()

	var wg sync.WaitGroup
	const goroutines = 20

	// Concurrent subscribers
	channels := make([]chan EmbeddingStatus, goroutines/2)
	for i := range goroutines / 2 {
		ch := ep.Subscribe()
		channels[i] = ch
	}

	// Concurrent updates
	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ep.UpdateProgress(i, goroutines, "note", true)
		}()
	}

	// Concurrent reads
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ep.GetStatus()
		}()
	}

	// Concurrent unsubscribes
	for _, ch := range channels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ep.Unsubscribe(ch)
		}()
	}

	wg.Wait()
}

func TestEmbeddingProgressMultipleSubscribers(t *testing.T) {
	ep := NewEmbeddingProgress()
	ch1 := ep.Subscribe()
	ch2 := ep.Subscribe()

	ep.UpdateProgress(3, 5, "note3", true)

	// Both subscribers should receive the update
	for i, ch := range []chan EmbeddingStatus{ch1, ch2} {
		select {
		case status := <-ch:
			if status.EmbeddedNotes != 3 {
				t.Errorf("subscriber %d: EmbeddedNotes = %d, want 3", i, status.EmbeddedNotes)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out waiting for update", i)
		}
	}
}
