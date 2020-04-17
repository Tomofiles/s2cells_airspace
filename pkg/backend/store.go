package backend

import (
	"context"

	"github.com/Tomofiles/s2cells-airspace/pkg/backend/models"

	"github.com/golang/geo/s2"
)

// Store .
type Store interface {
	Close() error

	InsertArea(ctx context.Context, s models.Area, areaType int32) (*models.Area, error)

	SearchAreas(ctx context.Context, cells s2.CellUnion, areaType int32) ([]*models.Area, error)
}

// NewNilStore .
func NewNilStore() Store {
	return nil
}
