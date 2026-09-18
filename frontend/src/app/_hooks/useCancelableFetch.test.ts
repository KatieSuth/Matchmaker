// Tests for the shared cancel-safe fetch effect, exercised against the real `fetchGames` service
// function and the MSW-mocked network layer (rather than a hand-rolled fake fetcher) so this also
// doubles as an end-to-end check that the MSW test infra (src/test/msw) is wired up correctly.
import { http, HttpResponse } from "msw";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { buildGame } from "@/test/fixtures";
import { fetchGames } from "@/app/_services/games";
import { useCancelableFetch } from "./useCancelableFetch";

describe("useCancelableFetch", () => {
  it("calls onStart synchronously, then onSuccess with the resolved data", async () => {
    const game = buildGame({ name: "Overwatch" });
    server.use(http.get(`${TEST_API_URL}/games`, () => HttpResponse.json([game])));

    const onStart = vi.fn();
    const onSuccess = vi.fn();
    const onError = vi.fn();

    renderHook(() =>
      useCancelableFetch({
        fetcher: (signal) => fetchGames(signal),
        onStart,
        onSuccess,
        onError,
        deps: [],
      }),
    );

    expect(onStart).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(onSuccess).toHaveBeenCalledWith([game]));
    expect(onError).not.toHaveBeenCalled();
  });

  it("calls onError (not onSuccess) when the request fails", async () => {
    server.use(http.get(`${TEST_API_URL}/games`, () => HttpResponse.json({ message: "boom" }, { status: 500 })));

    const onSuccess = vi.fn();
    const onError = vi.fn();

    renderHook(() =>
      useCancelableFetch({
        fetcher: (signal) => fetchGames(signal),
        onSuccess,
        onError,
        deps: [],
      }),
    );

    await waitFor(() => expect(onError).toHaveBeenCalledTimes(1));
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("skips the fetch entirely when enabled is false", async () => {
    server.use(http.get(`${TEST_API_URL}/games`, () => HttpResponse.json([buildGame()])));
    const onStart = vi.fn();

    renderHook(() =>
      useCancelableFetch({
        fetcher: (signal) => fetchGames(signal),
        enabled: false,
        onStart,
        onSuccess: vi.fn(),
        onError: vi.fn(),
        deps: [],
      }),
    );

    // Give any (incorrectly-firing) async work a tick to resolve, then assert nothing happened.
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(onStart).not.toHaveBeenCalled();
  });

  it("aborts the in-flight request on unmount and does not call onSuccess", async () => {
    // A fixed server-side delay (rather than a manually-captured resolver) keeps this
    // deterministic: the response can't resolve before `unmount()` runs synchronously below.
    server.use(
      http.get(`${TEST_API_URL}/games`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 50));
        return HttpResponse.json([buildGame()]);
      }),
    );

    const onSuccess = vi.fn();
    const { unmount } = renderHook(() =>
      useCancelableFetch({
        fetcher: (signal) => fetchGames(signal),
        onSuccess,
        onError: vi.fn(),
        deps: [],
      }),
    );

    unmount();
    // Wait past the handler's artificial delay so we know the response (if not aborted) would
    // have arrived by now.
    await new Promise((resolve) => setTimeout(resolve, 150));

    expect(onSuccess).not.toHaveBeenCalled();
  });
});
