export type Matrix = number[][];

export interface Stats {
  max: number;
  min: number;
  average: number;
  sum: number;
  anyDiagonal: boolean;
  diagonal: boolean[];
}

/** Valores con |x| <= EPSILON cuentan como cero (la QR en coma flotante deja residuos de ~1e-16). */
export const EPSILON = 1e-9;

/**
 * Indica si todo elemento fuera de la diagonal principal (i !== j) es ~0.
 * Acepta matrices no cuadradas ("diagonal rectangular"): [[1,0,0],[0,2,0]] es diagonal.
 */
export function isDiagonal(m: Matrix, eps = EPSILON): boolean {
  return m.every((row, i) => row.every((x, j) => i === j || Math.abs(x) <= eps));
}

/** Devuelve un mensaje de error, o null si es una lista no vacía de matrices rectangulares y finitas. */
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

/** Calcula max, min, sum, average y las banderas de diagonal sobre todos los valores en una sola pasada. */
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
