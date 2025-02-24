package main

import (
	"cmp"
	"log/slog"
	"slices"
	"time"
)

type AverageResponse struct {
	Value float64
	Time  time.Time
}

type ProbabilityResponse struct {
	Key   string
	Value bool
	Prob  float64
}

func (db *PsuedoDB) TrackView(user_id string, attrs map[string]bool) {
	if _, ok := db.userRecords[user_id]; !ok {
		db.userRecords[user_id] = UserRecord{
			userId: user_id,
			attrs:  attrs,
		}
	}

	db.rotateViewInterval(time.Now())
	interval := db.views[len(db.views)-1]
	interval.events++
	for attr := range attrs {
		if attrs[attr] {
			interval.typeCounts[attr]++
		}
	}
}

func (db *PsuedoDB) TrackBookDemo(user_id string) {
	record, ok := db.userRecords[user_id]
	if !ok {
		slog.Error("User not found, (didn't view prior to demoing)", "user_id", user_id)
		return
	}

	if record.gotDemo {
		slog.Info("User already demoed", "user_id", user_id)
		return
	}

	db.rotateDemoInterval(time.Now())
	interval := db.demos[len(db.demos)-1]
	interval.events++
	record.gotDemo = true
	for attr := range record.attrs {
		if record.attrs[attr] {
			interval.typeCounts[attr]++

		}
	}
}

func sumLast24hr(intervals []*Interval) int {
	// -2 to not count the in progress interval
	sum := 0
	count := 0
	for i := len(intervals) - 2; i >= 0; i-- {
		sum += (intervals)[i].events
		count++
		if count >= 24 {
			break
		}
	}
	return sum
}

func (db *PsuedoDB) GetViewsLast24hr() int {
	db.rotateViewInterval(time.Now())
	return sumLast24hr(db.views)
}

func (db *PsuedoDB) GetDemosLast24hr() int {
	db.rotateViewInterval(time.Now())
	return sumLast24hr(db.demos)
}

func (db *PsuedoDB) MovingAverageViews(duration int) *[]AverageResponse {
	db.rotateViewInterval(time.Now())
	window := 6
	averages := make([]AverageResponse, 0, duration)

	if duration+window > len(db.views) {
		slog.Warn("duration too long for available data, truncating duration to available data", "duration", duration, "window", window, "views", len(db.views))
		duration = len(db.views) - window
	}

	if duration < 1 {
		slog.Error("insufficient data for window", "window", window, "views", len(db.views))
		return nil
	}

	datapoints := make([]int, duration+window-1)
	for i := 0; i < len(datapoints); i++ {
		datapoints[i] = db.views[len(db.views)-2-i].events
	}

	windowed := getMovingAverage(datapoints, window)
	for i := 0; i < duration; i++ {
		averages = append(averages, AverageResponse{
			Value: windowed[i],
			Time:  db.views[len(db.views)-1-i].start,
		})
	}

	return &averages
}

func (db *PsuedoDB) MovingAverageViewsByQuery(duration int, key string, val bool) *[]AverageResponse {
	db.rotateViewInterval(time.Now())
	window := 6
	averages := make([]AverageResponse, 0, duration)

	if duration+window > len(db.views) {
		slog.Warn("duration too long for available data, truncating duration to available data", "duration", duration, "window", window, "views", len(db.views))
		duration = len(db.views) - window
	}

	if duration < 1 {
		slog.Error("insufficient data for window", "window", window, "views", len(db.views))
		return nil
	}

	queried_views := make([]int, duration+window-1)
	for i := 0; i < len(queried_views); i++ {
		data, ok := db.views[len(db.views)-2-i].typeCounts[key]
		if ok {
			if val {
				queried_views[i] = data
			} else {
				queried_views[i] = db.views[len(db.views)-2-i].events - data
			}
		}
	}

	windowed := getMovingAverage(queried_views, window)
	for i := 0; i < duration; i++ {
		averages = append(averages, AverageResponse{
			Value: windowed[i],
			Time:  db.views[len(db.views)-1-i].start,
		})
	}

	return &averages
}

func getMovingAverage(values []int, interval int) []float64 {
	averages := make([]float64, 0, len(values)-interval+1)
	sum := 0
	for i := 0; i < interval; i++ {
		sum += values[i]
	}
	averages = append(averages, float64(sum)/float64(interval))

	for i := interval; i < len(values); i++ {
		sum -= values[i-interval]
		sum += values[i]
		averages = append(averages, float64(sum)/float64(interval))
	}

	return averages
}

func (db *PsuedoDB) GetPredictor() ProbabilityResponse {
	now := time.Now()
	db.rotateViewInterval(now)
	db.rotateDemoInterval(now)

	viewTotals := make(map[string]int)
	viewSum := 0
	demoTotals := make(map[string]int)
	demoSum := 0

	// because views and intervals have the same rotation scheudle, we can assume they are the same length
	// should that change we need to sum them seperately
	for i := 0; i < len(db.views)-1; i++ {
		viewSum += db.views[i].events
		demoSum += db.demos[i].events

		// in theory we could use the same interator for both views and demos
		// but an interval have not views it would be a problem for the demo iterator
		for k, v := range db.views[i].typeCounts {
			viewTotals[k] += v
		}
		for k, v := range db.demos[i].typeCounts {
			demoTotals[k] += v
		}
	}

	if viewSum == 0 || demoSum == 0 {
		slog.Error("no views or demos in the last 24 hours")
		return ProbabilityResponse{}
	}

	probs := make([]ProbabilityResponse, 0, len(viewTotals)*2)
	prob_demo := float64(demoSum) / float64(viewSum)
	for k, v := range viewTotals {
		attr_prob := float64(v) / float64(viewSum)
		false_attr_prob := float64(viewSum-v) / float64(viewSum)

		prob_demo_given_attr := float64(demoTotals[k]) / float64(demoSum)
		false_demo_given_attr := float64(demoSum-demoTotals[k]) / float64(demoSum)

		if attr_prob != 0 {
			conditional_prob := (prob_demo_given_attr * prob_demo) / (attr_prob)
			probs = append(probs, ProbabilityResponse{
				Key:   k,
				Value: true,
				Prob:  conditional_prob,
			})
		} else {
			slog.Error("attr_prob is 0", "key", k)
		}

		if false_attr_prob != 0 {
			false_conditional_prob := (false_demo_given_attr * prob_demo) / (false_attr_prob)
			probs = append(probs, ProbabilityResponse{
				Key:   k,
				Value: false,
				Prob:  false_conditional_prob,
			})
		} else {
			slog.Error("false_attr_prob is 0", "key", k)
		}
	}

	slices.SortFunc(probs, func(a, b ProbabilityResponse) int {
		return cmp.Compare(b.Prob, a.Prob)
	})
	return probs[0]
}
