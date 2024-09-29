package models

import "time"

type WKTRequestItem struct {
	From time.Time
	To   *time.Time
	WKT  string
}
