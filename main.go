package main

import (
	"time"
)

const (
	INTERVAL_LENGTH = 1 * time.Hour
	MAX_INTERVALS   = 24
)

type PsuedoDB struct {
	userRecords map[string]UserRecord
	views       []*Interval
	demos       []*Interval
}

type UserRecord struct {
	userId  string
	attrs   map[string]bool
	gotDemo bool
}

type Interval struct {
	events     int
	typeCounts map[string]int
	start      time.Time
	end        time.Time
}

func main() {
	// for example purposes:
	//
	//	db := &PsuedoDB{
	//		userRecords: make(map[string]UserRecord),
	// +1 for interval that is in progress
	//		views: make([]*Interval, 0, MAX_INTERVALS+1),
	//		demos: make([]*Interval, 0, MAX_INTERVALS+1),
	//	}
}

func newInterval(start time.Time) *Interval {
	return &Interval{
		events:     0,
		typeCounts: make(map[string]int),
		start:      start,
		end:        start.Add(INTERVAL_LENGTH).Add(-1 * time.Nanosecond),
	}
}

func (db *PsuedoDB) rotateInterval(now time.Time, intervals *[]*Interval) {
	// remove old intervals
	cutoff := now.Add(-(MAX_INTERVALS + 1) * INTERVAL_LENGTH)
	for len(*intervals) > 0 && (*intervals)[0].end.Before(cutoff) {
		*intervals = (*intervals)[1:]
	}

	// add new intervals
	var start time.Time
	if len(*intervals) == 0 {
		start = now.Add(-(MAX_INTERVALS) * INTERVAL_LENGTH).Truncate(INTERVAL_LENGTH)
	} else {
		start = (*intervals)[len(*intervals)-1].start
	}

	// +1 for interval that is in progress
	for i := 0; i < MAX_INTERVALS+1; i++ {
		if len(*intervals) <= i {
			*intervals = append(*intervals, newInterval(start))
		}
		start = start.Add(INTERVAL_LENGTH)
	}
}

func (db *PsuedoDB) rotateViewInterval(now time.Time) {
	db.rotateInterval(now, &db.views)
}

func (db *PsuedoDB) rotateDemoInterval(now time.Time) {
	db.rotateInterval(now, &db.demos)
}
