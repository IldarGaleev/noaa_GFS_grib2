package models

type PGGridInfo struct {
	ID       int64        `gorm:"type:int8;index:idx_id,unique"`
	Geometry GISRectangle `gorm:"primaryKey;type:geometry(POLYGON, 4326)"`
	// Records  []PGRecord   `gorm:"constraint:OnDelete:CASCADE;foreignKey:GridID"`
}

func (PGGridInfo) TableName() string {
	return "grid"
}
