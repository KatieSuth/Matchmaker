// Mock module for `next/navigation`'s App Router hooks (`useRouter`, `useParams`, `usePathname`,
// `useSearchParams`). Next's real router hooks require an `<AppRouterContext>` provider that
// doesn't exist in jsdom, so any component/hook under test that calls them needs the module
// mocked outright.
//
// Usage — `vi.mock` calls are hoisted above imports by Vitest, so this must be a top-level call
// in the test file (not inside a `describe`/`beforeEach`):
//
//   import { mockRouter, setMockParams, resetNextNavigationMock } from "@/test/nextNavigationMock";
//   vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));
//
//   beforeEach(() => resetNextNavigationMock());
//   // ...
//   expect(mockRouter.push).toHaveBeenCalledWith("/my_events");
import { vi } from "vitest";

export const mockRouter = {
  push: vi.fn(),
  replace: vi.fn(),
  back: vi.fn(),
  forward: vi.fn(),
  refresh: vi.fn(),
  prefetch: vi.fn(),
};

let mockParams: Record<string, string> = {};
let mockPathname = "/";
let mockSearchParams = new URLSearchParams();

/** Sets the value `useParams()` returns for the current test (e.g. `{ groupId: "group-1" }`). */
export function setMockParams(params: Record<string, string>): void {
  mockParams = params;
}

/** Sets the value `usePathname()` returns for the current test. */
export function setMockPathname(pathname: string): void {
  mockPathname = pathname;
}

/** Sets the value `useSearchParams()` returns for the current test. */
export function setMockSearchParams(params: URLSearchParams | Record<string, string>): void {
  mockSearchParams = params instanceof URLSearchParams ? params : new URLSearchParams(params);
}

/** Clears router call history and resets params/pathname/searchParams to their defaults. */
export function resetNextNavigationMock(): void {
  mockRouter.push.mockClear();
  mockRouter.replace.mockClear();
  mockRouter.back.mockClear();
  mockRouter.forward.mockClear();
  mockRouter.refresh.mockClear();
  mockRouter.prefetch.mockClear();
  mockParams = {};
  mockPathname = "/";
  mockSearchParams = new URLSearchParams();
}

export function useRouter() {
  return mockRouter;
}

export function useParams() {
  return mockParams;
}

export function usePathname() {
  return mockPathname;
}

export function useSearchParams() {
  return mockSearchParams;
}
