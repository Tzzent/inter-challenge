import { test } from "node:test";
import assert from "node:assert/strict";
import { computeStats, isDiagonal, validateMatrices } from "../src/stats.js";

test("isDiagonal: square diagonal matrix", () => {
  assert.equal(isDiagonal([[2, 0], [0, 7]]), true);
});

test("isDiagonal: off-diagonal value breaks it", () => {
  assert.equal(isDiagonal([[2, 1], [0, 7]]), false);
});

test("isDiagonal: tolerates floating-point noise within epsilon", () => {
  assert.equal(isDiagonal([[1, 1e-16], [-1e-12, 1]]), true);
  assert.equal(isDiagonal([[1, 1e-6], [0, 1]]), false);
});

test("isDiagonal: rectangular diagonal (only i===j may be non-zero)", () => {
  assert.equal(isDiagonal([[1, 0, 0], [0, 2, 0]]), true);
  assert.equal(isDiagonal([[1, 0], [0, 2], [0, 0]]), true);
  assert.equal(isDiagonal([[1, 0], [0, 2], [3, 0]]), false);
});

test("computeStats: QR of [[3,0],[4,5]] (Householder signs)", () => {
  const q = [[-0.6, -0.8], [-0.8, 0.6]];
  const r = [[-5, -4], [0, 3]];
  const s = computeStats([q, r]);
  assert.equal(s.max, 3);
  assert.equal(s.min, -5);
  assert.ok(Math.abs(s.sum - -7.6) < 1e-9);
  assert.ok(Math.abs(s.average - -0.95) < 1e-9);
  assert.equal(s.anyDiagonal, false);
  assert.deepEqual(s.diagonal, [false, false]);
});

test("computeStats: anyDiagonal when one matrix is diagonal", () => {
  const s = computeStats([[[1, 2], [3, 4]], [[5, 0], [0, 6]]]);
  assert.equal(s.anyDiagonal, true);
  assert.deepEqual(s.diagonal, [false, true]);
  assert.equal(s.sum, 21);
  assert.equal(s.average, 21 / 8);
});

test("validateMatrices rejects bad input", () => {
  assert.match(validateMatrices(undefined) ?? "", /matrices/);
  assert.match(validateMatrices([]) ?? "", /empty/);
  assert.match(validateMatrices([[]]) ?? "", /empty/);
  assert.match(validateMatrices([[[]]]) ?? "", /empty/);
  assert.match(validateMatrices([[[1, 2], [3]]]) ?? "", /same length/);
  assert.match(validateMatrices([[[1, "2"]]]) ?? "", /finite/);
  assert.match(validateMatrices([[[1, null]]]) ?? "", /finite/);
  assert.equal(validateMatrices([[[1, 2], [3, 4]]]), null);
});
