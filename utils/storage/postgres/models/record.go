package models

import "time"

type PGRecord struct {
	ID   uint64 `gorm:"primaryKey;autoincrement;index:idx_record_item"`
	Geo  int64 `gorm:"index:idx_unique_item"`
	DateTime time.Time `gorm:"index:idx_unique_item"`
	IsGround bool
	Pressure float32
	Temperature float32
	UWind float32
	VWind float32
}

func (PGRecord) TableName() string {
	return "records"
}