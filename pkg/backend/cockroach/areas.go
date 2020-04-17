package cockroach

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Tomofiles/s2cells-airspace/pkg/backend/models"

	"github.com/golang/geo/s2"
	"github.com/lib/pq"
	"go.uber.org/multierr"
)

var areaFields = "areas.area_id, areas.area_name, areas.area_type, areas.area"
var areaFieldsWithoutPrefix = "area_id, area_name, area_type, area"

func (c *Store) fetchAreas(ctx context.Context, q queryable, query string, args ...interface{}) ([]*models.Area, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload []*models.Area
	for rows.Next() {
		a := new(models.Area)

		var b []byte
		var t int32
		err := rows.Scan(
			&a.AreaID,
			&a.AreaName,
			&t,
			&b,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &a.Area)
		if err != nil {
			return nil, err
		}

		payload = append(payload, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Store) fetchArea(ctx context.Context, q queryable, query string, args ...interface{}) (*models.Area, error) {
	as, err := c.fetchAreas(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(as) > 1 {
		return nil, multierr.Combine(err, fmt.Errorf("query returned %d areas", len(as)))
	}
	if len(as) == 0 {
		return nil, sql.ErrNoRows
	}
	return as[0], nil
}

func (c *Store) pushArea(ctx context.Context, q queryable, s *models.Area, areaType int32) (*models.Area, error) {
	var (
		upsertQuery = fmt.Sprintf(`
		UPSERT INTO
			areas
		  	(%s)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			%s`, areaFieldsWithoutPrefix, areaFields)
		areaCellQuery = `
		UPSERT INTO
			cells_areas
			(cell_id, area_id)
		VALUES
			($1, $2)
		`
	)

	cids := make([]int64, len(s.Cells))

	for i, cell := range s.Cells {
		cids[i] = int64(cell)
	}

	b, err := json.Marshal(s.Area)
	if err != nil {
		return nil, err
	}

	cells := s.Cells
	s, err = c.fetchArea(ctx, q, upsertQuery,
		s.AreaID,
		s.AreaName,
		areaType,
		b)
	if err != nil {
		return nil, err
	}
	s.Cells = cells

	for i := range cids {
		if _, err := q.ExecContext(ctx, areaCellQuery, cids[i], s.AreaID); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// InsertArea .
func (c *Store) InsertArea(ctx context.Context, s models.Area, areaType int32) (*models.Area, error) {

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}

	newArea, err := c.pushArea(ctx, tx, &s, areaType)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newArea, nil
}

// SearchAreas .
func (c *Store) SearchAreas(ctx context.Context, cells s2.CellUnion, areaType int32) ([]*models.Area, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				areas
			INNER JOIN
				(SELECT DISTINCT cells_areas.area_id FROM cells_areas WHERE cells_areas.cell_id = ANY($1))
			AS
				unique_area_ids
			ON
				areas.area_id = unique_area_ids.area_id
			WHERE
				areas.area_type = $2
			`, areaFields)
	)

	if len(cells) == 0 {
		return nil, fmt.Errorf("no location provided")
	}

	tx, err := c.Begin()
	if err != nil {
		return nil, err
	}

	cids := make([]int64, len(cells))
	for i, cell := range cells {
		cids[i] = int64(cell)
	}

	areas, err := c.fetchAreas(ctx, tx, query, pq.Array(cids), areaType)
	if err != nil {
		return nil, multierr.Combine(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return areas, nil
}
