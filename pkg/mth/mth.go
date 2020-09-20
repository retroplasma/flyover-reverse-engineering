package mth

import "math"

// QuaternionToMatrix converts quaternion values to 3x3 matrix
func QuaternionToMatrix(qx, qy, qz, qw float64) (m [9]float64) {
	m[0] = 1 - 2*qy*qy - 2*qz*qz
	m[1] = 2*qx*qy - 2*qw*qz
	m[2] = 2*qx*qz + 2*qw*qy
	m[3] = 2*qx*qy + 2*qw*qz
	m[4] = 1 - 2*qx*qx - 2*qz*qz
	m[5] = 2*qy*qz - 2*qw*qx
	m[6] = 2*qx*qz - 2*qw*qy
	m[7] = 2*qy*qz + 2*qw*qx
	m[8] = 1 - 2*qx*qx - 2*qy*qy
	return
}

// LatLonToTileOSM converts coordinates to OSM tile numbers for a zoom level
func LatLonToTileOSM(zoom int, lat, lon float64) (x, y int) {
	n := float64(pow2(zoom))
	x = int(n * ((lon + 180) / 360))
	latRad := lat / 180 * math.Pi
	y = int(n * (1 - (math.Log(math.Tan(latRad)+1/math.Cos(latRad)) / math.Pi)) / 2)
	return
}

// LatLonToTileTMS converts coordinates to TMS tile numbers for a zoom level
func LatLonToTileTMS(zoom int, lat, lon float64) (x, y int) {
	n := float64(pow2(zoom))
	x = int(n * ((lon + 180) / 360))
	latRad := lat / 180 * math.Pi
	y = int((math.Log(math.Tan(latRad*0.5+math.Pi/4))*1/(2*math.Pi) + 0.5) * n)
	return
}

// TileCountPerAxis gives the number of tiles per axis X or Y for zoom level
func TileCountPerAxis(zoom int) int {
	return pow2(zoom)
}

func pow2(y int) int {
	return 1 << y
}
