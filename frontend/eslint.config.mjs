import { defineConfig, globalIgnores } from 'eslint/config'
import nextVitals from 'eslint-config-next/core-web-vitals'
import testingLibrary from 'eslint-plugin-testing-library'
import jestDom from 'eslint-plugin-jest-dom'

const eslintConfig = defineConfig([
  ...nextVitals,
  // Testing Library + jest-dom recommended rules, scoped to *.test.{ts,tsx} files only (not
  // src/test/'s shared setup/helpers — e.g. setup.ts's global `cleanup()` call is expected to
  // trip `no-manual-cleanup`, a rule meant for individual test bodies where RTL's own
  // auto-cleanup already applies). At the scale of the frontend test suite these catch very
  // common RTL/jest-dom mistakes (missing `await` on `findBy*`, asserting on the wrong node,
  // redundant `act()` wraps, non-jest-dom-idiomatic assertions) that are otherwise easy to
  // introduce and easy to miss in review.
  {
    files: ['**/*.test.{ts,tsx}'],
    ...testingLibrary.configs['flat/react'],
  },
  {
    files: ['**/*.test.{ts,tsx}'],
    ...jestDom.configs['flat/recommended'],
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    '.next/**',
    'out/**',
    'build/**',
    'next-env.d.ts',
    'coverage/**',
  ]),
])
 
export default eslintConfig
