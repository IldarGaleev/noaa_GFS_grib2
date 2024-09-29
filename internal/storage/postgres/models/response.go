package models

import "time"

type PGResponse struct {
	Sec         string
	Temperature float64
	Pressure    float64
	CRain       float64
	RHumidity   float64
	UWind       float64
	VWind       float64
	Date        time.Time
}
