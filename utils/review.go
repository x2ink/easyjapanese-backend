package utils

import (
	"math"
	"time"
)

type ReviewResult struct {
	Easiness       float64   `json:"easiness"`
	Interval       int       `json:"interval"`
	Repetitions    int       `json:"repetitions"`
	ReviewDateTime time.Time `json:"review_datetime"`
}

func Review(quality int, easiness float64, interval int, repetitions int, reviewDateTime ...time.Time) ReviewResult {
	var now time.Time
	if len(reviewDateTime) > 0 {
		now = reviewDateTime[0]
	} else {
		now = time.Now().UTC()
	}

	if quality < 3 {
		interval = 1
		repetitions = 0
	} else {
		switch repetitions {
		case 0:
			interval = 1
		case 1:
			interval = 6
		default:
			interval = int(math.Ceil(float64(interval) * easiness))
		}
		repetitions++
	}

	easiness += 0.1 - (5.0-float64(quality))*(0.08+(5.0-float64(quality))*0.02)
	if easiness < 1.3 {
		easiness = 1.3
	}

	nextReviewTime := now.Add(time.Duration(interval) * 24 * time.Hour)

	return ReviewResult{
		Easiness:       easiness,
		Interval:       interval,
		Repetitions:    repetitions,
		ReviewDateTime: nextReviewTime,
	}
}
func FirstReview(quality int, reviewDateTime ...time.Time) ReviewResult {
	var now time.Time
	if len(reviewDateTime) > 0 {
		now = reviewDateTime[0]
	} else {
		now = time.Now().UTC()
	}
	return Review(quality, 2.5, 0, 0, now)
}
