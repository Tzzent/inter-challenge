// Prueba end-to-end contra el stack en ejecución (Go -> Node). Sale con código != 0 si falla.
import assert from "node:assert/strict";

const base = process.env.GO_API_URL ?? "http://localhost:3000";
const close = (a, b, msg) => assert.ok(Math.abs(a - b) < 1e-9, `${msg}: got ${a}, want ${b}`);
const closeMatrix = (got, want, name) =>
  want.forEach((row, i) => row.forEach((x, j) => close(got[i][j], x, `${name}[${i}][${j}]`)));
const post = (path, body, token) =>
  fetch(base + path, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...(token && { Authorization: `Bearer ${token}` }) },
    body: JSON.stringify(body),
  });

try {
  const health = await (await fetch(`${base}/health`)).json();
  assert.equal(health.statsApi, "ok", `health: ${JSON.stringify(health)}`);

  const tokenRes = await post("/auth/token", {
    username: process.env.AUTH_USERNAME,
    password: process.env.AUTH_PASSWORD,
  });
  assert.equal(tokenRes.status, 200, "token request");
  const { token } = await tokenRes.json();

  const unauth = await post("/api/v1/qr", { matrix: [[3, 0], [4, 5]] });
  assert.equal(unauth.status, 401, "call without token must be 401");

  const res = await post("/api/v1/qr", { matrix: [[3, 0], [4, 5]] }, token);
  assert.equal(res.status, 200, `qr status: ${await res.clone().text()}`);
  const { q, r, stats } = await res.json();
  closeMatrix(q, [[-0.6, -0.8], [-0.8, 0.6]], "Q");
  closeMatrix(r, [[-5, -4], [0, 3]], "R");
  close(stats.sum, -7.6, "sum");
  close(stats.average, -0.95, "average");
  close(stats.max, 3, "max");
  close(stats.min, -5, "min");
  assert.equal(stats.anyDiagonal, false, "anyDiagonal");

  console.log("integration test passed");
} catch (err) {
  console.error("integration test FAILED:", err.message);
  process.exit(1);
}
