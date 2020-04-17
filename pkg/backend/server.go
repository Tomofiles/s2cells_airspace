package backend

import (
	"context"
	"encoding/csv"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tomofiles/s2cells-airspace/pkg/backend/models"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/labstack/echo"
	geojson "github.com/paulmach/go.geojson"
)

const (
	// DefaultMinimumCellLevel .
	DefaultMinimumCellLevel int = 13
	// DefaultMaximumCellLevel .
	DefaultMaximumCellLevel int = 13
	maxAllowedAreaKm2Bounds     = 100000.0
	maxAllowedAreaKm2           = 2500.0
)

const (
	areaTypeDid     = 0
	areaTypeAirport = 1
)

const (
	earthAreaKm2   = 510072000.0
	earthRadiusKm  = 6371.01
	airportRadius  = 9.0
	heliportRadius = 3.0
)

var (
	defaultRegionCoverer = &s2.RegionCoverer{
		MinLevel: DefaultMinimumCellLevel,
		MaxLevel: DefaultMaximumCellLevel,
	}
	// RegionCoverer .
	RegionCoverer = defaultRegionCoverer
)

func loopAreaKm2(loop *s2.Loop) float64 {
	if loop.IsEmpty() {
		return 0
	}
	return (loop.Area() * earthAreaKm2) / 4.0 * math.Pi
}

func kmToAngle(km float64) s1.Angle {
	return s1.Angle(km / earthRadiusKm)
}

// Server .
type Server struct {
	Store Store
}

// CreateDidAreas .
func (s *Server) CreateDidAreas(c echo.Context) error {
	ctx := context.Background()

	requestBody, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	// featurecollectionパース
	fc, err := geojson.UnmarshalFeatureCollection(requestBody)
	if err != nil {
		return err
	}

	for _, f := range fc.Features {

		// polygonパース
		g := geojson.NewPolygonGeometry(f.Geometry.Polygon)

		// s2変換
		points := make([]s2.Point, 0)
		for _, v := range g.Polygon[0] {
			point := s2.PointFromLatLng(s2.LatLngFromDegrees(v[1], v[0]))
			points = append(points, point)
		}
		loop := s2.LoopFromPoints(points)

		// 反時計回りに整列
		if loopAreaKm2(loop) > maxAllowedAreaKm2 {
			for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
				points[i], points[j] = points[j], points[i]
			}
		}
		loop = s2.LoopFromPoints(points)
		if loopAreaKm2(loop) > maxAllowedAreaKm2 {
			return echo.ErrInternalServerError
		}

		// CellID取得
		cells := RegionCoverer.Covering(loop)

		areaID, _ := f.PropertyInt("DIDid")
		areaName, _ := f.PropertyString("市町村名称")
		area := models.Area{
			AreaID:   strconv.Itoa(areaID),
			AreaName: areaName,
			Area:     f.Geometry.Polygon,
			Cells:    cells,
		}

		_, err = s.Store.InsertArea(ctx, area, areaTypeDid)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateAirportAreas .
func (s *Server) CreateAirportAreas(c echo.Context) error {
	ctx := context.Background()

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	reader := csv.NewReader(src)

	for {
		line, err := reader.Read()
		if err != nil {
			break
		}
		if line[8] == "JP" && line[2] != "closed" {
			lng, _ := strconv.ParseFloat(line[5], 64)
			lat, _ := strconv.ParseFloat(line[4], 64)

			radius := airportRadius
			if line[2] == "heliport" {
				radius = heliportRadius
			}

			loop := s2.RegularLoop(s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)), s1.Angle(kmToAngle(radius)), 32)
			if loopAreaKm2(loop) > maxAllowedAreaKm2 {
				return echo.ErrInternalServerError
			}

			// CellID取得
			cells := RegionCoverer.Covering(loop)

			multiPolygon := make([][][]float64, 0)
			polygon := make([][]float64, 0)

			for _, v := range loop.Vertices() {
				latlng := s2.LatLngFromPoint(v)
				point := make([]float64, 0)
				point = append(point, latlng.Lng.Degrees())
				point = append(point, latlng.Lat.Degrees())
				polygon = append(polygon, point)
			}
			point := make([]float64, 0)
			point = append(point, polygon[0][0])
			point = append(point, polygon[0][1])
			polygon = append(polygon, point)

			multiPolygon = append(multiPolygon, polygon)

			areaID := line[1]
			areaName := line[3]
			area := models.Area{
				AreaID:   areaID,
				AreaName: areaName,
				Area:     multiPolygon,
				Cells:    cells,
			}

			_, err = s.Store.InsertArea(ctx, area, areaTypeAirport)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetDidAreas .
func (s *Server) GetDidAreas(c echo.Context) error {
	ctx := context.Background()

	bounds := c.QueryParam("bounds")

	latlon := make([]float64, 0)
	for _, v := range strings.Split(bounds, ",") {
		i, _ := strconv.ParseFloat(v, 64)
		latlon = append(latlon, i)
	}

	points := make([]s2.Point, 0)
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[0], latlon[1])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[2], latlon[3])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[4], latlon[5])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[6], latlon[7])))
	loop := s2.LoopFromPoints(points)

	// 反時計回りに整列
	if loopAreaKm2(loop) > maxAllowedAreaKm2Bounds {
		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}
	}
	loop = s2.LoopFromPoints(points)

	featureCollection := geojson.NewFeatureCollection()

	if loopAreaKm2(loop) <= maxAllowedAreaKm2Bounds {
		cells := RegionCoverer.Covering(loop)

		areas, err := s.Store.SearchAreas(ctx, cells, areaTypeDid)
		if err != nil {
			return err
		}

		// Areasをgeojsonに変換
		for _, v := range areas {
			multiPolygon := v.Area

			properties := make(map[string]interface{})
			properties["area_id"] = v.AreaID
			properties["area_name"] = v.AreaName

			polygonGeometry := geojson.NewPolygonGeometry(multiPolygon)
			feature := geojson.NewFeature(polygonGeometry)
			feature.Properties = properties

			featureCollection.Features = append(featureCollection.Features, feature)
		}
	}

	// jsonパース
	b, err := featureCollection.MarshalJSON()
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "application/json", b)
}

// GetAirportAreas .
func (s *Server) GetAirportAreas(c echo.Context) error {
	ctx := context.Background()

	bounds := c.QueryParam("bounds")

	latlon := make([]float64, 0)
	for _, v := range strings.Split(bounds, ",") {
		i, _ := strconv.ParseFloat(v, 64)
		latlon = append(latlon, i)
	}

	points := make([]s2.Point, 0)
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[0], latlon[1])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[2], latlon[3])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[4], latlon[5])))
	points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(latlon[6], latlon[7])))
	loop := s2.LoopFromPoints(points)

	// 反時計回りに整列
	if loopAreaKm2(loop) > maxAllowedAreaKm2Bounds {
		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}
	}
	loop = s2.LoopFromPoints(points)

	featureCollection := geojson.NewFeatureCollection()

	if loopAreaKm2(loop) <= maxAllowedAreaKm2Bounds {
		cells := RegionCoverer.Covering(loop)

		areas, err := s.Store.SearchAreas(ctx, cells, areaTypeAirport)
		if err != nil {
			return err
		}

		// Areasをgeojsonに変換
		for _, v := range areas {
			multiPolygon := v.Area

			properties := make(map[string]interface{})
			properties["area_id"] = v.AreaID
			properties["area_name"] = v.AreaName

			polygonGeometry := geojson.NewPolygonGeometry(multiPolygon)
			feature := geojson.NewFeature(polygonGeometry)
			feature.Properties = properties

			featureCollection.Features = append(featureCollection.Features, feature)
		}
	}

	// jsonパース
	b, err := featureCollection.MarshalJSON()
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "application/json", b)
}
