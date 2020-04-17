package models

import "github.com/golang/geo/s2"

// Area .
type Area struct {
	AreaID   string
	AreaName string
	Area     [][][]float64
	Cells    s2.CellUnion
}
