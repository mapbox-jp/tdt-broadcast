package model

import "time"

type Location struct {
	Longitude float64
	Latitude  float64
	Timestamp time.Time
}
