export type Matrix = number[][];

export interface Stats {
  max: number;
  min: number;
  average: number;
  sum: number;
  anyDiagonal: boolean;
  diagonal: boolean[];
}

/** Values with |x| <= EPSILON count as zero (floating-point QR leaves ~1e-16 residues). */
export const EPSILON = 1e-9;

/**
 * Reports whether every element off the main diagonal (i !== j) is ~0.
 * Non-square matrices are accepted ("rectangular diagonal"): [[1,0,0],[0,2,0]] is diagonal.
 */
export function isDiagonal(m: Matrix, eps = EPSILON): boolean {
  return m.every((row, i) => row.every((x, j) => i === j || Math.abs(x) <= eps));
}

/** Returns an error message for invalid input, or null when it is a non-empty list of rectangular finite matrices. */
export function validateMatrices(input: unknown): string | null {
  if (!Array.isArray(input)) return "matrices must be an array of matrices";
  if (input.length === 0) return "matrices must not be empty";
  for (const [k, m] of input.entries()) {
    if (!Array.isArray(m) || m.length === 0) return `matrix ${k} must be a non-empty array of rows`;
    const cols = Array.isArray(m[0]) ? m[0].length : 0;
    if (cols === 0) return `matrix ${k} has an empty first row`;
    for (const row of m) {
      if (!Array.isArray(row) || row.length !== cols) return `matrix ${k}: all rows must have the same length`;
      if (!row.every((x) => typeof x === "number" && Number.isFinite(x))) return `matrix ${k}: values must be finite numbers`;
    }
  }
  return null;
}

/** Computes max, min, sum, average and the diagonal flags over every value of every matrix in a single pass. */
export function computeStats(matrices: Matrix[]): Stats {
  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;
  const diagonal = matrices.map((m) => {
    let diag = true;
    m.forEach((row, i) =>
      row.forEach((x, j) => {
        if (x > max) max = x;
        if (x < min) min = x;
        sum += x;
        count++;
        if (i !== j && Math.abs(x) > EPSILON) diag = false;
      }),
    );
    return diag;
  });
  return { max, min, average: sum / count, sum, anyDiagonal: diagonal.some(Boolean), diagonal };
}
