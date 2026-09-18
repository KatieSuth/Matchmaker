// Shared RTL render wrapper used across component/form/page tests.
import { ReactElement, ReactNode } from "react";
import { render, RenderOptions, RenderResult } from "@testing-library/react";
import { AuthContext, AuthContextType } from "@/app/_context/AuthContext";
import type { User } from "@/app/_types/types";

export interface RenderWithProvidersOptions extends Omit<RenderOptions, "wrapper"> {
  user?: User | null;
  isAuthenticated?: boolean;
  isLoading?: boolean;
  setUser?: AuthContextType["setUser"];
  setIsAuthenticated?: AuthContextType["setIsAuthenticated"];
  logout?: AuthContextType["logout"];
}

/**
 * Renders `ui` inside a stub `AuthContext.Provider` with a controlled auth state, instead of the
 * real `AuthProvider` (whose bootstrap-on-mount effect performs a real session-refresh network
 * call on every mount). Defaults to a logged-out, finished-loading viewer; pass `user` and
 * `isAuthenticated: true` to simulate a signed-in viewer. `setUser`/`setIsAuthenticated`/`logout`
 * default to no-op stubs — pass `vi.fn()` spies when a test needs to assert on them being called.
 */
export function renderWithProviders(
  ui: ReactElement,
  {
    user = null,
    isAuthenticated = false,
    isLoading = false,
    setUser = () => {},
    setIsAuthenticated = () => {},
    logout = async () => {},
    ...renderOptions
  }: RenderWithProvidersOptions = {},
): RenderResult {
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <AuthContext.Provider value={{ user, isAuthenticated, isLoading, setUser, setIsAuthenticated, logout }}>
        {children}
      </AuthContext.Provider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}

// Re-export RTL's `screen`/`within`/etc. surface so most test files only need one import line:
//   import { renderWithProviders, screen, userEvent } from "@/test/render";
export * from "@testing-library/react";
export { default as userEvent } from "@testing-library/user-event";
