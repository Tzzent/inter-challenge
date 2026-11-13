package qr

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

const tol = 1e-9

func matMul(a, b [][]float64) [][]float64 {
	out := make([][]float64, len(a))
	for i := range a {
		out[i] = make([]float64, len(b[0]))
		for j := range b[0] {
			for k := range b {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return out
}

func transpose(a [][]float64) [][]float64 {
	out := make([][]float64, len(a[0]))
	for j := range out {
		out[j] = make([]float64, len(a))
		for i := range a {
			out[j][i] = a[i][j]
		}
	}
	return out
}

func assertClose(t *testing.T, name string, got, want [][]float64) {
	t.Helper()
	if len(got) != len(want) || len(got[0]) != len(want[0]) {
		t.Fatalf("%s: shape %dx%d, want %dx%d", name, len(got), len(got[0]), len(want), len(want[0]))
	}
	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > tol {
				t.Fatalf("%s[%d][%d] = %v, want %v", name, i, j, got[i][j], want[i][j])
			}
		}
	}
}

func identity(n int) [][]float64 {
	out := make([][]float64, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1
	}
	return out
}

var shapes = map[string][][]float64{
	"square 2x2 (3-4-5)": {{3, 0}, {4, 5}},
	"square 3x3":         {{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}},
	"tall 4x2 (m>n)":     {{1, 2}, {3, 4}, {5, 6}, {7, 8}},
	"wide 2x3 (m<n)":     {{1, 2, 3}, {4, 5, 6}},
	"single row 1x3":     {{2, -1, 7}},
	"single col 3x1":     {{1}, {2}, {2}},
	"rank deficient 3x3": {{1, 2, 3}, {2, 4, 6}, {1, 1, 1}},
	"zero column 3x2":    {{0, 1}, {0, 2}, {0, 3}},
}

func TestDecomposeProperties(t *testing.T) {
	for name, a := range shapes {
		t.Run(name, func(t *testing.T) {
			m, n := len(a), len(a[0])
			q, r := Decompose(a)
			if len(q) != m || len(q[0]) != m {
				t.Fatalf("Q shape %dx%d, want %dx%d", len(q), len(q[0]), m, m)
			}
			if len(r) != m || len(r[0]) != n {
				t.Fatalf("R shape %dx%d, want %dx%d", len(r), len(r[0]), m, n)
			}
			assertClose(t, "QR", matMul(q, r), a)
			assertClose(t, "QtQ", matMul(transpose(q), q), identity(m))
			for i := range r {
				for j := 0; j < i && j < n; j++ {
					if r[i][j] != 0 {
						t.Fatalf("R[%d][%d] = %v, want exactly 0 below diagonal", i, j, r[i][j])
					}
				}
			}
		})
	}
}

func TestDecomposeKnownExample(t *testing.T) {
	q, r := Decompose([][]float64{{3, 0}, {4, 5}})
	assertClose(t, "Q", q, [][]float64{{-0.6, -0.8}, {-0.8, 0.6}})
	assertClose(t, "R", r, [][]float64{{-5, -4}, {0, 3}})
}

func TestDecomposeDoesNotMutateInput(t *testing.T) {
	a := [][]float64{{3, 0}, {4, 5}}
	Decompose(a)
	assertClose(t, "A", a, [][]float64{{3, 0}, {4, 5}})
}

// normalize invierte signos para que R tenga diagonal no negativa; (-Q_i)(-R_i) no altera QR.
func normalize(q, r [][]float64) {
	for i := 0; i < len(r) && i < len(r[0]); i++ {
		if r[i][i] < 0 {
			for j := range r[i] {
				r[i][j] = -r[i][j]
			}
			for k := range q {
				q[k][i] = -q[k][i]
			}
		}
	}
}

// TestDecomposeMatchesGonum usa gonum solo como oráculo. gonum exige m >= n y las últimas m-n
// columnas de una Q completa no son únicas, así que solo se comparan R y las primeras n columnas de Q.
func TestDecomposeMatchesGonum(t *testing.T) {
	for name, a := range shapes {
		m, n := len(a), len(a[0])
		if m < n || name == "rank deficient 3x3" || name == "zero column 3x2" {
			continue // oráculo indefinido (m<n) o no único (rango deficiente)
		}
		t.Run(name, func(t *testing.T) {
			dense := mat.NewDense(m, n, nil)
			for i := range a {
				dense.SetRow(i, a[i])
			}
			var f mat.QR
			f.Factorize(dense)
			var gq, gr mat.Dense
			f.QTo(&gq)
			f.RTo(&gr)
			wantQ, wantR := make([][]float64, m), make([][]float64, m)
			for i := 0; i < m; i++ {
				wantQ[i] = mat.Row(nil, i, &gq)
				wantR[i] = mat.Row(nil, i, &gr)
			}
			q, r := Decompose(a)
			normalize(q, r)
			normalize(wantQ, wantR)
			assertClose(t, "R", r, wantR)
			for i := 0; i < m; i++ {
				q[i], wantQ[i] = q[i][:n], wantQ[i][:n]
			}
			assertClose(t, "Q[:, :n]", q, wantQ)
		})
	}
}
