// Package qr calcula la factorización QR de una matriz densa con reflexiones de Householder.
package qr

import "math"

// Decompose devuelve la factorización QR completa A = Q·R de una matriz m×n.
// Q es m×m ortogonal y R es m×n triangular superior, para cualquier forma (m>n, m<n, m=n).
// Bajo la diagonal de R todo es exactamente 0. No modifica A.
func Decompose(a [][]float64) (q, r [][]float64) {
	m, n := len(a), len(a[0])
	r = make([][]float64, m)
	q = make([][]float64, m)
	for i := range a {
		r[i] = append([]float64(nil), a[i]...)
		q[i] = make([]float64, m)
		q[i][i] = 1
	}

	v := make([]float64, m)
	for k := 0; k < min(m-1, n); k++ {
		// Refleja x = R[k:, k] sobre alpha·e1. alpha toma el signo opuesto a x[0]
		// para que v = x - alpha·e1 no sufra cancelación.
		var norm float64
		for i := k; i < m; i++ {
			norm = math.Hypot(norm, r[i][k])
		}
		if norm == 0 {
			continue // columna ya nula bajo la diagonal
		}
		alpha := -norm
		if r[k][k] < 0 {
			alpha = norm
		}
		var vv float64
		for i := k; i < m; i++ {
			v[i] = r[i][k]
			if i == k {
				v[i] -= alpha
			}
			vv += v[i] * v[i]
		}

		// R ← H·R con H = I - 2·v·vᵀ/(vᵀv), solo filas k.. y columnas k..
		for j := k; j < n; j++ {
			var s float64
			for i := k; i < m; i++ {
				s += v[i] * r[i][j]
			}
			f := 2 * s / vv
			for i := k; i < m; i++ {
				r[i][j] -= f * v[i]
			}
		}
		// Q ← Q·H, solo columnas k..
		for i := 0; i < m; i++ {
			var s float64
			for l := k; l < m; l++ {
				s += q[i][l] * v[l]
			}
			f := 2 * s / vv
			for l := k; l < m; l++ {
				q[i][l] -= f * v[l]
			}
		}
	}

	for i := 1; i < m; i++ {
		for j := 0; j < i && j < n; j++ {
			r[i][j] = 0
		}
	}
	return q, r
}
