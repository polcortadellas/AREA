import { describe, expect, it } from "bun:test";
import server from "./index";

describe("GET /ping", () => {
  it("returns exact body {\"status\":\"ok\"}", async () => {
    const res = await server.fetch(new Request("http://localhost/ping"));
    expect(res.status).toBe(200);
    const text = await res.text();
    expect(text).toBe('{"status":"ok"}');
  });
});
