import { test } from "node:test";
import assert from "node:assert/strict";
import jwt from "jsonwebtoken";
import request from "supertest";

process.env.JWT_SECRET = "test-secret";
const { app } = await import("../src/app.js");

const token = jwt.sign({ sub: "tester" }, "test-secret", { algorithm: "HS256", expiresIn: "1h" });
const auth = { Authorization: `Bearer ${token}` };
const body = { matrices: [[[-0.6, -0.8], [-0.8, 0.6]], [[-5, -4], [0, 3]]] };

test("GET /health is public", async () => {
  await request(app).get("/health").expect(200, { status: "ok" });
});

test("POST /api/v1/stats without token -> 401", async () => {
  await request(app).post("/api/v1/stats").send(body).expect(401);
});

test("POST /api/v1/stats with bad token -> 401", async () => {
  const forged = jwt.sign({ sub: "x" }, "other-secret", { algorithm: "HS256" });
  await request(app).post("/api/v1/stats").set("Authorization", `Bearer ${forged}`).send(body).expect(401);
  await request(app).post("/api/v1/stats").set("Authorization", "Bearer not.a.jwt").send(body).expect(401);
});

test("POST /api/v1/stats with alg=none token -> 401", async () => {
  const none = jwt.sign({ sub: "x" }, "", { algorithm: "none" });
  await request(app).post("/api/v1/stats").set("Authorization", `Bearer ${none}`).send(body).expect(401);
});

test("POST /api/v1/stats with HS512 token -> 401", async () => {
  const hs512 = jwt.sign({ sub: "x" }, "test-secret", { algorithm: "HS512" });
  await request(app).post("/api/v1/stats").set("Authorization", `Bearer ${hs512}`).send(body).expect(401);
});

test("POST /api/v1/stats with valid token -> 200 and stats", async () => {
  const res = await request(app).post("/api/v1/stats").set(auth).send(body).expect(200);
  assert.equal(res.body.max, 3);
  assert.equal(res.body.min, -5);
  assert.ok(Math.abs(res.body.sum - -7.6) < 1e-9);
  assert.ok(Math.abs(res.body.average - -0.95) < 1e-9);
  assert.equal(res.body.anyDiagonal, false);
  assert.deepEqual(res.body.diagonal, [false, false]);
});

for (const [name, payload] of [
  ["missing matrices", {}],
  ["empty matrices", { matrices: [] }],
  ["empty matrix", { matrices: [[]] }],
  ["ragged rows", { matrices: [[[1, 2], [3]]] }],
  ["non-numeric value", { matrices: [[[1, "a"]]] }],
] as const) {
  test(`POST /api/v1/stats ${name} -> 400`, async () => {
    const res = await request(app).post("/api/v1/stats").set(auth).send(payload).expect(400);
    assert.equal(typeof res.body.error, "string");
  });
}

test("POST /api/v1/stats malformed JSON -> 400", async () => {
  const res = await request(app)
    .post("/api/v1/stats")
    .set(auth)
    .set("Content-Type", "application/json")
    .send('{"matrices": [[1,2]')
    .expect(400);
  assert.equal(typeof res.body.error, "string");
});
