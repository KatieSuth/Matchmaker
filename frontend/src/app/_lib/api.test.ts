// Tests for the minimal unauthenticated fetch helper (no cookies/auth headers — see the module
// comment on when to use this over the shared `axios` client).
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { api } from "./api";

describe("api.get", () => {
  it("returns parsed JSON data with no error on success", async () => {
    server.use(http.get(`${TEST_API_URL}/widgets`, () => HttpResponse.json({ id: "w1" })));

    const result = await api.get<{ id: string }>("/widgets");

    expect(result).toEqual({ data: { id: "w1" }, error: null });
  });

  it("returns the response body text as the error on a non-2xx response", async () => {
    server.use(http.get(`${TEST_API_URL}/widgets`, () => new HttpResponse("widget not found", { status: 404 })));

    const result = await api.get<unknown>("/widgets");

    expect(result).toEqual({ data: null, error: "widget not found" });
  });

  it("falls back to statusText when the error response body is empty", async () => {
    server.use(http.get(`${TEST_API_URL}/widgets`, () => new HttpResponse(null, { status: 500 })));

    const result = await api.get<unknown>("/widgets");

    expect(result.data).toBeNull();
    expect(result.error).toBeTruthy();
  });

  it("returns the exception message when the request throws (e.g. network failure)", async () => {
    server.use(http.get(`${TEST_API_URL}/widgets`, () => HttpResponse.error()));

    const result = await api.get<unknown>("/widgets");

    expect(result.data).toBeNull();
    expect(result.error).toBeTruthy();
  });
});

describe("api.post / put / delete", () => {
  it("sends a JSON-stringified body with POST", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(`${TEST_API_URL}/widgets`, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ ok: true });
      }),
    );

    await api.post("/widgets", { name: "Gadget" });

    expect(receivedBody).toEqual({ name: "Gadget" });
  });

  it("sends a JSON-stringified body with PUT", async () => {
    let receivedBody: unknown;
    server.use(
      http.put(`${TEST_API_URL}/widgets/w1`, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ ok: true });
      }),
    );

    await api.put("/widgets/w1", { name: "Updated" });

    expect(receivedBody).toEqual({ name: "Updated" });
  });

  it("sends no body with DELETE", async () => {
    server.use(http.delete(`${TEST_API_URL}/widgets/w1`, () => HttpResponse.json({ ok: true })));

    const result = await api.delete<{ ok: boolean }>("/widgets/w1");

    expect(result).toEqual({ data: { ok: true }, error: null });
  });
});
