package geo

import "math"

// CirclePoints generates lat/lon points around a center at equal angular spacing.
func CirclePoints(centerLat, centerLon float64, radiusM float64, points int) [][2]float64 {
	if points < 3 {
		points = 3
	}
	out := make([][2]float64, points)
	for i := 0; i < points; i++ {
		bearing := float64(i) * 360 / float64(points)
		lat, lon := destinationPoint(centerLat, centerLon, radiusM, bearing)
		out[i] = [2]float64{lat, lon}
	}
	return out
}

func destinationPoint(lat, lon, distanceM, bearingDeg float64) (float64, float64) {
	rad := math.Pi / 180
	lat1 := lat * rad
	lon1 := lon * rad
	brng := bearingDeg * rad
	angDist := distanceM / earthRadiusM

	lat2 := math.Asin(math.Sin(lat1)*math.Cos(angDist) +
		math.Cos(lat1)*math.Sin(angDist)*math.Cos(brng))
	lon2 := lon1 + math.Atan2(
		math.Sin(brng)*math.Sin(angDist)*math.Cos(lat1),
		math.Cos(angDist)-math.Sin(lat1)*math.Sin(lat2),
	)
	return lat2 / rad, lon2 / rad
}
