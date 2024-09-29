package models

import "time"

type Record struct {
	DateTime    time.Time
	Lat         float32
	Lng         float32
	IsGround    bool
	Pressure    float32
	Temperature float32
	UWind       float32
	VWind       float32
	CRain       float32
	RHUmidity   float32
	Visibility  float32
}
