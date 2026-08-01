package geo_test

import (
	"math"
	"testing"

	"github.com/autopilothub/zeroflight/pkg/geo"
)

func TestCirclePoints(t *testing.T) {
	points := geo.CirclePoints(37.0, 127.0, 50, 4)
	if len(points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(points))
	}
	for _, p := range points {
		d := geo.DistanceM(37.0, 127.0, p[0], p[1])
		if math.Abs(d-50) > 2 {
			t.Fatalf("point distance %.1fm expected ~50m", d)
		}
	}
}
