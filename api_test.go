package main

import (
	"testing"
	"time"
)

func TestTrackView(t *testing.T) {
	// Test that track_view adds a new user record
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	db.TrackView("user1", attrs)
	if len(db.userRecords) != 1 {
		t.Errorf("Expected 1 user record, got %d", len(db.userRecords))
	}
	if len(db.views) != 25 {
		t.Errorf("Expected 1 view interval, got %d", len(db.views))
	}

	lastEle := db.views[len(db.views)-1]
	if lastEle.events != 1 {
		t.Errorf("Expected 1 event, got %d", lastEle.events)
	}
	if lastEle.typeCounts["attr1"] != 1 {
		t.Errorf("Expected 1 attr1, got %d", lastEle.typeCounts["attr1"])
	}
	if lastEle.typeCounts["attr2"] != 1 {
		t.Errorf("Expected 1 attr2, got %d", lastEle.typeCounts["attr2"])
	}
	if _, ok := lastEle.typeCounts["attr3"]; ok {
		t.Errorf("Expected to not find attr3, found it, counts: %+v", lastEle.typeCounts)
	}
}

func TestTrackBookDemo(t *testing.T) {
	// Test that track_book_demo adds a new user record
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	db.TrackBookDemo("user1")
	if len(db.demos) != 0 {
		t.Errorf("Expected 0 demo interval, got %d", len(db.demos))
	}

	db.TrackView("user1", attrs)
	db.TrackBookDemo("user1")
	if len(db.demos) != 25 {
		t.Errorf("Expected 1 demo interval, got %d", len(db.demos))
	}

	lastEle := db.demos[len(db.demos)-1]
	if lastEle.events != 1 {
		t.Errorf("Expected 1 event, got %d", lastEle.events)
	}
	if lastEle.typeCounts["attr1"] != 1 {
		t.Errorf("Expected 1 attr1, got %d", lastEle.typeCounts["attr1"])
	}
	if lastEle.typeCounts["attr2"] != 1 {
		t.Errorf("Expected 1 attr2, got %d", lastEle.typeCounts["attr2"])
	}
	if _, ok := lastEle.typeCounts["attr3"]; ok {
		t.Errorf("Expected to not find attr3, found it, counts: %+v", lastEle.typeCounts)
	}
}

func TestGetViewsLast24hr(t *testing.T) {
	// Test that get_views_last_24hr returns the correct sum
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateViewInterval(time.Now())
	db.rotateDemoInterval(time.Now())

	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	for i := 0; i < len(db.views); i++ {
		interval := db.views[i]
		interval.events = 2
		for attr := range attrs {
			if attrs[attr] {
				interval.typeCounts[attr] = i
			}
		}
	}

	sum := db.GetViewsLast24hr()
	if sum != 48 {
		t.Errorf("Expected 48 views, got %d", sum)
	}
}

func TestGetDemosLast24hr(t *testing.T) {
	// Test that get_demos_last_24hr returns the correct sum
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateViewInterval(time.Now())
	db.rotateDemoInterval(time.Now())

	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	for i := 0; i < len(db.demos); i++ {
		interval := db.demos[i]
		interval.events = 5
		for attr := range attrs {
			if attrs[attr] {
				interval.typeCounts[attr] = i
			}
		}
	}

	sum := db.GetDemosLast24hr()
	if sum != 120 {
		t.Errorf("Expected 4 demos, got %d", sum)
	}
}

func TestMovingAverageViews(t *testing.T) {
	// Test that moving_average_views returns the correct average
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateViewInterval(time.Now())
	db.rotateDemoInterval(time.Now())

	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	for i := 0; i < len(db.views); i++ {
		interval := db.views[i]
		interval.events = 2
		for attr := range attrs {
			if attrs[attr] {
				interval.typeCounts[attr] = i
			}
		}
	}

	avg := db.MovingAverageViews(3)
	if len(*avg) != 3 {
		t.Errorf("Expected 3 averages, got %d", len(*avg))
	}
	now := time.Now()
	now = now.Truncate(INTERVAL_LENGTH)
	for i := 0; i < 3; i++ {
		if (*avg)[i].Value != 2 {
			t.Errorf("Expected average to be 2, got %f", (*avg)[i].Value)
		}
		if (*avg)[i].Time != now.Add(-time.Duration(i)*INTERVAL_LENGTH) {
			t.Errorf("Expected start to be %v, got %v", now.Add(-time.Duration(i)*INTERVAL_LENGTH), (*avg)[i].Time)
		}
	}
}

func TestMovingAverageViewsByQuery(t *testing.T) {
	// Test that moving_average_views returns the correct average
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateViewInterval(time.Now())
	db.rotateDemoInterval(time.Now())

	attrs := map[string]bool{
		"attr1": true,
		"attr2": true,
		"attr3": false,
	}
	for i := 0; i < len(db.views); i++ {
		interval := db.views[i]
		interval.events = 2
		for attr := range attrs {
			if attrs[attr] {
				interval.typeCounts[attr] = i
			}
		}
	}

	avg := db.MovingAverageViewsByQuery(3, "attr1", true)
	if len(*avg) != 3 {
		t.Errorf("Expected 3 averages, got %d", len(*avg))
	}
	now := time.Now()
	now = now.Truncate(INTERVAL_LENGTH)
	for i := 0; i < 3; i++ {
		if (*avg)[i].Value != 20.5-float64(i) {
			t.Errorf("Expected average to be 20.5 - %d, got %f", i, (*avg)[i].Value)
		}
		if (*avg)[i].Time != now.Add(-time.Duration(i)*INTERVAL_LENGTH) {
			t.Errorf("Expected start to be %v, got %v", now.Add(-time.Duration(i)*INTERVAL_LENGTH), (*avg)[i].Time)
		}
	}
}

func TestGetPredictor(t *testing.T) {
	// Test that get_probability returns the correct probability
	db := &PsuedoDB{
		userRecords: make(map[string]UserRecord),
		views:       make([]*Interval, 0, MAX_INTERVALS+1),
		demos:       make([]*Interval, 0, MAX_INTERVALS+1),
	}
	db.rotateViewInterval(time.Now())
	db.rotateDemoInterval(time.Now())

	viewEle := db.views[len(db.views)-2]
	viewEle.events = 100
	viewEle.typeCounts["attr1"] = 50
	viewEle.typeCounts["attr2"] = 33
	viewEle.typeCounts["attr3"] = 20

	demoEle := db.demos[len(db.demos)-2]
	demoEle.events = 50
	demoEle.typeCounts["attr1"] = 40
	demoEle.typeCounts["attr2"] = 20
	demoEle.typeCounts["attr3"] = 10

	prob := db.GetPredictor()
	t.Logf("Prob: %+v", prob)

	if prob.Key != "attr1" {
		t.Errorf("Expected key to be attr1, got %s", prob.Key)
	}
	if prob.Value != true {
		t.Errorf("Expected value to be true, got %t", prob.Value)
	}
	if prob.Prob != 0.8 {
		t.Errorf("Expected prob to be 0.8, got %f", prob.Prob)
	}
}
