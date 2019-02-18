package mth

// QuaternionToMatrix converts quaternion values to 3x3 matrix
func QuaternionToMatrix(qx, qy, qz, qw float64) (m [10]float64) {
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
