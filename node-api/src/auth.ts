import type { NextFunction, Request, Response } from "express";
import jwt from "jsonwebtoken";

/** Rejects requests without a valid HS256 bearer token signed with JWT_SECRET. Other algorithms (incl. "none") are refused. */
export function requireAuth(req: Request, res: Response, next: NextFunction): void {
  const secret = process.env.JWT_SECRET;
  if (!secret) {
    res.status(500).json({ error: "server misconfigured: JWT_SECRET is not set" });
    return;
  }
  const [scheme, token] = (req.headers.authorization ?? "").split(" ");
  if (scheme !== "Bearer" || !token) {
    res.status(401).json({ error: "missing bearer token" });
    return;
  }
  try {
    jwt.verify(token, secret, { algorithms: ["HS256"] });
    next();
  } catch {
    res.status(401).json({ error: "invalid or expired token" });
  }
}
