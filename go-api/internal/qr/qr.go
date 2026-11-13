// Package qr computes the QR factorization of a dense matrix with Householder reflections.
package qr

import "math"

// Decompose returns the full QR factorization A = Q·R of an m×n matrix (rows of equal length).
// Q is m×m orthogonal and R is m×n upper triangular, for any shape (m>n, m<n, m=n).
// Entries below R's diagonal are exactly 0. A is not modified.
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
		// Reflect x = R[k:, k] onto alpha·e1. alpha takes the sign opposite to x[0]
		// so v = x - alpha·e1 never suffers cancellation.
		var norm float64
		for i := k; i < m; i++ {
			norm = math.Hypot(norm, r[i][k])
		}
		if norm == 0 {
			continue // column already zero below the diagonal
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

		// R ← H·R with H = I - 2·v·vᵀ/(vᵀv), touching rows k.. and columns k..
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
		// Q ← Q·H, touching columns k..
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
