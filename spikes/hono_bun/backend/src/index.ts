import { Hono } from "hono";
import { cors } from "hono/cors";

const rawPort = Bun.env.PORT ?? process.env.PORT;
if (!rawPort) {
  throw new Error("PORT environment variable is required");
}

const port = Number.parseInt(rawPort, 10);
if (Number.isNaN(port)) {
  throw new Error(`Invalid PORT environment variable: ${rawPort}`);
}

const app = new Hono();

app.use("*", cors());

app.get("/ping", (c) => {
  return c.json({ status: "ok" });
});

export default {
  port,
  fetch: app.fetch,
};
