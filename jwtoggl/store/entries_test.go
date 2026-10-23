package store

import (
	"testing"
	"time"
)

func TestMarkStoppedAtMarksEntryModifiedAndComputesDuration(t *testing.T) {
	start := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	stop := start.Add(5 * time.Minute)
	entry := TimeEntry{}
	entry.Start = start

	entry.MarkStoppedAt(stop)

	if !entry.Stop.Equal(stop) {
		t.Fatalf("stop = %s, want %s", entry.Stop, stop)
	}
	if entry.Modified == nil || !entry.Modified.Equal(stop) {
		t.Fatalf("modified = %v, want %s", entry.Modified, stop)
	}
	if entry.Duration != int64((5 * time.Minute).Seconds()) {
		t.Fatalf("duration = %d, want 300", entry.Duration)
	}
}
