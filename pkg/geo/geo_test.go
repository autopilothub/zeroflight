package geo_test

import (
	"math"
	"testing"

	"github.com/autopilothub/zeroflight/pkg/geo"
)

func TestDistanceM(t *testing.T) {
	// Seoul city hall to nearby point (~1.1 km)
	d := geo.DistanceM(37.5665, 126.9780, 37.5765, 126.9780)
	if math.Abs(d-1112) > 50 {
		t.Fatalf("unexpected distance: %.1fm", d)
	}
}

func TestBearingDeg(t *testing.T) {
	b := geo.BearingDeg(0, 0, 0, 1)
	if math.Abs(b-90) > 0.1 {
		t.Fatalf("unexpected bearing: %.1f", b)
	}
}
