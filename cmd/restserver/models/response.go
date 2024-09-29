package models

import "time"

type WindInfo struct {
	U float64 `json:"u"`
	V float64 `json:"v"`
}

type ForecastDetail struct {
	DateTime    time.Time `json:"date-time"`
	Temperature *float64  `json:"temperature-2m,omitempty"`
	Pressure    *float64  `json:"pressure-surface,omitempty"`
	RHumidity   *float64  `json:"rhumidity-surface,omitempty"`
	CRain       *float64  `json:"crain-surface,omitempty"`
	Wind        *WindInfo `json:"wind-10m,omitempty"`
}

type ForecastResponse struct {
	Shape    string           `json:"shape"`
	Forecast []ForecastDetail `json:"forecast"`
}
