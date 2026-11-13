import express, { type NextFunction, type Request, type Response } from "express";
import { requireAuth } from "./auth.js";
import { computeStats, validateMatrices, type Matrix } from "./stats.js";

/** Express app without listen(), so tests can import it. */
export const app = express();

app.use(express.json({ limit: "1mb" }));

app.get("/health", (_req, res) => {
  res.json({ status: "ok" });
});

app.use("/api/v1", requireAuth);

app.post("/api/v1/stats", (req, res) => {
  const matrices: unknown = req.body?.matrices;
  const error = validateMatrices(matrices);
  if (error) {
    res.status(400).json({ error });
    return;
  }
  res.json(computeStats(matrices as Matrix[]));
});

// Body-parser errors (malformed JSON, too large) carry a status; anything else is a 500.
app.use((err: { status?: number; message?: string }, _req: Request, res: Response, _next: NextFunction) => {
  const status = err.status && err.status >= 400 && err.status < 500 ? err.status : 500;
  res.status(status).json({ error: status === 500 ? "internal error" : (err.message ?? "bad request") });
});
