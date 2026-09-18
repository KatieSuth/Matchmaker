// vitest-axe@0.1.0 ships an ambient type augmentation (`vitest-axe/extend-expect`) written against
// an older, single-generic `Vi.Assertion<T = any>` shape. The installed Vitest 5's `Assertion<R, T>`
// takes two required generics with no defaults, so that augmentation silently fails to merge and
// `toHaveNoViolations` doesn't type-check. Declare it ourselves against the current shape instead
// of depending on the (also currently broken — see setup.ts) upstream declaration.
declare module "vitest" {
  interface Assertion<R, T> {
    toHaveNoViolations(): R;
  }
}

export {};
