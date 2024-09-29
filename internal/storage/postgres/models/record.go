package models

import "time"

type PGRecord struct {
	ID     uint64 `gorm:"primaryKey;autoincrement;"`
	GridID int64  `gorm:"index:idx_unique_item,unique"`
	// Grid        PGGridInfo `gorm:"constraint:OnDelete:CASCADE"`
	DateTime    time.Time `gorm:"index:idx_unique_item,unique"`
	IsGround    bool
	Pressure    float32
	Temperature float32
	UWind       float32
	VWind       float32
	CRain       float32
	RHumidity   float32
	Visibility  float32
}

func (PGRecord) TableName() string {
	return "records"
}
