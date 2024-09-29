package models

import "time"

type WKTRequestBody struct {
	Components []string     `json:"components,omitempty"`
	Shapes     []WKTRequest `json:"shapes"`
}

type WKTRequest struct {
	WKT  string     `json:"wkt"`
	From time.Time  `json:"from"`
	To   *time.Time `json:"to,omitempty"`
}
