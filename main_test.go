package main

import (
	"testing"
	"time"
)

func TestNewInterval(t *testing.T) {
	// Test that newInterval returns a valid Interval
	start := time.Now()
	interval := newInterval(start)
	if interval.start != start {
		t.Errorf("Expected interval start to be %v, got %v", start, interval.start)
	}
	if interval.end != start.Add(INTERVAL_LENGTH).Add(-1*time.Nanosecond) {
		t.Errorf("Expected interval end to be %v, got %v", start.Add(INTERVAL_LENGTH).Add(-1*time.Nanosecond), interval.end)
	}
}

func TestRotateInterval(t *testing.T) {
	// Test that ensureIntervals removes old intervals
	now := time.Now()
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateInterval(now, &db.views)
	if len(db.views) != 25 {
		t.Errorf("Expected 25 intervals, got %d", len(db.views))
	}

	start := now.Add(-(MAX_INTERVALS) * INTERVAL_LENGTH)
	for i := 0; i < 25; i++ {
		t.Logf("Interval %d: %v", i, db.views[i].start)
		if db.views[i].start != start.Truncate(INTERVAL_LENGTH) {
			t.Errorf("Expected interval %d start to be %v, got %v", i, start.Truncate(INTERVAL_LENGTH), db.views[i].start)
		}
		start = start.Add(INTERVAL_LENGTH)
	}

}
