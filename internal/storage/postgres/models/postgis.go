package models

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GISRectangle struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

const (
	polygonFormatString = "POLYGON ((%f %f, %f %f, %f %f, %f %f, %f %f))"
)

func (g *GISRectangle) Scan(val interface{}) error {
	var point string
	switch v := val.(type) {
	case []byte:
		point = string(v)
	case string:
		point = v
	default:
		return fmt.Errorf("cannot convert %T to GISSquare", val)
	}

	var x1, y1 float64
	var x2, y2 float64
	var x3, y3 float64
	var x4, y4 float64
	var x5, y5 float64

	_, err := fmt.Sscanf(
		point,
		polygonFormatString,
		&x1, &y1, &x2, &y2, &x3, &y3, &x4, &y4, &x5, &y5,
	)

	if err != nil {
		return err
	}

	g.X1 = x1
	g.Y1 = y1
	g.X2 = x3
	g.Y2 = y3
	return nil
}

func (g GISRectangle) GormDataType() string {
	return "geometry"
}

func (g GISRectangle) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL: "st_polygonfromtext(?)",
		Vars: []interface{}{fmt.Sprintf(
			polygonFormatString,
			g.X1, g.Y1, g.X1, g.Y2, g.X2, g.Y2, g.X2, g.Y1, g.X1, g.Y1,
		)},
	}
}
